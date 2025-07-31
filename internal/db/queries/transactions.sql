-- name: ListTransactions :many
SELECT
  t.id, t.email_id, t.account_id, t.tx_date, t.tx_amount, t.tx_currency,
  t.tx_direction, t.tx_desc, t.balance_after, t.category_id, t.cat_status,
  t.merchant, t.user_notes, t.suggestions, t.receipt_id,
  t.foreign_currency, t.foreign_amount, t.exchange_rate,
  t.created_at, t.updated_at,
  c.slug  AS category_slug,
  c.label AS category_label,
  c.color AS category_color
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE TRUE
  AND (sqlc.narg('cursor_date')::timestamptz IS NULL OR sqlc.narg('cursor_id')::bigint IS NULL
       OR (t.tx_date, t.id) < (sqlc.narg('cursor_date')::timestamptz, sqlc.narg('cursor_id')::bigint))
  AND (sqlc.narg('start')::timestamptz IS NULL OR t.tx_date >= sqlc.narg('start')::timestamptz)
  AND (sqlc.narg('end')::timestamptz   IS NULL OR t.tx_date <= sqlc.narg('end')::timestamptz)
  AND (sqlc.narg('amount_min')::numeric IS NULL OR t.tx_amount >= sqlc.narg('amount_min')::numeric)
  AND (sqlc.narg('amount_max')::numeric IS NULL OR t.tx_amount <= sqlc.narg('amount_max')::numeric)
  AND (sqlc.narg('direction')::text IS NULL OR t.tx_direction = sqlc.narg('direction')::text)
  AND (sqlc.narg('account_ids')::bigint[] IS NULL OR t.account_id = ANY(sqlc.narg('account_ids')::bigint[]))
  AND (sqlc.narg('categories')::text[] IS NULL OR c.slug = ANY(sqlc.narg('categories')::text[]))
  AND (sqlc.narg('merchant_q')::text IS NULL OR t.merchant ILIKE ('%' || sqlc.narg('merchant_q')::text || '%'))
  AND (sqlc.narg('desc_q')::text     IS NULL OR t.tx_desc ILIKE ('%' || sqlc.narg('desc_q')::text || '%'))
  AND (sqlc.narg('currency')::char(3) IS NULL OR t.tx_currency = sqlc.narg('currency')::char(3))
  AND (sqlc.narg('tod_start')::time IS NULL OR t.tx_date::time >= sqlc.narg('tod_start')::time)
  AND (sqlc.narg('tod_end')::time   IS NULL OR t.tx_date::time <= sqlc.narg('tod_end')::time)
ORDER BY t.tx_date DESC, t.id DESC;

-- name: GetTransaction :one
SELECT
  t.id, t.email_id, t.account_id, t.tx_date, t.tx_amount, t.tx_currency,
  t.tx_direction, t.tx_desc, t.balance_after, t.category_id, t.cat_status,
  t.merchant, t.user_notes, t.suggestions, t.receipt_id,
  t.foreign_currency, t.foreign_amount, t.exchange_rate,
  t.created_at, t.updated_at,
  c.slug  AS category_slug,
  c.label AS category_label,
  c.color AS category_color
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.id = @id::bigint;

-- name: CreateTransaction :one
INSERT INTO transactions (
  email_id, account_id, tx_date, tx_amount, tx_currency, tx_direction,
  tx_desc, balance_after, category_id, merchant, user_notes,
  foreign_currency, foreign_amount, exchange_rate, suggestions, receipt_id
) VALUES (
  sqlc.narg('email_id')::text,
  @account_id::bigint,
  @tx_date::timestamptz,
  @tx_amount::numeric,
  @tx_currency::char(3),
  @tx_direction::text,
  sqlc.narg('tx_desc')::text,
  sqlc.narg('balance_after')::numeric,
  sqlc.narg('category_id')::bigint,
  sqlc.narg('merchant')::text,
  sqlc.narg('user_notes')::text,
  sqlc.narg('foreign_currency')::char(3),
  sqlc.narg('foreign_amount')::numeric,
  sqlc.narg('exchange_rate')::numeric,
  sqlc.narg('suggestions')::text[],
  sqlc.narg('receipt_id')::bigint
)
RETURNING id;

-- name: UpdateTransactionPartial :one
UPDATE transactions SET
  email_id         = CASE WHEN @email_id_set::bool         THEN sqlc.narg('email_id')::text           ELSE email_id         END,
  tx_date          = CASE WHEN @tx_date_set::bool          THEN @tx_date::timestamptz                 ELSE tx_date          END,
  tx_amount        = CASE WHEN @tx_amount_set::bool        THEN @tx_amount::numeric                   ELSE tx_amount        END,
  tx_currency      = CASE WHEN @tx_currency_set::bool      THEN @tx_currency::char(3)                 ELSE tx_currency      END,
  tx_direction     = CASE WHEN @tx_direction_set::bool     THEN @tx_direction::text                   ELSE tx_direction     END,
  tx_desc          = CASE WHEN @tx_desc_set::bool          THEN sqlc.narg('tx_desc')::text            ELSE tx_desc          END,
  category_id      = CASE WHEN @category_id_set::bool      THEN sqlc.narg('category_id')::bigint      ELSE category_id      END,
  merchant         = CASE WHEN @merchant_set::bool         THEN sqlc.narg('merchant')::text           ELSE merchant         END,
  user_notes       = CASE WHEN @user_notes_set::bool       THEN sqlc.narg('user_notes')::text         ELSE user_notes       END,
  foreign_currency = CASE WHEN @foreign_currency_set::bool THEN sqlc.narg('foreign_currency')::char(3) ELSE foreign_currency END,
  foreign_amount   = CASE WHEN @foreign_amount_set::bool   THEN sqlc.narg('foreign_amount')::numeric  ELSE foreign_amount   END,
  exchange_rate    = CASE WHEN @exchange_rate_set::bool    THEN sqlc.narg('exchange_rate')::numeric   ELSE exchange_rate    END,
  suggestions      = CASE WHEN @suggestions_set::bool      THEN sqlc.narg('suggestions')::text[]      ELSE suggestions      END,
  receipt_id       = CASE WHEN @receipt_id_set::bool       THEN sqlc.narg('receipt_id')::bigint       ELSE receipt_id       END
WHERE id = @id::bigint
RETURNING account_id;

-- name: DeleteTransactionReturningAccount :one
DELETE FROM transactions
WHERE id = @id::bigint
RETURNING account_id;

-- name: SetTransactionReceipt :execrows
UPDATE transactions
SET receipt_id = @receipt_id::bigint
WHERE id = @id::bigint AND receipt_id IS NULL;

-- name: SyncAccountBalances :exec
WITH transaction_deltas AS (
  SELECT id,
         SUM(CASE WHEN tx_direction = 'incoming' THEN tx_amount ELSE -tx_amount END)
           OVER (PARTITION BY account_id ORDER BY tx_date, id) AS running_delta
  FROM transactions
  WHERE account_id = @account_id::bigint
),
anchor_point AS (
  SELECT a.anchor_balance,
         COALESCE(SUM(CASE WHEN t.tx_direction = 'incoming' THEN t.tx_amount ELSE -t.tx_amount END), 0.0) AS delta_at_anchor
  FROM accounts a
  LEFT JOIN transactions t ON t.account_id = a.id AND t.tx_date < a.anchor_date
  WHERE a.id = @account_id::bigint
  GROUP BY a.id, a.anchor_balance
)
UPDATE transactions
SET balance_after = ap.anchor_balance + td.running_delta - ap.delta_at_anchor
FROM transaction_deltas td, anchor_point ap
WHERE transactions.id = td.id
  AND transactions.account_id = @account_id::bigint;

-- name: FindCandidateTransactions :many
SELECT
  t.id, t.email_id, t.account_id, t.tx_date, t.tx_amount, t.tx_currency,
  t.tx_direction, t.tx_desc, t.balance_after, t.category_id, t.cat_status,
  t.merchant, t.user_notes, t.suggestions, t.receipt_id,
  t.foreign_currency, t.foreign_amount, t.exchange_rate,
  t.created_at, t.updated_at,
  c.slug  AS category_slug,
  c.label AS category_label,
  c.color AS category_color,
  similarity(t.tx_desc::text, sqlc.arg(merchant)::text) AS merchant_score
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.receipt_id IS NULL
  AND t.tx_direction = 'outgoing'
  AND t.tx_date >= (sqlc.arg(date)::date - interval '60 days')
  AND t.tx_amount BETWEEN sqlc.arg(total)::numeric AND (sqlc.arg(total)::numeric * 1.20)
  AND similarity(t.tx_desc::text, sqlc.arg(merchant)::text) > 0.3
ORDER BY merchant_score DESC
LIMIT 10;
