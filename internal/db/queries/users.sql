-- name: ListUsers :many
SELECT id, email, display_name, created_at, updated_at
FROM users
ORDER BY created_at DESC;

-- name: GetUser :one
SELECT id, email, display_name, created_at, updated_at
FROM users
WHERE id = @id::uuid;

-- name: GetUserByEmail :one
SELECT id, email, display_name, created_at, updated_at
FROM users
WHERE lower(email) = lower(@email::text);

-- name: CreateUser :one
INSERT INTO users (email, display_name)
VALUES (@email::text, sqlc.narg('display_name')::text)
RETURNING id, email, display_name, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET email = COALESCE(sqlc.narg('email')::text, email),
    display_name = COALESCE(sqlc.narg('display_name')::text, display_name)
WHERE id = @id::uuid
RETURNING id, email, display_name, created_at, updated_at;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = @display_name::text
WHERE id = @id::uuid
RETURNING id, email, display_name, created_at, updated_at;

-- name: DeleteUser :execrows
DELETE FROM users WHERE id = @id::uuid;

-- name: CheckUserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = @id::uuid) AS exists;