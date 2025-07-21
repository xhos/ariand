package service

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"

	"github.com/charmbracelet/log"
)

type CategoryService interface {
	List(ctx context.Context) ([]domain.Category, error)
	Get(ctx context.Context, id int64) (*domain.Category, error)
	Create(ctx context.Context, slug, label, colour string) (int64, error)
	Update(ctx context.Context, id int64, fields map[string]any) error
	Delete(ctx context.Context, id int64) error
	BySlug(ctx context.Context, slug string) (*domain.Category, error)
	ListSlugs(ctx context.Context) ([]string, error)
}

type catSvc struct {
	store db.Store
	log   *log.Logger
}

func newCatSvc(store db.Store, lg *log.Logger) CategoryService {
	return &catSvc{store: store, log: lg}
}

func (s *catSvc) List(ctx context.Context) ([]domain.Category, error) {
	return s.store.ListCategories(ctx)
}

func (s *catSvc) Get(ctx context.Context, id int64) (*domain.Category, error) {
	return s.store.GetCategory(ctx, id)
}

func (s *catSvc) Create(ctx context.Context, slug, label, colour string) (int64, error) {
	return s.store.CreateCategory(ctx, slug, label, colour)
}

func (s *catSvc) Update(ctx context.Context, id int64, fields map[string]any) error {
	return s.store.UpdateCategory(ctx, id, fields)
}

func (s *catSvc) Delete(ctx context.Context, id int64) error {
	return s.store.DeleteCategory(ctx, id)
}

func (s *catSvc) BySlug(ctx context.Context, slug string) (*domain.Category, error) {
	return s.store.GetCategoryBySlug(ctx, slug)
}

func (s *catSvc) ListSlugs(ctx context.Context) ([]string, error) {
	return s.store.ListCategorySlugs(ctx)
}
