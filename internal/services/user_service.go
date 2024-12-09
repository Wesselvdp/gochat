package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	database "gochat/internal/db"
	"gochat/internal/schema"
	"gochat/pkg/utils"
	"strings"
)

type UserParams struct {
	Name       *string
	Email      string
	ExternalID *string
}

type UserSearchParams struct {
}

type UserService struct {
	queries *schema.Queries
}

func getDomain(email string) (string, error) {
	// Split the email address at the '@' symbol
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email address: %s", email)
	}
	return parts[1], nil
}

func NewUserService() *UserService {
	queries, _, err := database.Init()
	if err != nil {
		fmt.Println("Error initializing queries for user service: " + err.Error())
		return nil
	}
	return &UserService{queries: queries}
}

func (us *UserService) getAccountFromEmail(ctx context.Context, email string) (*string, error) {
	domain, err := getDomain(email)

	if err != nil {
		return nil, err
	}
	accountID, err := us.queries.GetAccountByDomain(ctx, domain)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &accountID, nil
}

func (us *UserService) Create(ctx context.Context, params UserParams) (*schema.User, error) {
	accountID, err := us.getAccountFromEmail(ctx, params.Email)
	if err != nil {
		fmt.Println("error getting account from email: " + err.Error())
		return nil, fmt.Errorf("error getting account by domain: %s", err.Error())
	}

	if accountID == nil {
		eventService := NewEventService("")
		eventService.Create(UnknownAccount, map[string]interface{}{
			"email":  params.Email,
			"status": "error",
		})
		return nil, fmt.Errorf("Hi there, it seems your organisation has no active subscription. If you like access, please contact wessel@torgon.io")
	}

	user := schema.CreateUserParams{
		ID:      uuid.New().String(),
		Email:   params.Email,
		Account: *accountID,
	}

	if params.Name != nil {
		user.Name = utils.StringToNullString(*params.Name)
	}
	if params.ExternalID != nil {
		user.Externalid = utils.StringToNullString(*params.ExternalID)
	}

	newUser, err := us.queries.CreateUser(ctx, user)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &newUser, nil
}

func (us *UserService) Get(ctx context.Context, id string) (*schema.GetUserRow, error) {
	user, err := us.queries.GetUser(ctx, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	user.AccountID.String = "e4694570-f591-4c52-bba9-a5865dc4ba09"
	return &user, nil
}

//func (us *UserService) GetFromContext(ctx context.Context) (*schema.User, error) {
//	ctx.Get()
//}

func (us *UserService) GetUserByEmail(ctx context.Context, email string) (*schema.User, error) {
	userResponse, err := us.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the "no rows" case explicitly
			return nil, nil
		}
		// Handle other errors
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}
	return &userResponse, nil

}

func (us *UserService) GetOrCreate(ctx context.Context, params UserParams) (*schema.User, error) {
	externalID := *params.ExternalID
	var user *schema.User
	// Until we have our own login
	if externalID == "" {
		return nil, fmt.Errorf("externalID is required")
	}

	userResponse, err := us.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if userResponse != nil {
		return userResponse, nil
	}

	user, err = us.Create(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
