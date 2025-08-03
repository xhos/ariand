package service

import (
	sqlc "ariand/internal/db/sqlc"
	"context"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type UserService interface {
	Get(ctx context.Context, id uuid.UUID) (*sqlc.User, error)
	GetByEmail(ctx context.Context, email string) (*sqlc.User, error)
	Create(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error)
	Update(ctx context.Context, params sqlc.UpdateUserParams) (*sqlc.User, error)
	UpdateDisplayName(ctx context.Context, params sqlc.UpdateUserDisplayNameParams) (*sqlc.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]sqlc.User, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type userSvc struct {
	queries *sqlc.Queries
	log     *log.Logger
}

func newUserSvc(queries *sqlc.Queries, lg *log.Logger) UserService {
	return &userSvc{queries: queries, log: lg}
}

func (s *userSvc) Get(ctx context.Context, id uuid.UUID) (*sqlc.User, error) {
	user, err := s.queries.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userSvc) GetByEmail(ctx context.Context, email string) (*sqlc.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userSvc) Create(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error) {
	user, err := s.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userSvc) Update(ctx context.Context, params sqlc.UpdateUserParams) (*sqlc.User, error) {
	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userSvc) UpdateDisplayName(ctx context.Context, params sqlc.UpdateUserDisplayNameParams) (*sqlc.User, error) {
	user, err := s.queries.UpdateUserDisplayName(ctx, params)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userSvc) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.queries.DeleteUser(ctx, id)
	return err
}

func (s *userSvc) List(ctx context.Context) ([]sqlc.User, error) {
	return s.queries.ListUsers(ctx)
}

func (s *userSvc) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.queries.CheckUserExists(ctx, id)
}
