package queries

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Dashboard struct {
	db *sqlx.DB
}

func NewDashboard(db *sqlx.DB) *Dashboard {
	return &Dashboard{db}
}

func (q *Dashboard) GetDashboardBalance(ctx context.Context) (float64, error) {
	var v float64
	query := `
		SELECT COALESCE(SUM(a.anchor_balance + COALESCE(d.delta, 0)), 0) AS total_balance
		FROM accounts a
		LEFT JOIN LATERAL (
			SELECT SUM(
				CASE
					WHEN t.tx_direction = 'in' THEN t.tx_amount
					WHEN t.tx_direction = 'out' THEN -t.tx_amount
				END
			) AS delta
			FROM transactions t
			WHERE t.account_id = a.id AND t.tx_date > a.anchor_date
		) d ON TRUE
	`
	err := q.db.GetContext(ctx, &v, query)
	return v, err
}

func (q *Dashboard) GetDashboardDebt(ctx context.Context) (float64, error) {
	var v float64
	query := `
		SELECT COALESCE(SUM(a.anchor_balance + COALESCE(d.delta, 0)), 0) AS total_debt
		FROM accounts a
		LEFT JOIN LATERAL (
			SELECT SUM(
				CASE
					WHEN t.tx_direction = 'in' THEN t.tx_amount
					WHEN t.tx_direction = 'out' THEN -t.tx_amount
				END
			) AS delta
			FROM transactions t
			WHERE t.account_id = a.id AND t.tx_date > a.anchor_date
		) d ON TRUE
		WHERE a.account_type = 'credit_card'
	`
	err := q.db.GetContext(ctx, &v, query)
	return v, err
}

func (q *Dashboard) GetDashboardTrends(ctx context.Context, opt db.ListOpts) ([]domain.TrendPoint, error) {
	var args []any
	var whereClauses []string

	if opt.Start != nil {
		args = append(args, *opt.Start)
		whereClauses = append(whereClauses, fmt.Sprintf("tx_date >= $%d", len(args)))
	}
	if opt.End != nil {
		args = append(args, *opt.End)
		whereClauses = append(whereClauses, fmt.Sprintf("tx_date <= $%d", len(args)))
	}

	query := `
		SELECT
			to_char(tx_date::date, 'YYYY-MM-DD') AS date,
			SUM(CASE WHEN tx_direction = 'in' THEN tx_amount ELSE 0 END) AS income,
			SUM(CASE WHEN tx_direction = 'out' THEN tx_amount ELSE 0 END) AS expenses
		FROM transactions
	`

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += `
		GROUP BY date
		ORDER BY date
	`

	var pts []domain.TrendPoint
	if err := q.db.SelectContext(ctx, &pts, query, args...); err != nil {
		return nil, err
	}
	return pts, nil
}
