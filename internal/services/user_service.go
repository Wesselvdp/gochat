package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	database "gochat/internal/db"
	"gochat/internal/schema"
)

type UserCreate struct {
	Name       *string
	Email      string
	ExternalID string
	Account    *string
}

type UserSearchParams struct {
}

func getUser(id string) (*schema.User, error) {
	ctx := context.Background()
	queries, _, err := database.Init()
	if err != nil {
		return nil, err
	}
	user, err := queries.GetUser(ctx, id)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetOrCreateUser(userData UserCreate) (*schema.User, error) {
	var user *schema.User
	fmt.Println("coming in:", userData.Email)
	user, err := GetUserByEmail(userData.Email)

	if err != nil {
		return nil, err
	}

	if user != nil {
		fmt.Printf("User %v already exists\n", user.Email)
		return user, nil
	}

	user, err = CreateUser(userData)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByEmail(email string) (*schema.User, error) {

	ctx := context.Background()
	queries, _, err := database.Init()
	if err != nil {
		return nil, err
	}
	fmt.Println("Searching user by email", email)
	user, err := queries.GetUserByEmail(ctx, email)
	//

	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the "no rows" case explicitly
			fmt.Println("No user found with the given email:", email)
			return nil, nil // or return a custom error
		}
		// Handle other errors
		fmt.Println("Error getting user by email:", err.Error())
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}

	return &user, nil
}

func CreateUser(create UserCreate) (*schema.User, error) {
	fmt.Println("Creating new user", create)
	ctx := context.Background()
	queries, _, err := database.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	user := schema.CreateUserParams{
		ID:    uuid.New().String(),
		Email: create.Email,
		Name:  "", // default empty string
	}

	if create.Name != nil {
		user.Name = *create.Name
	}
	if create.Account != nil {
		user.Account = *create.Account
	}

	newUser, err := queries.CreateUser(ctx, user)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &newUser, nil
}
