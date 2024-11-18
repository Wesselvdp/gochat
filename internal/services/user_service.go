package services

import (
	"context"
	"database/sql"
	"fmt"
	database "gochat/internal/db"
	"gochat/internal/schema"
	"gochat/pkg/utils"
)

type User struct {
	Name        string
	Email       string
	ExternalID  int64
	ID          int64
	account     string
	CreatedAt   string
	UpdatededAt string
}

type UserSearchParams struct {
}

func getUser(id int64) (*schema.User, error) {
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

func GetOrCreateUser(email string) (*User, error) {
	var user *User
	user, err := GetUserByEmail(email)

	if err != nil {
		return nil, err
	}

	if user != nil {
		fmt.Printf("User %v already exists\n", user.Email)
		return user, nil
	}

	user, err = CreateUser(User{Name: "James", Email: email})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByEmail(email string) (*User, error) {
	ctx := context.Background()
	queries, _, err := database.Init()
	if err != nil {
		return nil, err
	}

	user, err := queries.GetUserByEmail(ctx, sql.NullString{String: email, Valid: true})

	//if err != nil {
	//	return nil, err
	//}

	if user.Email.String == "" {
		return nil, nil
	}
	readableUser := User{
		ID:          user.ID,
		Name:        user.Name.String,
		Email:       user.Email.String,
		CreatedAt:   user.Createdat.String,
		UpdatededAt: user.Updatedat.String,
	}

	fmt.Println(readableUser)

	//if err != nil {
	//	return nil, err
	//}
	return &readableUser, nil
}

func CreateUser(user User) (*User, error) {
	fmt.Println("Creating new user")
	now := utils.GetTime()
	ctx := context.Background()
	queries, _, err := database.Init()

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	newUser, err := queries.CreateUser(ctx, schema.CreateUserParams{
		Name:      sql.NullString{String: user.Name, Valid: true},
		Email:     sql.NullString{String: user.Email, Valid: true},
		Updatedat: now,
		Createdat: now,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user = User{
		Name:  newUser.Name.String,
		Email: newUser.Email.String,
		ID:    newUser.ID,
	}
	return &user, nil
}
