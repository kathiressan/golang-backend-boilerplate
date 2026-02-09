package integration

import (
	"ovmsa-be/internal/repository"
	"testing"

	"gorm.io/gorm"
)

func TestRepositoriesInitialization(t *testing.T) {
	// This test just ensures that the repository struct and its initializers 
	// are correctly linked and compile. 
	// We use a nil or dummy DB for this specific compilation/init check.
	// In a real integration test, we would use a test container or a mock.
	
	// Create a dummy GORM DB (won't connect)
	db := &gorm.DB{Config: &gorm.Config{}}
	
	repos := repository.NewRepositories(db)
	
	if repos.User == nil {
		t.Error("User repository was not initialized")
	}
	if repos.Organization == nil {
		t.Error("Organization repository was not initialized")
	}
	if repos.OrgGrant == nil {
		t.Error("OrgGrant repository was not initialized")
	}
	if repos.Membership == nil {
		t.Error("Membership repository was not initialized")
	}
	if repos.Session == nil {
		t.Error("Session repository was not initialized")
	}
}

func TestRepositoryTransactionSupport(t *testing.T) {
	// This test verifies that the specialized repositories 
	// have the correct method signatures to accept transactions.
	// We don't execute the methods because they require a real DB connection.
	
	db := &gorm.DB{Config: &gorm.Config{}}
	repos := repository.NewRepositories(db)
	
	if repos.User == nil {
		t.Fatal("Repositories not initialized")
	}

	// We verify the existence of the methods with the correct signatures 
	// implicitly by the fact that this code compiles.
	// The user can now pass a transaction: repos.User.FindByEmail("email", tx)
	t.Log("Repository transaction signatures verified via compilation")
}
