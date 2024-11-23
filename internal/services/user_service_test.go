package services_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gochat/internal/services"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Set up shared environment variables or configurations
	os.Setenv("DB_PATH", "../../database.db")
	os.Exit(m.Run())
}

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	name := "Billy"
	externalID := uuid.New().String()
	newUserParams := services.UserParams{
		Name:       &name,
		Email:      externalID + "@test.com",
		ExternalID: &externalID,
	}

	// New service
	userService := services.NewUserService()
	assert.NotNil(t, userService)

	// Create User
	createdUser, err := userService.Create(ctx, newUserParams)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)

	// Assert created user details
	assert.Equal(t, *(newUserParams.ExternalID), createdUser.Externalid.String)
	assert.Equal(t, "A1234", createdUser.Account)
	assert.Equal(t, *(newUserParams.Name), createdUser.Name.String)
	assert.Equal(t, newUserParams.Email, createdUser.Email)
}

func TestUserService_CreateUser_Account(t *testing.T) {
	ctx := context.Background()
	name := "Billy"
	externalID := uuid.New().String()
	newUserParams := services.UserParams{
		Name:       &name,
		Email:      externalID + "@google.com",
		ExternalID: &externalID,
	}

	// New service
	userService := services.NewUserService()
	assert.NotNil(t, userService)

	// Create User
	createdUser, err := userService.Create(ctx, newUserParams)
	assert.Error(t, err)
	assert.Nil(t, createdUser)

}

func TestUserService_GetUser(t *testing.T) {
	ctx := context.Background()
	name := "Billy"
	externalID := uuid.New().String()
	newUserParams := services.UserParams{
		Name:       &name,
		Email:      externalID + "@test.com",
		ExternalID: &externalID,
	}

	// New service
	userService := services.NewUserService()

	// Create a user first
	createdUser, err := userService.Create(ctx, newUserParams)
	assert.NoError(t, err)

	// Retrieve the user by ID
	retrievedUser, err := userService.Get(ctx, createdUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedUser)

	// Assert retrieved user details
	assert.Equal(t, retrievedUser.ID, createdUser.ID)

	// Retrieve the user by ID
	nonExistentUser, err := userService.Get(ctx, "idontexist")
	assert.NoError(t, err)
	assert.Nil(t, nonExistentUser)
}

func TestUserService_GetOrCreate_ExistingUser(t *testing.T) {
	ctx := context.Background()
	name := "Billy"
	externalID := uuid.New().String()
	newUserParams := services.UserParams{
		Name:       &name,
		Email:      externalID + "@test.com",
		ExternalID: &externalID,
	}

	// New service
	userService := services.NewUserService()

	// Create a user first
	createdUser, err := userService.Create(ctx, newUserParams)
	assert.NoError(t, err)

	// Call GetOrCreate for the same user
	alreadyCreated, err := userService.GetOrCreate(ctx, newUserParams)
	assert.NoError(t, err)

	// Assert it returned the existing user
	assert.Equal(t, alreadyCreated.ID, createdUser.ID)
}

func TestUserService_GetOrCreate_NewUser(t *testing.T) {
	ctx := context.Background()
	externalID := uuid.New().String()

	// Call GetOrCreate with a new user
	userService := services.NewUserService()
	newlyCreated, err := userService.GetOrCreate(ctx, services.UserParams{
		Name:       nil,
		Email:      uuid.New().String() + "@test.com",
		ExternalID: &externalID,
	})
	assert.NoError(t, err)

	// Assert new user details
	assert.NotNil(t, newlyCreated)
	fmt.Println("newly created: ", newlyCreated.ID)
	assert.Equal(t, externalID, newlyCreated.Externalid.String)
}
