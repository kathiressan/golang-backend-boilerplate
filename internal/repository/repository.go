package repository

import "gorm.io/gorm"

var Repo *Repositories

// Repositories aggregates all repository instances for easy dependency injection
type Repositories struct {
	User         *UserRepository
	Membership   *MembershipRepository
	Session      *SessionRepository
	Organization *OrganizationRepository
	OrgGrant     *OrgGrantRepository
	SigningKey   *SigningKeyRepository
}

// NewRepositories creates and initializes all repositories
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:         NewUserRepository(db),
		Membership:   NewMembershipRepository(db),
		Session:      NewSessionRepository(db),
		Organization: NewOrganizationRepository(db),
		OrgGrant:     NewOrgGrantRepository(db),
		SigningKey:   NewSigningKeyRepository(db),
	}
}

func Initialize(db *gorm.DB) {
	Repo = NewRepositories(db)
}
