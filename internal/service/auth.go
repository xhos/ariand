package service

import (
	sqlc "ariand/internal/db/sqlc"
	"context"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type AuthService interface {
	// Credential management
	CreateCredential(ctx context.Context, params sqlc.CreateCredentialParams) (*sqlc.UserCredential, error)
	GetCredential(ctx context.Context, id uuid.UUID) (*sqlc.UserCredential, error)
	GetCredentialByCredentialID(ctx context.Context, credentialID []byte) (*sqlc.UserCredential, error)
	GetCredentialForUser(ctx context.Context, params sqlc.GetCredentialForUserParams) (*sqlc.UserCredential, error)
	ListCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.UserCredential, error)
	UpdateCredentialSignCount(ctx context.Context, params sqlc.UpdateCredentialSignCountParams) error
	UpdateCredentialSignCountByCredentialID(ctx context.Context, params sqlc.UpdateCredentialSignCountByCredentialIdParams) error
	DeleteCredentialForUser(ctx context.Context, params sqlc.DeleteCredentialForUserParams) error
	DeleteAllCredentialsForUser(ctx context.Context, userID uuid.UUID) error
	CountCredentialsForUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CheckCredentialExists(ctx context.Context, credentialID []byte) (bool, error)
}

type authSvc struct {
	queries *sqlc.Queries
	log     *log.Logger
}

func newAuthSvc(queries *sqlc.Queries, lg *log.Logger) AuthService {
	return &authSvc{queries: queries, log: lg}
}

func (s *authSvc) CreateCredential(ctx context.Context, params sqlc.CreateCredentialParams) (*sqlc.UserCredential, error) {
	credential, err := s.queries.CreateCredential(ctx, params)
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (s *authSvc) GetCredential(ctx context.Context, id uuid.UUID) (*sqlc.UserCredential, error) {
	credential, err := s.queries.GetCredential(ctx, id)
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (s *authSvc) GetCredentialByCredentialID(ctx context.Context, credentialID []byte) (*sqlc.UserCredential, error) {
	credential, err := s.queries.GetCredentialByCredentialId(ctx, credentialID)
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (s *authSvc) GetCredentialForUser(ctx context.Context, params sqlc.GetCredentialForUserParams) (*sqlc.UserCredential, error) {
	credential, err := s.queries.GetCredentialForUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (s *authSvc) ListCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.UserCredential, error) {
	return s.queries.ListCredentialsByUser(ctx, userID)
}

func (s *authSvc) UpdateCredentialSignCount(ctx context.Context, params sqlc.UpdateCredentialSignCountParams) error {
	_, err := s.queries.UpdateCredentialSignCount(ctx, params)
	return err
}

func (s *authSvc) UpdateCredentialSignCountByCredentialID(ctx context.Context, params sqlc.UpdateCredentialSignCountByCredentialIdParams) error {
	_, err := s.queries.UpdateCredentialSignCountByCredentialId(ctx, params)
	return err
}

func (s *authSvc) DeleteCredentialForUser(ctx context.Context, params sqlc.DeleteCredentialForUserParams) error {
	_, err := s.queries.DeleteCredentialForUser(ctx, params)
	return err
}

func (s *authSvc) DeleteAllCredentialsForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.queries.DeleteAllCredentialsForUser(ctx, userID)
	return err
}

func (s *authSvc) CountCredentialsForUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.queries.CountCredentialsForUser(ctx, userID)
}

func (s *authSvc) CheckCredentialExists(ctx context.Context, credentialID []byte) (bool, error) {
	return s.queries.CheckCredentialExists(ctx, credentialID)
}
