-- name: ListReceipts :many
SELECT
  id,
  engine,
  parse_status,
  link_status,
  match_ids,
  merchant,
  purchase_date,
  total_amount,
  currency,
  tax_amount,
  raw_payload,
  canonical_data,
  image_url,
  image_sha256,
  lat,
  lon,
  location_source,
  location_label,
  created_at,
  updated_at
FROM receipts
ORDER BY created_at DESC;

-- name: GetReceipt :one
SELECT
  id,
  engine,             -- text enum: 'gemini' | 'local'
  parse_status,       -- 'pending' | 'success' | 'failed'
  link_status,        -- 'unlinked' | 'matched' | 'needs_verification'
  match_ids,          -- bigint[]
  merchant,
  purchase_date,
  total_amount,
  currency,
  tax_amount,
  raw_payload,
  canonical_data,
  image_url,
  image_sha256,
  lat,
  lon,
  location_source,
  location_label,
  created_at,
  updated_at
FROM receipts
WHERE id = @id::bigint;

-- name: ListReceiptItemsForReceipt :many
SELECT
  id, receipt_id, line_no, name, qty, unit_price, line_total, sku, category_hint,
  created_at, updated_at
FROM receipt_items
WHERE receipt_id = @receipt_id::bigint
ORDER BY line_no NULLS LAST, id;

-- name: CreateReceipt :one
INSERT INTO receipts (
  engine, parse_status, link_status, match_ids,
  merchant, purchase_date, total_amount, currency, tax_amount,
  raw_payload, canonical_data, image_url, image_sha256,
  lat, lon, location_source, location_label
) VALUES (
  @engine::text,
  COALESCE(sqlc.narg('parse_status')::text, 'pending'),
  COALESCE(sqlc.narg('link_status')::text,  'unlinked'),
  sqlc.narg('match_ids')::bigint[],
  sqlc.narg('merchant')::text,
  sqlc.narg('purchase_date')::date,
  sqlc.narg('total_amount')::numeric,
  sqlc.narg('currency')::char(3),
  sqlc.narg('tax_amount')::numeric,
  sqlc.narg('raw_payload')::jsonb,
  sqlc.narg('canonical_data')::jsonb,
  sqlc.narg('image_url')::text,
  sqlc.narg('image_sha256')::bytea,
  sqlc.narg('lat')::double precision,
  sqlc.narg('lon')::double precision,
  sqlc.narg('location_source')::text,
  sqlc.narg('location_label')::text
)
RETURNING
  id,
  engine, parse_status, link_status, match_ids,
  merchant, purchase_date, total_amount, currency, tax_amount,
  raw_payload, canonical_data, image_url, image_sha256,
  lat, lon, location_source, location_label,
  created_at, updated_at;

-- name: InsertReceiptItem :exec
INSERT INTO receipt_items (
  receipt_id, line_no, name, qty, unit_price, line_total, sku, category_hint
) VALUES (
  @receipt_id::bigint,
  sqlc.narg('line_no')::int,
  @name::text,
  sqlc.narg('qty')::numeric,
  sqlc.narg('unit_price')::numeric,
  sqlc.narg('line_total')::numeric,
  sqlc.narg('sku')::text,
  sqlc.narg('category_hint')::text
);

-- name: UpdateReceiptPartial :execrows
UPDATE receipts
SET
  engine           = CASE WHEN @engine_set::bool           THEN @engine::text                 ELSE engine           END,
  parse_status     = CASE WHEN @parse_status_set::bool     THEN @parse_status::text          ELSE parse_status     END,
  link_status      = CASE WHEN @link_status_set::bool      THEN @link_status::text           ELSE link_status      END,
  match_ids        = CASE WHEN @match_ids_set::bool        THEN sqlc.narg('match_ids')::bigint[] ELSE match_ids   END,
  merchant         = CASE WHEN @merchant_set::bool         THEN sqlc.narg('merchant')::text  ELSE merchant         END,
  purchase_date    = CASE WHEN @purchase_date_set::bool    THEN sqlc.narg('purchase_date')::date ELSE purchase_date END,
  total_amount     = CASE WHEN @total_amount_set::bool     THEN sqlc.narg('total_amount')::numeric ELSE total_amount END,
  currency         = CASE WHEN @currency_set::bool         THEN sqlc.narg('currency')::char(3) ELSE currency       END,
  tax_amount       = CASE WHEN @tax_amount_set::bool       THEN sqlc.narg('tax_amount')::numeric ELSE tax_amount   END,
  raw_payload      = CASE WHEN @raw_payload_set::bool      THEN sqlc.narg('raw_payload')::jsonb ELSE raw_payload   END,
  canonical_data   = CASE WHEN @canonical_data_set::bool   THEN sqlc.narg('canonical_data')::jsonb ELSE canonical_data END,
  image_url        = CASE WHEN @image_url_set::bool        THEN sqlc.narg('image_url')::text ELSE image_url        END,
  image_sha256     = CASE WHEN @image_sha256_set::bool     THEN sqlc.narg('image_sha256')::bytea ELSE image_sha256 END,
  lat              = CASE WHEN @lat_set::bool              THEN sqlc.narg('lat')::double precision ELSE lat        END,
  lon              = CASE WHEN @lon_set::bool              THEN sqlc.narg('lon')::double precision ELSE lon        END,
  location_source  = CASE WHEN @location_source_set::bool  THEN sqlc.narg('location_source')::text ELSE location_source END,
  location_label   = CASE WHEN @location_label_set::bool   THEN sqlc.narg('location_label')::text ELSE location_label END
WHERE id = @id::bigint;

-- name: DeleteReceipt :execrows
DELETE FROM receipts WHERE id = @id::bigint;
