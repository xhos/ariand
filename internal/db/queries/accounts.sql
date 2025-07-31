-- name: ListAccounts :many
SELECT id, name, bank, account_type, alias,
       anchor_date, anchor_balance, anchor_currency,
       created_at, updated_at
FROM accounts
ORDER BY created_at;

-- name: GetAccount :one
SELECT id, name, bank, account_type, alias,
       anchor_date, anchor_balance, anchor_currency,
       created_at, updated_at
FROM accounts
WHERE id = @id::bigint;

-- name: CreateAccount :one
INSERT INTO accounts (name, bank, account_type, alias, anchor_balance, anchor_currency)
VALUES (@name::text, @bank::text, @account_type::text,
        sqlc.narg('alias')::text,
        @anchor_balance::numeric, @anchor_currency::char(3))
RETURNING id, name, bank, account_type, alias,
          anchor_date, anchor_balance, anchor_currency,
          created_at, updated_at;

-- name: DeleteAccount :execrows
DELETE FROM accounts WHERE id = @id::bigint;

-- name: SetAccountAnchor :execrows
UPDATE accounts
SET anchor_date     = NOW()::date,
    anchor_balance  = @anchor_balance::numeric,
    anchor_currency = @anchor_currency::char(3)
WHERE id = @id::bigint;

-- name: UpdateAccountPartial :one
UPDATE accounts
SET name            = CASE WHEN @name_set::bool            THEN @name::text               ELSE name            END,
    bank            = CASE WHEN @bank_set::bool            THEN @bank::text               ELSE bank            END,
    account_type    = CASE WHEN @type_set::bool            THEN @account_type::text       ELSE account_type    END,
    alias           = CASE WHEN @alias_set::bool           THEN sqlc.narg('alias')::text  ELSE alias           END,
    anchor_date     = CASE WHEN @anchor_date_set::bool     THEN @anchor_date::date        ELSE anchor_date     END,
    anchor_balance  = CASE WHEN @anchor_balance_set::bool  THEN @anchor_balance::numeric  ELSE anchor_balance  END,
    anchor_currency = CASE WHEN @anchor_currency_set::bool THEN @anchor_currency::char(3) ELSE anchor_currency END
WHERE id = @id::bigint
RETURNING id, name, bank, account_type, alias,
          anchor_date, anchor_balance, anchor_currency,
          created_at, updated_at;

-- name: GetAccountBalance :one
SELECT balance_after 
FROM transactions 
WHERE account_id = @account_id::bigint 
ORDER BY tx_date DESC, id DESC 
LIMIT 1;

-- name: GetAccountAnchorBalance :one
SELECT anchor_balance 
FROM accounts 
WHERE id = @id::bigint;
