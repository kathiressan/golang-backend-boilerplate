// Package auth provides authentication business logic and services
package auth

import (
	"context"
	"errors"
	"fmt"
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/repository"
	"ovmsa-be/pkg/jwt"
	"ovmsa-be/pkg/password"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrInvalidCredentials is returned when email or password is incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserNotFound is returned when user doesn't exist
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidRefreshToken is returned when refresh token is invalid or expired
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	// ErrSessionNotFound is returned when session doesn't exist
	ErrSessionNotFound = errors.New("session not found")
	// ErrNoOrganizationAccess is returned when user has no access to any organization
	ErrNoOrganizationAccess = errors.New("user has no organization access")
)

// LoginInput represents login credentials
type LoginInput struct {
	Email    string
	Password string
	OrgID    string // Optional: for multi-org users
	Audience string // Optional: "web" or "mobile"
}

// LoginResult contains tokens and user info
type LoginResult struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	TokenType    string   `json:"token_type"`
	User         UserInfo `json:"user"`
}

// UserInfo contains safe user information
type UserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	IsRoot  bool   `json:"is_root"`
	OrgID   string `json:"org_id,omitempty"`
	OrgPath string `json:"org_path,omitempty"`
	Role    string `json:"role,omitempty"`
}

// AuthService handles authentication business logic
type AuthService struct {
	db *gorm.DB
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

// Login authenticates user and creates session
func (s *AuthService) Login(ctx context.Context, input LoginInput, ipAddress, userAgent string) (*LoginResult, error) {
	// 1. Find user by email
	user, err := repository.Repo.User.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// 2. Verify password
	if err := password.VerifyPassword(input.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 3. Determine organization context
	var orgID, orgPath, role string
	var isRoot bool = user.IsRoot

	if user.IsRoot {
		// Root users can login without organization context
		orgID = ""
		orgPath = "/"
		role = "root"
	} else {
		// Non-root users need organization membership
		if input.OrgID != "" {
			// User specified an organization
			membership, err := repository.Repo.Membership.FindByUserAndOrg(ctx, user.ID, input.OrgID)
			if err != nil || membership == nil {
				return nil, fmt.Errorf("user does not have access to organization %s", input.OrgID)
			}
			orgID = membership.OrgID
			orgPath = membership.OrgPath
			role = membership.Role
		} else {
			// Get user's first organization (default)
			memberships, err := repository.Repo.Membership.FindAllByUser(ctx, user.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to find user memberships: %w", err)
			}
			if len(memberships) == 0 {
				return nil, ErrNoOrganizationAccess
			}
			// Use the first membership as default
			orgID = memberships[0].OrgID
			orgPath = memberships[0].OrgPath
			role = memberships[0].Role
		}
	}

	// 4. Generate refresh token
	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 5. Create session with hashed refresh token
	hashedRefreshToken := jwt.HashToken(refreshToken)
	session := &entities.Session{
		UserID:       user.ID,
		RefreshToken: hashedRefreshToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(time.Duration(30*24) * time.Hour), // 30 days from config
	}

	if err := repository.Repo.Session.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 6. Generate access token
	identity := jwt.UserIdentity{
		UserID:    user.ID,
		SessionID: session.ID,
		OrgID:     orgID,
		OrgPath:   orgPath,
		Role:      role,
		IsRoot:    isRoot,
		Audience:  input.Audience,
	}

	accessToken, err := jwt.GenerateAccessToken(identity)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 7. Return login result
	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Return unhashed token to client
		ExpiresIn:    60 * 60,      // 60 minutes in seconds (from config)
		TokenType:    "Bearer",
		User: UserInfo{
			ID:      user.ID,
			Email:   user.Email,
			Name:    user.Name,
			IsRoot:  isRoot,
			OrgID:   orgID,
			OrgPath: orgPath,
			Role:    role,
		},
	}, nil
}

// RefreshToken generates new access token from refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, audience string) (*LoginResult, error) {
	// 1. Hash the refresh token to look it up
	hashedToken := jwt.HashToken(refreshToken)

	// 2. Find session by hashed refresh token
	session, err := repository.Repo.Session.FindByRefreshToken(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	if session == nil {
		return nil, ErrInvalidRefreshToken
	}

	// 3. Check if session is expired
	if session.IsExpired() {
		// Delete expired session
		_, _ = repository.Repo.Session.Delete(ctx, session.ID)
		return nil, ErrInvalidRefreshToken
	}

	// 4. Get user information
	user, err := repository.Repo.User.FindByID(ctx, session.UserID, nil)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	// 5. Determine organization context (same logic as login)
	var orgID, orgPath, role string
	var isRoot bool = user.IsRoot

	if user.IsRoot {
		orgID = ""
		orgPath = "/"
		role = "root"
	} else {
		// Get user's first organization
		memberships, err := repository.Repo.Membership.FindAllByUser(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to find user memberships: %w", err)
		}
		if len(memberships) == 0 {
			return nil, ErrNoOrganizationAccess
		}
		orgID = memberships[0].OrgID
		orgPath = memberships[0].OrgPath
		role = memberships[0].Role
	}

	// 6. Generate new refresh token (token rotation)
	newRefreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 7. Update session with new hashed refresh token
	newHashedToken := jwt.HashToken(newRefreshToken)
	newExpiry := time.Now().Add(time.Duration(30*24) * time.Hour) // 30 days

	if err := repository.Repo.Session.RotateRefreshToken(ctx, session.ID, hashedToken, newHashedToken, newExpiry); err != nil {
		return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	// 8. Generate new access token
	identity := jwt.UserIdentity{
		UserID:    user.ID,
		SessionID: session.ID,
		OrgID:     orgID,
		OrgPath:   orgPath,
		Role:      role,
		IsRoot:    isRoot,
		Audience:  audience,
	}

	accessToken, err := jwt.GenerateAccessToken(identity)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 9. Return new tokens
	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken, // Return unhashed token to client
		ExpiresIn:    60 * 60,         // 60 minutes in seconds
		TokenType:    "Bearer",
		User: UserInfo{
			ID:      user.ID,
			Email:   user.Email,
			Name:    user.Name,
			IsRoot:  isRoot,
			OrgID:   orgID,
			OrgPath: orgPath,
			Role:    role,
		},
	}, nil
}

// Logout revokes a specific session
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrSessionNotFound
	}

	// Hard delete the session using Unscoped (permanently remove it)
	db := repository.Repo.Session.GetDB(ctx).Unscoped()
	result := db.Where("id = ?", sessionID).Delete(&entities.Session{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// LogoutAll revokes all sessions for a user
func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrUserNotFound
	}

	// Delete all sessions for the user
	if err := repository.Repo.Session.DeleteAllByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete all sessions: %w", err)
	}

	return nil
}

// GetCurrentUser retrieves user info from identity
func (s *AuthService) GetCurrentUser(ctx context.Context, identity *entities.Identity) (*UserInfo, error) {
	if identity == nil {
		return nil, ErrUserNotFound
	}

	// Get full user information
	user, err := repository.Repo.User.FindByID(ctx, identity.UserID, nil)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	return &UserInfo{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		IsRoot:  identity.IsRoot,
		OrgID:   identity.OrgID,
		OrgPath: identity.OrgPath,
		Role:    identity.Role,
	}, nil
}
