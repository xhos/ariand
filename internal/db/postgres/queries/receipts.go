package queries

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Receipts wraps CRUD helpers for the receipts and receipt_items tables.
type Receipts struct{ db *sqlx.DB }

// NewReceipts creates a new instance of the Receipts query helper.
func NewReceipts(db *sqlx.DB) *Receipts { return &Receipts{db} }

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func (q *Receipts) GetReceipt(ctx context.Context, id int64) (*domain.Receipt, error) {
	var r domain.Receipt
	query := fmt.Sprintf("SELECT %s FROM receipts WHERE id=$1", getReceiptFields())

	if err := q.db.GetContext(ctx, &r, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}

	// Load receipt items, ordering by line_no (NULLs last) then id
	if err := q.db.SelectContext(
		ctx, &r.Items,
		`SELECT * FROM receipt_items WHERE receipt_id=$1 ORDER BY line_no NULLS LAST, id`,
		id,
	); err != nil {
		return nil, err
	}

	return &r, nil
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func (q *Receipts) CreateReceipt(ctx context.Context, r *domain.Receipt) (*domain.Receipt, error) {
	tx, err := q.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	insert := `
		INSERT INTO receipts (
			transaction_id, provider, parse_status, merchant, purchase_date,
			total_amount, currency, tax_amount, raw_payload, canonical_data,
			image_url, image_sha256, lat, lon, location_source, location_label
		) VALUES (
			:transaction_id, :provider, :parse_status, :merchant, :purchase_date,
			:total_amount, :currency, :tax_amount, :raw_payload, :canonical_data,
			:image_url, :image_sha256, :lat, :lon, :location_source, :location_label
		) RETURNING id
	`

	// Prepare a named statement and retrieve the new ID
	stmt, err := tx.PrepareNamedContext(ctx, insert)
	if err != nil {
		return nil, fmt.Errorf("prepare receipt insert: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &r.ID, r); err != nil {
		return nil, fmt.Errorf("execute receipt insert: %w", err)
	}

	// Insert items (if any)
	if len(r.Items) > 0 {
		for i := range r.Items {
			r.Items[i].ReceiptID = r.ID
		}
		itemInsert := `
			INSERT INTO receipt_items (
				receipt_id, line_no, name, qty, unit_price,
				line_total, sku, category_hint
			) VALUES (
				:receipt_id, :line_no, :name, :qty, :unit_price,
				:line_total, :sku, :category_hint
			)
		`
		if _, err := tx.NamedExecContext(ctx, itemInsert, r.Items); err != nil {
			return nil, fmt.Errorf("insert receipt items: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Re-load to populate CreatedAt/UpdatedAt and items
	return q.GetReceipt(ctx, r.ID)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (q *Receipts) UpdateReceipt(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	set, args, err := buildUpdateClauses(fields, allowedReceiptCols)
	if err != nil {
		return err
	}
	args = append(args, id)

	res, err := q.db.ExecContext(
		ctx,
		fmt.Sprintf("UPDATE receipts SET %s WHERE id=$%d", set, len(args)),
		args...,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return db.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func (q *Receipts) DeleteReceipt(ctx context.Context, id int64) error {
	// CASCADE on receipt_items handles child rows
	res, err := q.db.ExecContext(ctx, "DELETE FROM receipts WHERE id=$1", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return db.ErrNotFound
	}
	return nil
}
