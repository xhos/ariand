package service

import (
	sqlc "ariand/internal/db/sqlc"
	"context"
	"database/sql"
	"errors"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type AuthService interface {
	CreateCredential(ctx context.Context, params sqlc.CreateCredentialParams) (*sqlc.UserCredential, error)
	GetCredential(ctx context.Context, id uuid.UUID) (*sqlc.UserCredential, error)
	GetCredentialByCredentialID(ctx context.Context, credentialID []byte) (*sqlc.UserCredential, error)
	GetCredentialForUser(ctx context.Context, params sqlc.GetCredentialForUserParams) (*sqlc.UserCredential, error)
	ListCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.UserCredential, error)
	UpdateCredentialSignCount(ctx context.Context, params sqlc.UpdateCredentialSignCountParams) error
	UpdateCredentialSignCountByCredentialID(ctx context.Context, params sqlc.UpdateCredentialSignCountByCredentialIdParams) error
	DeleteCredentialForUser(ctx context.Context, params sqlc.DeleteCredentialForUserParams, callerID uuid.UUID) error
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
	s.log.Info("CreateCredential", "user", params.UserID, "cred_id_len", len(params.CredentialID))
	credential, err := s.queries.CreateCredential(ctx, params)
	if err != nil {
		return nil, wrapErr("AuthService.CreateCredential", err)
	}
	return &credential, nil
}

func (s *authSvc) GetCredential(ctx context.Context, id uuid.UUID) (*sqlc.UserCredential, error) {
	credential, err := s.queries.GetCredential(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wrapErr("AuthService.GetCredential", ErrNotFound)
	}

	if err != nil {
		return nil, wrapErr("AuthService.GetCredential", err)
	}

	return &credential, nil
}

func (s *authSvc) GetCredentialByCredentialID(ctx context.Context, credentialID []byte) (*sqlc.UserCredential, error) {
	credential, err := s.queries.GetCredentialByCredentialId(ctx, credentialID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wrapErr("AuthService.GetCredentialByCredentialID", ErrNotFound)
	}

	if err != nil {
		return nil, wrapErr("AuthService.GetCredentialByCredentialID", err)
	}

	return &credential, nil
}

func (s *authSvc) GetCredentialForUser(ctx context.Context, params sqlc.GetCredentialForUserParams) (*sqlc.UserCredential, error) {
	credential, err := s.queries.GetCredentialForUser(ctx, params)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wrapErr("AuthService.GetCredentialForUser", ErrNotFound)
	}

	if err != nil {
		return nil, wrapErr("AuthService.GetCredentialForUser", err)
	}

	return &credential, nil
}

func (s *authSvc) ListCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.UserCredential, error) {
	credentials, err := s.queries.ListCredentialsByUser(ctx, userID)
	if err != nil {
		return nil, wrapErr("AuthService.ListCredentialsByUser", err)
	}
	return credentials, nil
}

func (s *authSvc) UpdateCredentialSignCount(ctx context.Context, params sqlc.UpdateCredentialSignCountParams) error {
	_, err := s.queries.UpdateCredentialSignCount(ctx, params)
	if err != nil {
		return wrapErr("AuthService.UpdateCredentialSignCount", err)
	}
	return nil
}

func (s *authSvc) UpdateCredentialSignCountByCredentialID(ctx context.Context, params sqlc.UpdateCredentialSignCountByCredentialIdParams) error {
	_, err := s.queries.UpdateCredentialSignCountByCredentialId(ctx, params)
	if err != nil {
		return wrapErr("AuthService.UpdateCredentialSignCountByCredentialID", err)
	}
	return nil
}

func (s *authSvc) DeleteCredentialForUser(ctx context.Context, params sqlc.DeleteCredentialForUserParams, callerID uuid.UUID) error {
	credential, err := s.queries.GetCredentialForUser(ctx, sqlc.GetCredentialForUserParams(params))
	if errors.Is(err, sql.ErrNoRows) {
		return wrapErr("AuthService.DeleteCredentialForUser", ErrNotFound)
	}

	if err != nil {
		return wrapErr("AuthService.DeleteCredentialForUser", err)
	}

	if credential.UserID != callerID {
		s.log.Warn("DeleteCredentialForUser permission denied", "user", credential.UserID, "caller", callerID)
		return wrapErr("AuthService.DeleteCredentialForUser", ErrPermission)
	}

	_, err = s.queries.DeleteCredentialForUser(ctx, params)
	if err != nil {
		return wrapErr("AuthService.DeleteCredentialForUser", err)
	}

	return nil
}

func (s *authSvc) DeleteAllCredentialsForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.queries.DeleteAllCredentialsForUser(ctx, userID)
	if err != nil {
		return wrapErr("AuthService.DeleteAllCredentialsForUser", err)
	}
	return nil
}

func (s *authSvc) CountCredentialsForUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.queries.CountCredentialsForUser(ctx, userID)
	if err != nil {
		return 0, wrapErr("AuthService.CountCredentialsForUser", err)
	}
	return count, nil
}

func (s *authSvc) CheckCredentialExists(ctx context.Context, credentialID []byte) (bool, error) {
	exists, err := s.queries.CheckCredentialExists(ctx, credentialID)
	if err != nil {
		return false, wrapErr("AuthService.CheckCredentialExists", err)
	}
	return exists, nil
}
