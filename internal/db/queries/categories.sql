-- name: ListCategories :many
SELECT id, slug, label, color, created_at, updated_at
FROM categories
ORDER BY slug;

-- name: GetCategory :one
SELECT id, slug, label, color, created_at, updated_at
FROM categories
WHERE id = @id::bigint;

-- name: GetCategoryBySlug :one
SELECT id, slug, label, color, created_at, updated_at
FROM categories
WHERE slug = @slug::text;

-- name: ListCategorySlugs :many
SELECT slug
FROM categories
ORDER BY slug;

-- name: CreateCategory :one
INSERT INTO categories (slug, label, color)
VALUES (@slug::text, @label::text, @color::text)
RETURNING id, slug, label, color, created_at, updated_at;

-- name: DeleteCategory :execrows
DELETE FROM categories
WHERE id = @id::bigint;

-- name: UpdateCategoryPartial :one
UPDATE categories
SET
  slug  = CASE WHEN @slug_set::bool  THEN @slug::text  ELSE slug  END,
  label = CASE WHEN @label_set::bool THEN @label::text ELSE label END,
  color = CASE WHEN @color_set::bool THEN @color::text ELSE color END
WHERE id = @id::bigint
RETURNING id, slug, label, color, created_at, updated_at;
