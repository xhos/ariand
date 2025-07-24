package queries

import (
	"ariand/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ariand/internal/db"

	"github.com/jmoiron/sqlx"
)

type Categories struct{ db *sqlx.DB }

func NewCategories(db *sqlx.DB) *Categories { return &Categories{db} }

func (c *Categories) ListCategories(ctx context.Context) ([]domain.Category, error) {
	var cats []domain.Category
	query := fmt.Sprintf("SELECT %s FROM categories ORDER BY slug", getCategoryFields())
	err := c.db.SelectContext(ctx, &cats, query)
	return cats, err
}

func (c *Categories) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	var cat domain.Category
	query := fmt.Sprintf("SELECT %s FROM categories WHERE id=$1", getCategoryFields())

	if err := c.db.GetContext(ctx, &cat, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (c *Categories) CreateCategory(ctx context.Context, slug, label, colour string) (int64, error) {
	var id int64
	err := c.db.GetContext(ctx, &id,
		`INSERT INTO categories(slug,label,color) VALUES($1,$2,$3) RETURNING id`,
		slug, label, colour,
	)
	return id, err
}

func (c *Categories) UpdateCategory(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	set, args, err := buildUpdateClauses(fields, allowedCategoryCols)
	if err != nil {
		return err
	}
	args = append(args, id)

	res, err := c.db.ExecContext(ctx,
		fmt.Sprintf(`UPDATE categories SET %s WHERE id=$%d`, set, len(args)),
		args...,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return db.ErrNotFound
	}
	return nil
}

func (c *Categories) DeleteCategory(ctx context.Context, id int64) error {
	res, err := c.db.ExecContext(ctx, `DELETE FROM categories WHERE id=$1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return db.ErrNotFound
	}
	return nil
}

func (c *Categories) ListCategorySlugs(ctx context.Context) ([]string, error) {
	var sl []string
	err := c.db.SelectContext(ctx, &sl, `SELECT slug FROM categories ORDER BY slug`)
	return sl, err
}

func (c *Categories) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var cat domain.Category
	query := fmt.Sprintf("SELECT %s FROM categories WHERE slug=$1", getCategoryFields())

	if err := c.db.GetContext(ctx, &cat, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}
