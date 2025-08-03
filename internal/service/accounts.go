package service

import (
	sqlc "ariand/internal/db/sqlc"
	"context"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type AccountService interface {
	// User-scoped operations
	ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListAccountsForUserRow, error)
	GetForUser(ctx context.Context, params sqlc.GetAccountForUserParams) (*sqlc.GetAccountForUserRow, error)
	Create(ctx context.Context, params sqlc.CreateAccountParams) (*sqlc.Account, error)
	Update(ctx context.Context, params sqlc.UpdateAccountParams) (*sqlc.Account, error)
	DeleteForUser(ctx context.Context, params sqlc.DeleteAccountForUserParams) error
	GetUserAccountsCount(ctx context.Context, userID uuid.UUID) (int64, error)
	CheckUserAccountAccess(ctx context.Context, params sqlc.CheckUserAccountAccessParams) (bool, error)

	// Account balance operations
	GetAnchorBalance(ctx context.Context, id int64) (*sqlc.GetAccountAnchorBalanceRow, error)
	GetBalance(ctx context.Context, accountID int64) (*decimal.Decimal, error)

	// Collaboration management
	AddCollaborator(ctx context.Context, params sqlc.AddAccountCollaboratorParams) (*sqlc.AccountUser, error)
	RemoveCollaborator(ctx context.Context, params sqlc.RemoveAccountCollaboratorParams) error
	ListCollaborators(ctx context.Context, params sqlc.ListAccountCollaboratorsParams) ([]sqlc.ListAccountCollaboratorsRow, error)
	GetCollaboratorCount(ctx context.Context, accountID int64) (int64, error)
	LeaveCollaboration(ctx context.Context, params sqlc.LeaveAccountCollaborationParams) error
	ListUserCollaborations(ctx context.Context, userID uuid.UUID) ([]sqlc.ListUserCollaborationsRow, error)
	RemoveUserFromAllAccounts(ctx context.Context, userID uuid.UUID) error
}

type acctSvc struct {
	queries *sqlc.Queries
	log     *log.Logger
}

// WithTx creates a new service instance with a transaction
func (s *acctSvc) WithTx(tx pgx.Tx) AccountService {
	return &acctSvc{
		queries: s.queries.WithTx(tx),
		log:     s.log,
	}
}

func newAcctSvc(queries *sqlc.Queries, lg *log.Logger) AccountService {
	return &acctSvc{queries: queries, log: lg}
}

func (s *acctSvc) ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListAccountsForUserRow, error) {
	return s.queries.ListAccountsForUser(ctx, userID)
}

func (s *acctSvc) GetForUser(ctx context.Context, params sqlc.GetAccountForUserParams) (*sqlc.GetAccountForUserRow, error) {
	account, err := s.queries.GetAccountForUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *acctSvc) Create(ctx context.Context, params sqlc.CreateAccountParams) (*sqlc.Account, error) {
	if params.AnchorCurrency == "" {
		params.AnchorCurrency = "CAD" // force everyone to be canadian, eh?
	}

	if params.AnchorBalance.IsZero() {
		params.AnchorBalance = decimal.NewFromInt(0)
	}

	created, err := s.queries.CreateAccount(ctx, params)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *acctSvc) Update(ctx context.Context, params sqlc.UpdateAccountParams) (*sqlc.Account, error) {
	updated, err := s.queries.UpdateAccount(ctx, params)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (s *acctSvc) DeleteForUser(ctx context.Context, params sqlc.DeleteAccountForUserParams) error {
	_, err := s.queries.DeleteAccountForUser(ctx, params)
	return err
}

func (s *acctSvc) GetUserAccountsCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.queries.GetUserAccountsCount(ctx, userID)
}

func (s *acctSvc) CheckUserAccountAccess(ctx context.Context, params sqlc.CheckUserAccountAccessParams) (bool, error) {
	return s.queries.CheckUserAccountAccess(ctx, params)
}

func (s *acctSvc) GetAnchorBalance(ctx context.Context, id int64) (*sqlc.GetAccountAnchorBalanceRow, error) {
	result, err := s.queries.GetAccountAnchorBalance(ctx, id)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *acctSvc) GetBalance(ctx context.Context, accountID int64) (*decimal.Decimal, error) {
	return s.queries.GetAccountBalance(ctx, accountID)
}

func (s *acctSvc) AddCollaborator(ctx context.Context, params sqlc.AddAccountCollaboratorParams) (*sqlc.AccountUser, error) {
	collaborator, err := s.queries.AddAccountCollaborator(ctx, params)
	if err != nil {
		return nil, err
	}
	return &collaborator, nil
}

func (s *acctSvc) RemoveCollaborator(ctx context.Context, params sqlc.RemoveAccountCollaboratorParams) error {
	_, err := s.queries.RemoveAccountCollaborator(ctx, params)
	return err
}

func (s *acctSvc) ListCollaborators(ctx context.Context, params sqlc.ListAccountCollaboratorsParams) ([]sqlc.ListAccountCollaboratorsRow, error) {
	return s.queries.ListAccountCollaborators(ctx, params)
}

func (s *acctSvc) GetCollaboratorCount(ctx context.Context, accountID int64) (int64, error) {
	return s.queries.GetAccountCollaboratorCount(ctx, accountID)
}

func (s *acctSvc) LeaveCollaboration(ctx context.Context, params sqlc.LeaveAccountCollaborationParams) error {
	_, err := s.queries.LeaveAccountCollaboration(ctx, params)
	return err
}

func (s *acctSvc) ListUserCollaborations(ctx context.Context, userID uuid.UUID) ([]sqlc.ListUserCollaborationsRow, error) {
	return s.queries.ListUserCollaborations(ctx, userID)
}

func (s *acctSvc) RemoveUserFromAllAccounts(ctx context.Context, userID uuid.UUID) error {
	_, err := s.queries.RemoveUserFromAllAccounts(ctx, userID)
	return err
}
