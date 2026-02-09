package repository

import "gorm.io/gorm"

// Repositories aggregates all repository instances for easy dependency injection
type Repositories struct {
	User       *UserRepository
	Membership *MembershipRepository
	Session    *SessionRepository
}

// NewRepositories creates and initializes all repositories
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:       NewUserRepository(db),
		Membership: NewMembershipRepository(db),
		Session:    NewSessionRepository(db),
	}
}
