CREATE TABLE IF NOT EXISTS accounts (
  id             SERIAL PRIMARY KEY,
  name           TEXT NOT NULL UNIQUE,
  bank           TEXT NOT NULL,
  type           TEXT NOT NULL,
  alias          TEXT,

  anchor_date    DATE    NOT NULL,
  anchor_balance NUMERIC(18,2) NOT NULL,
  
  created_at     TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transactions (
  id            SERIAL PRIMARY KEY,
  email_id      TEXT NOT NULL UNIQUE,
  account_id    INT  NOT NULL REFERENCES accounts(id),
  
  tx_date       TIMESTAMPTZ NOT NULL,
  tx_amount     NUMERIC(18,2) NOT NULL,
  tx_currency   TEXT NOT NULL,
  tx_direction  TEXT NOT NULL CHECK (tx_direction IN ('in','out')),
  tx_desc       TEXT,
  
  category      TEXT,
  merchant      TEXT,
  user_notes    TEXT,
  
  foreign_currency TEXT,
  foreign_amount   NUMERIC(18,2),
  exchange_rate    NUMERIC(18,6)
);

CREATE INDEX IF NOT EXISTS idx_tx_account_date
  ON transactions(account_id, tx_date);
