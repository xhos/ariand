package queries

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Dashboard struct {
	db *sqlx.DB
}

func NewDashboard(db *sqlx.DB) *Dashboard {
	return &Dashboard{db}
}

func (q *Transactions) GetDashboardBalance(ctx context.Context) (float64, error) {
	var v float64
	err := q.db.GetContext(ctx, &v, `
		SELECT SUM(a.anchor_balance + COALESCE(d.delta,0)) AS total_balance
		  FROM accounts a
		  LEFT JOIN LATERAL (
		      SELECT SUM(CASE
		                WHEN t.tx_direction = 'in'  THEN  t.tx_amount
		                WHEN t.tx_direction = 'out' THEN -t.tx_amount
		               END) AS delta
		        FROM transactions t
		       WHERE t.account_id = a.id
		         AND t.tx_date    > a.anchor_date
		  ) d ON TRUE`)
	return v, err
}

func (q *Transactions) GetDashboardDebt(ctx context.Context) (float64, error) {
	var v float64
	err := q.db.GetContext(ctx, &v, `
		SELECT COALESCE(SUM(a.anchor_balance + COALESCE(d.delta,0)),0) AS total_debt
		  FROM accounts a
		  LEFT JOIN LATERAL (
		      SELECT SUM(CASE
		                WHEN t.tx_direction = 'out' THEN  t.tx_amount
		                WHEN t.tx_direction = 'in'  THEN -t.tx_amount
		               END) AS delta
		        FROM transactions t
		       WHERE t.account_id = a.id
		         AND t.tx_date    > a.anchor_date
		  ) d ON TRUE
		 WHERE a.type = 'credit_card'`)
	return v, err
}

func (q *Dashboard) GetDashboardTrends(ctx context.Context, opt db.ListOpts) ([]domain.TrendPoint, error) {
	where, args := []string{}, []any{}
	if opt.Start != nil {
		where = append(where, "tx_date >= $1")
		args = append(args, *opt.Start)
	}
	if opt.End != nil {
		where = append(where, "tx_date <= $2")
		args = append(args, *opt.End)
	}
	clause := ""
	if len(where) > 0 {
		clause = "WHERE " + strings.Join(where, " AND ")
	}

	sql := `
	SELECT to_char(tx_date::date,'YYYY-MM-DD') AS date,
	       SUM(CASE WHEN tx_direction='in'  THEN tx_amount ELSE 0 END) AS income,
	       SUM(CASE WHEN tx_direction='out' THEN tx_amount ELSE 0 END) AS expense
	  FROM transactions ` + clause + `
	 GROUP BY date
	 ORDER BY date`

	var pts []domain.TrendPoint
	if err := q.db.SelectContext(ctx, &pts, sql, args...); err != nil {
		return nil, err
	}
	return pts, nil
}
