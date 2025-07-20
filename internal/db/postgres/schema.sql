-- Enable trigram search (needed for fuzzy description lookups)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- =====================================
--            accounts
-- =====================================
CREATE TABLE IF NOT EXISTS accounts (
  id             SERIAL PRIMARY KEY,
  name           TEXT UNIQUE NOT NULL,
  bank           TEXT NOT NULL,
  type           TEXT NOT NULL,
  alias          TEXT,
  anchor_date    DATE          NOT NULL,
  anchor_balance NUMERIC(18,2) NOT NULL,
  created_at     TIMESTAMPTZ   DEFAULT now()
);

-- =====================================
--         category metadata
-- =====================================
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'category_status') THEN
    CREATE TYPE category_status AS ENUM ('auto','ai','user');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS categories (
  id    SERIAL PRIMARY KEY,
  slug  TEXT UNIQUE NOT NULL,   -- e.g. 'food.takeout'
  label TEXT         NOT NULL,  -- user-friendly
  color TEXT         NOT NULL   -- hex assigned server-side
);

-- =====================================
--            transactions
-- =====================================
CREATE TABLE IF NOT EXISTS transactions (
  id               SERIAL PRIMARY KEY,
  email_id         TEXT UNIQUE NOT NULL,
  account_id       INT  NOT NULL REFERENCES accounts(id),

  tx_date          TIMESTAMPTZ   NOT NULL,
  tx_amount        NUMERIC(18,2) NOT NULL,
  tx_currency      TEXT          NOT NULL,
  tx_direction     TEXT          NOT NULL CHECK (tx_direction IN ('in','out')),
  tx_desc          TEXT,
  balance_after    NUMERIC(18,2) NOT NULL,

  -- AI / user enrichment
  merchant         TEXT,
  category_id      INT REFERENCES categories(id),
  cat_status       category_status NOT NULL DEFAULT 'auto',
  suggestions      TEXT[],

  user_notes       TEXT,

  foreign_currency TEXT,
  foreign_amount   NUMERIC(18,2),
  exchange_rate    NUMERIC(18,6)
);

-- =====================================
--                 indexes
-- =====================================
-- used in account ledger queries
CREATE INDEX IF NOT EXISTS idx_tx_account_date
  ON transactions(account_id, tx_date);

-- trigram index for fast fuzzy search on descriptions
CREATE INDEX IF NOT EXISTS idx_tx_desc_trgm
  ON transactions USING gin (lower(tx_desc) gin_trgm_ops);
