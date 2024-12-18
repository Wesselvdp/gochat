package services

import (
	"context"
	"fmt"
	database "gochat/internal/db"
	"gochat/internal/schema"
)

type AccountService struct {
	queries *schema.Queries
}

type AddDomainParams struct {
	domain    string
	accountID string
}

func NewAccountService() *AccountService {
	queries, _, err := database.Init()
	if err != nil {
		fmt.Println("Error initializing queries for account service: " + err.Error())
		return nil
	}
	return &AccountService{queries: queries}
}

func (as *AccountService) Get(ctx context.Context, id string) (*schema.GetAccountRow, error) {
	account, err := as.queries.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (as *AccountService) Create(c context.Context, accountData schema.CreateAccountParams) (*schema.Account, error) {
	account, err := as.queries.CreateAccount(c, accountData)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (as *AccountService) CreateAccountDomain(c context.Context, params schema.CreateAccountDomainParams) (*schema.AccountDomain, error) {
	accountDomain, err := as.queries.CreateAccountDomain(c, params)
	if err != nil {
		return nil, err
	}
	return &accountDomain, nil
}

func (as *AccountService) DeleteAccountDomain(c context.Context, domain string) error {
	err := as.queries.DeleteAccountDomain(c, domain)
	if err != nil {
		return err
	}
	return nil
}

func (as *AccountService) UpdateUserAccount(c context.Context, params schema.UpdateUserAccountParams) error {
	err := as.queries.UpdateUserAccount(c, params)
	if err != nil {
		return err
	}
	return nil
}
