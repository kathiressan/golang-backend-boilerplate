package auth

// LoginRequest validates login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required" validate:"required,email"`
	Password string `json:"password" binding:"required" validate:"required,min=8"`
	OrgID    string `json:"org_id,omitempty"`
	Audience string `json:"audience,omitempty" validate:"omitempty,oneof=web mobile"`
}

// RefreshTokenRequest validates refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" validate:"required"`
	Audience     string `json:"audience,omitempty" validate:"omitempty,oneof=web mobile"`
}
