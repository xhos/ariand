-- name: GetDashboardBalance :one
SELECT COALESCE(SUM(a.anchor_balance + COALESCE(d.delta, 0)), 0)::double precision AS total_balance
FROM accounts a
LEFT JOIN LATERAL (
  SELECT SUM(
    CASE
      WHEN t.tx_direction = 'incoming' THEN t.tx_amount
      WHEN t.tx_direction = 'outgoing' THEN -t.tx_amount
    END
  ) AS delta
  FROM transactions t
  WHERE t.account_id = a.id
    AND t.tx_date > a.anchor_date
) d ON TRUE;

-- name: GetDashboardDebt :one
SELECT COALESCE(SUM(a.anchor_balance + COALESCE(d.delta, 0)), 0)::double precision AS total_debt
FROM accounts a
LEFT JOIN LATERAL (
  SELECT SUM(
    CASE
      WHEN t.tx_direction = 'incoming' THEN t.tx_amount
      WHEN t.tx_direction = 'outgoing' THEN -t.tx_amount
    END
  ) AS delta
  FROM transactions t
  WHERE t.account_id = a.id
    AND t.tx_date > a.anchor_date
) d ON TRUE
WHERE a.account_type = 'credit_card';

-- name: GetDashboardTrends :many
SELECT
  to_char(tx_date::date, 'YYYY-MM-DD') AS date,
  SUM(CASE WHEN tx_direction = 'incoming' THEN tx_amount ELSE 0 END) AS income,
  SUM(CASE WHEN tx_direction = 'outgoing' THEN tx_amount ELSE 0 END) AS expenses
FROM transactions
WHERE (sqlc.narg('start')::timestamptz IS NULL OR tx_date >= sqlc.narg('start')::timestamptz)
  AND (sqlc.narg('end')::timestamptz   IS NULL OR tx_date <= sqlc.narg('end')::timestamptz)
GROUP BY date
ORDER BY date;
