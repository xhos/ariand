package service

import (
	sqlc "ariand/internal/db/sqlc"
	"context"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type CategoryService interface {
	List(ctx context.Context) ([]sqlc.Category, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListCategoriesForUserRow, error)
	Get(ctx context.Context, id int64) (*sqlc.Category, error)
	Create(ctx context.Context, params sqlc.CreateCategoryParams) (*sqlc.Category, error)
	Update(ctx context.Context, params sqlc.UpdateCategoryParams) (*sqlc.Category, error)
	Delete(ctx context.Context, id int64) error
	BySlug(ctx context.Context, slug string) (*sqlc.Category, error)
	ListSlugs(ctx context.Context) ([]string, error)
}

type catSvc struct {
	queries *sqlc.Queries
	log     *log.Logger
}

func newCatSvc(queries *sqlc.Queries, lg *log.Logger) CategoryService {
	return &catSvc{queries: queries, log: lg}
}

func (s *catSvc) List(ctx context.Context) ([]sqlc.Category, error) {
	return s.queries.ListCategories(ctx)
}

func (s *catSvc) ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListCategoriesForUserRow, error) {
	return s.queries.ListCategoriesForUser(ctx, userID)
}

func (s *catSvc) Get(ctx context.Context, id int64) (*sqlc.Category, error) {
	category, err := s.queries.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *catSvc) Create(ctx context.Context, params sqlc.CreateCategoryParams) (*sqlc.Category, error) {
	category, err := s.queries.CreateCategory(ctx, params)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *catSvc) Update(ctx context.Context, params sqlc.UpdateCategoryParams) (*sqlc.Category, error) {
	category, err := s.queries.UpdateCategory(ctx, params)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *catSvc) Delete(ctx context.Context, id int64) error {
	_, err := s.queries.DeleteCategory(ctx, id)
	return err
}

func (s *catSvc) BySlug(ctx context.Context, slug string) (*sqlc.Category, error) {
	category, err := s.queries.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *catSvc) ListSlugs(ctx context.Context) ([]string, error) {
	return s.queries.ListCategorySlugs(ctx)
}
