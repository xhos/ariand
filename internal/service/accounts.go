package service

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"

	"github.com/charmbracelet/log"
)

type AccountService interface {
	List(ctx context.Context) ([]domain.Account, error)
	Get(ctx context.Context, id int64) (*domain.Account, error)
	Create(ctx context.Context, acc *domain.Account) (*domain.Account, error)
	Delete(ctx context.Context, id int64) error
	SetAnchor(ctx context.Context, id int64, bal float64, currency string) error
	Balance(ctx context.Context, id int64) (balance float64, currency string, err error)
}

type acctSvc struct {
	store db.Store
	log   *log.Logger
}

func newAcctSvc(store db.Store, lg *log.Logger) AccountService {
	return &acctSvc{store: store, log: lg}
}

func (s *acctSvc) List(ctx context.Context) ([]domain.Account, error) {
	return s.store.ListAccounts(ctx)
}

func (s *acctSvc) Get(ctx context.Context, id int64) (*domain.Account, error) {
	return s.store.GetAccount(ctx, id)
}

func (s *acctSvc) Create(ctx context.Context, acc *domain.Account) (*domain.Account, error) {
	if acc.AnchorCurrency == "" {
		acc.AnchorCurrency = "CAD" // TODO: we don't force everyone to be canadian, eh?
	}

	return s.store.CreateAccount(ctx, acc)
}

func (s *acctSvc) Delete(ctx context.Context, id int64) error {
	return s.store.DeleteAccount(ctx, id)
}

func (s *acctSvc) SetAnchor(ctx context.Context, id int64, bal float64, currency string) error {
	return s.store.SetAccountAnchor(ctx, id, bal, currency)
}

func (s *acctSvc) Balance(ctx context.Context, id int64) (float64, string, error) {
	return s.store.GetAccountBalance(ctx, id)
}
