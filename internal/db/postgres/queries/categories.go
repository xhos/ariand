package queries

import (
	"ariand/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ariand/internal/db"
	"github.com/jmoiron/sqlx"
)

// Categories wraps CRUD helpers for the categories table.
type Categories struct{ db *sqlx.DB }

func NewCategories(db *sqlx.DB) *Categories { return &Categories{db} }

// --- list -------------------------------------------------------------------

func (c *Categories) ListCategories(ctx context.Context) ([]domain.Category, error) {
	var cats []domain.Category
	err := c.db.SelectContext(ctx, &cats, `SELECT * FROM categories ORDER BY slug`)
	return cats, err
}

// --- read -------------------------------------------------------------------

func (c *Categories) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	var cat domain.Category
	if err := c.db.GetContext(ctx, &cat, `SELECT * FROM categories WHERE id=$1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

// --- create -----------------------------------------------------------------

func (c *Categories) CreateCategory(ctx context.Context, slug, label, colour string) (int64, error) {
	var id int64
	err := c.db.GetContext(
		ctx,
		&id,
		`INSERT INTO categories(slug,label,color) VALUES($1,$2,$3) RETURNING id`,
		slug, label, colour,
	)
	return id, err
}

// --- update -----------------------------------------------------------------

func (c *Categories) UpdateCategory(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	set, args := make([]string, 0, len(fields)), make([]any, 0, len(fields)+1)
	i := 1
	for k, v := range fields {
		set = append(set, k+"=$"+fmt.Sprint(i))
		args = append(args, v)
		i++
	}
	args = append(args, id)

	query := `UPDATE categories SET ` + strings.Join(set, ", ") + ` WHERE id=$` + fmt.Sprint(i)
	res, err := c.db.ExecContext(ctx, query, args...)
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

// --- delete -----------------------------------------------------------------

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

// --- utility ----------------------------------------------------------------

func (c *Categories) ListCategorySlugs(ctx context.Context) ([]string, error) {
	var sl []string
	err := c.db.SelectContext(ctx, &sl, `SELECT slug FROM categories ORDER BY slug`)
	return sl, err
}

func (c *Categories) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var cat domain.Category
	if err := c.db.GetContext(ctx, &cat, `SELECT * FROM categories WHERE slug=$1`, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}
