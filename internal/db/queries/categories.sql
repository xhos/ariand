-- name: ListCategories :many
SELECT id, slug, label, color, created_at, updated_at
FROM categories
ORDER BY slug;

-- name: ListCategoriesWithUsage :many
SELECT 
  c.id, c.slug, c.label, c.color, c.created_at, c.updated_at,
  COUNT(t.id) AS usage_count,
  COALESCE(SUM(t.tx_amount), 0) AS total_amount
FROM categories c
LEFT JOIN transactions t ON c.id = t.category_id
GROUP BY c.id, c.slug, c.label, c.color, c.created_at, c.updated_at
ORDER BY usage_count DESC, c.slug;

-- name: ListCategoriesForUser :many
SELECT DISTINCT
  c.id, c.slug, c.label, c.color, c.created_at, c.updated_at,
  COUNT(t.id) AS user_usage_count,
  COALESCE(SUM(t.tx_amount), 0) AS user_total_amount
FROM categories c
LEFT JOIN transactions t ON c.id = t.category_id
LEFT JOIN accounts a ON t.account_id = a.id
LEFT JOIN account_users au ON a.id = au.account_id AND au.user_id = @user_id::uuid
WHERE t.id IS NULL OR (a.owner_id = @user_id::uuid OR au.user_id IS NOT NULL)
GROUP BY c.id, c.slug, c.label, c.color, c.created_at, c.updated_at
ORDER BY user_usage_count DESC, c.slug;

-- name: GetCategory :one
SELECT id, slug, label, color, created_at, updated_at
FROM categories
WHERE id = @id::bigint;

-- name: GetCategoryBySlug :one
SELECT id, slug, label, color, created_at, updated_at
FROM categories
WHERE slug = @slug::text;

-- name: GetCategoryWithStats :one
SELECT 
  c.id, c.slug, c.label, c.color, c.created_at, c.updated_at,
  COUNT(t.id) AS usage_count,
  COALESCE(SUM(t.tx_amount), 0) AS total_amount,
  COALESCE(AVG(t.tx_amount), 0) AS avg_amount,
  MIN(t.tx_date) AS first_used,
  MAX(t.tx_date) AS last_used
FROM categories c
LEFT JOIN transactions t ON c.id = t.category_id
WHERE c.id = @id::bigint
GROUP BY c.id, c.slug, c.label, c.color, c.created_at, c.updated_at;

-- name: ListCategorySlugs :many
SELECT slug
FROM categories
ORDER BY slug;

-- name: CreateCategory :one
INSERT INTO categories (slug, label, color)
VALUES (@slug::text, @label::text, @color::text)
RETURNING id, slug, label, color, created_at, updated_at;

-- name: UpdateCategory :one
UPDATE categories
SET slug = COALESCE(sqlc.narg('slug')::text, slug),
    label = COALESCE(sqlc.narg('label')::text, label),
    color = COALESCE(sqlc.narg('color')::text, color)
WHERE id = @id::bigint
RETURNING id, slug, label, color, created_at, updated_at;

-- name: DeleteCategory :execrows
DELETE FROM categories
WHERE id = @id::bigint;

-- name: DeleteUnusedCategories :execrows
DELETE FROM categories
WHERE id NOT IN (
  SELECT DISTINCT category_id 
  FROM transactions 
  WHERE category_id IS NOT NULL
);

-- name: SearchCategories :many
SELECT id, slug, label, color, created_at, updated_at
FROM categories
WHERE slug ILIKE ('%' || @query::text || '%') 
   OR label ILIKE ('%' || @query::text || '%')
ORDER BY 
  CASE WHEN slug ILIKE (@query::text || '%') THEN 1 ELSE 2 END,
  slug;

-- name: GetMostUsedCategoriesForUser :many
SELECT 
  c.id, c.slug, c.label, c.color,
  COUNT(t.id) AS usage_count,
  SUM(t.tx_amount) AS total_amount
FROM categories c
JOIN transactions t ON c.id = t.category_id
JOIN accounts a ON t.account_id = a.id
LEFT JOIN account_users au ON a.id = au.account_id AND au.user_id = @user_id::uuid
WHERE (a.owner_id = @user_id::uuid OR au.user_id IS NOT NULL)
  AND (sqlc.narg('start')::timestamptz IS NULL OR t.tx_date >= sqlc.narg('start')::timestamptz)
  AND (sqlc.narg('end')::timestamptz IS NULL OR t.tx_date <= sqlc.narg('end')::timestamptz)
GROUP BY c.id, c.slug, c.label, c.color
ORDER BY usage_count DESC
LIMIT COALESCE(sqlc.narg('limit')::int, 10);

-- name: GetUnusedCategories :many
SELECT c.id, c.slug, c.label, c.color, c.created_at, c.updated_at
FROM categories c
LEFT JOIN transactions t ON c.id = t.category_id
WHERE t.id IS NULL
ORDER BY c.created_at DESC;

-- name: BulkCreateCategories :copyfrom
INSERT INTO categories (slug, label, color) 
VALUES (@slug, @label, @color);