package services

import (
	"context"

	httpclient "github.com/swastiijain24/bank/internals/http_client"
)

type AccountService interface {
	DiscoverAccounts(ctx context.Context, phone string, bankCode string) ([]string, error)
	SetMpin(ctx context.Context, accountId string, bankCode string, mpinEn string) error
	ChangeMpin(ctx context.Context, accountId string, bankCode string, oldMpinEn string, newMpinEn string) error
	GetBalance(ctx context.Context, accountId string, bankCode string, mpinEn string) (int64, error)
}

type AccountSvc struct {
	bankClient httpclient.HttpClient
}

func NewAccountService(bankClient httpclient.HttpClient) AccountService {
	return &AccountSvc{
		bankClient: bankClient,
	}
}

func (s *AccountSvc) DiscoverAccounts(ctx context.Context, phone string, bankCode string) ([]string, error) {
	return s.bankClient.DiscoverAccounts(ctx, phone)
}

func (s *AccountSvc) SetMpin(ctx context.Context, accountId string, bankCode string, mpinEn string) error {
	return s.bankClient.SetMpin(ctx, accountId, mpinEn)
}

func (s *AccountSvc) ChangeMpin(ctx context.Context, accountId string, bankCode string, oldMpinEn string, newMpinEn string) error {
	return s.bankClient.ChangeMpin(ctx, accountId, oldMpinEn, newMpinEn)
}

func (s *AccountSvc) GetBalance(ctx context.Context, accountId string, bankCode string, mpinEn string) (int64, error) {
	return s.bankClient.GetBalance(ctx, accountId, mpinEn)
}
