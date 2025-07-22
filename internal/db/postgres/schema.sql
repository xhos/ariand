/* ---------- EXTENSIONS ------------------------------------- */
CREATE EXTENSION IF NOT EXISTS pg_trgm;

/* ---------- ENUMS & DOMAINS -------------------------------- */
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'account_type') THEN
    CREATE TYPE account_type AS ENUM (
      'chequing','savings','credit_card','investment','other'
    );
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'category_status') THEN
    CREATE TYPE category_status AS ENUM ('auto','ai','user');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'receipt_provider') THEN
    CREATE TYPE receipt_provider AS ENUM ('gemini','local');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'receipt_parse_status') THEN
    CREATE TYPE receipt_parse_status AS ENUM ('pending','parsed','failed');
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'currency_code') THEN
    CREATE DOMAIN currency_code AS CHAR(3)
      CHECK (VALUE ~ '^[A-Z]{3}$');
  END IF;
END$$;

/* ---------- GENERIC updated_at TRIGGER --------------------- */
CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END$$;

/* ---------- ACCOUNTS --------------------------------------- */
CREATE TABLE IF NOT EXISTS accounts (
  id              SERIAL             PRIMARY KEY,
  name            TEXT               NOT NULL UNIQUE,
  bank            TEXT               NOT NULL,
  account_type    account_type       NOT NULL,
  alias           TEXT,
  anchor_date     DATE               NOT NULL DEFAULT CURRENT_DATE,
  anchor_balance  NUMERIC(18,2)      NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_accounts_update'
      AND tgrelid = 'accounts'::regclass
  ) THEN
    CREATE TRIGGER trg_accounts_update
      BEFORE UPDATE ON accounts
      FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
  END IF;
END$$;

/* ---------- CATEGORIES ------------------------------------- */
CREATE TABLE IF NOT EXISTS categories (
  id         SERIAL             PRIMARY KEY,
  slug       TEXT               UNIQUE NOT NULL,
  label      TEXT               NOT NULL,
  color      TEXT               NOT NULL,
  created_at TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_categories_update'
      AND tgrelid = 'categories'::regclass
  ) THEN
    CREATE TRIGGER trg_categories_update
      BEFORE UPDATE ON categories
      FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
  END IF;
END$$;

/* ---------- TRANSACTIONS ----------------------------------- */
CREATE TABLE IF NOT EXISTS transactions (
  id               SERIAL             PRIMARY KEY,
  account_id       INT                NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,

  email_id         TEXT,                            -- may be NULL
  tx_date          TIMESTAMPTZ        NOT NULL,
  tx_amount        NUMERIC(18,2)      NOT NULL,
  tx_currency      currency_code      NOT NULL DEFAULT 'CAD',
  tx_direction     TEXT               NOT NULL CHECK (tx_direction IN ('in','out')),
  tx_desc          TEXT,

  balance_after    NUMERIC(18,2),

  merchant         TEXT,
  category_id      INT                REFERENCES categories(id),
  cat_status       category_status    NOT NULL DEFAULT 'auto',
  suggestions      TEXT[],

  user_notes       TEXT,

  foreign_currency currency_code,
  foreign_amount   NUMERIC(18,2),
  exchange_rate    NUMERIC(18,6),

  receipt_id       INT UNIQUE,       -- link to receipts

  created_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_transactions_update'
      AND tgrelid = 'transactions'::regclass
  ) THEN
    CREATE TRIGGER trg_transactions_update
      BEFORE UPDATE ON transactions
      FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
  END IF;
END$$;

CREATE UNIQUE INDEX IF NOT EXISTS ux_transactions_email_id_notnull
  ON transactions(email_id) WHERE email_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tx_account_date
  ON transactions(account_id, tx_date);

CREATE INDEX IF NOT EXISTS idx_tx_desc_trgm
  ON transactions USING gin (lower(tx_desc) gin_trgm_ops);

/* ---------- RECEIPTS --------------------------------------- */
CREATE TABLE IF NOT EXISTS receipts (
  id               SERIAL               PRIMARY KEY,
  transaction_id   INT UNIQUE           REFERENCES transactions(id) ON DELETE CASCADE,

  provider         receipt_provider     NOT NULL,
  parse_status     receipt_parse_status NOT NULL DEFAULT 'pending',

  merchant         TEXT,
  purchase_date    DATE,
  total_amount     NUMERIC(18,2),

  currency         currency_code,
  tax_amount       NUMERIC(18,2),

  raw_payload      JSONB,
  canonical_data   JSONB,

  image_url        TEXT,
  image_sha256     BYTEA,

  lat              DOUBLE PRECISION,
  lon              DOUBLE PRECISION,
  location_source  TEXT,
  location_label   TEXT,

  created_at       TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ          NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_receipts_update'
      AND tgrelid = 'receipts'::regclass
  ) THEN
    CREATE TRIGGER trg_receipts_update
      BEFORE UPDATE ON receipts
      FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
  END IF;
END$$;

CREATE INDEX IF NOT EXISTS idx_receipts_purchase_date
  ON receipts(purchase_date);

CREATE INDEX IF NOT EXISTS idx_receipts_merchant_trgm
  ON receipts USING gin (lower(merchant) gin_trgm_ops);

/* ---------- FOREIGN KEY FROM transactions to receipts -------- */
/* ─── FOREIGN KEY FROM transactions → receipts (idempotent) ─── */
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
      FROM pg_constraint
     WHERE conname = 'fk_tx_receipt'
       AND conrelid = 'transactions'::regclass
  ) THEN
    ALTER TABLE transactions
      ADD CONSTRAINT fk_tx_receipt
        FOREIGN KEY (receipt_id)
        REFERENCES receipts(id)
        ON DELETE SET NULL;
  END IF;
END
$$;


/* ---------- RECEIPT ITEMS ---------------------------------- */
CREATE TABLE IF NOT EXISTS receipt_items (
  id            SERIAL             PRIMARY KEY,
  receipt_id    INT                NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
  line_no       INT,
  name          TEXT               NOT NULL,
  qty           NUMERIC(18,4)      DEFAULT 1,
  unit_price    NUMERIC(18,4),
  line_total    NUMERIC(18,4),
  sku           TEXT,
  category_hint TEXT,

  created_at    TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_receipt_items_update'
      AND tgrelid = 'receipt_items'::regclass
  ) THEN
    CREATE TRIGGER trg_receipt_items_update
      BEFORE UPDATE ON receipt_items
      FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
  END IF;
END$$;

CREATE INDEX IF NOT EXISTS idx_receipt_items_receipt_id
  ON receipt_items(receipt_id);

CREATE INDEX IF NOT EXISTS idx_receipt_items_name_trgm
  ON receipt_items USING gin (lower(name) gin_trgm_ops);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
      FROM pg_constraint
     WHERE conname = 'fk_tx_receipt'
       AND conrelid = 'transactions'::regclass
  ) THEN
    ALTER TABLE transactions
      ADD CONSTRAINT fk_tx_receipt
        FOREIGN KEY (receipt_id)
        REFERENCES receipts(id)
        ON DELETE SET NULL;
  END IF;
END$$;
