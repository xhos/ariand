-- name: ListCredentialsByUser :many
SELECT id, user_id, credential_id, public_key, sign_count, created_at
FROM user_credentials
WHERE user_id = @user_id::uuid
ORDER BY created_at DESC;

-- name: GetCredentialForUser :one
SELECT id, user_id, credential_id, public_key, sign_count, created_at
FROM user_credentials
WHERE id = @id::uuid AND user_id = @user_id::uuid;

-- name: GetCredentialByCredentialId :one
SELECT id, user_id, credential_id, public_key, sign_count, created_at
FROM user_credentials
WHERE credential_id = @credential_id::bytea;

-- name: CreateCredential :one
INSERT INTO user_credentials (user_id, credential_id, public_key, sign_count)
VALUES (@user_id::uuid, @credential_id::bytea, @public_key::bytea, @sign_count::bigint)
RETURNING id, user_id, credential_id, public_key, sign_count, created_at;

-- name: UpdateCredentialSignCount :execrows
UPDATE user_credentials
SET sign_count = @sign_count::bigint
WHERE id = @id::uuid AND user_id = @user_id::uuid;

-- name: UpdateCredentialSignCountByCredentialId :execrows
UPDATE user_credentials
SET sign_count = @sign_count::bigint
WHERE credential_id = @credential_id::bytea;

-- name: DeleteCredentialForUser :execrows
DELETE FROM user_credentials 
WHERE id = @id::uuid AND user_id = @user_id::uuid;

-- name: DeleteAllCredentialsForUser :execrows
DELETE FROM user_credentials 
WHERE user_id = @user_id::uuid;

-- name: CountCredentialsForUser :one
SELECT COUNT(*) AS credential_count
FROM user_credentials
WHERE user_id = @user_id::uuid;

-- name: CheckCredentialExists :one
SELECT EXISTS(
  SELECT 1 FROM user_credentials 
  WHERE credential_id = @credential_id::bytea
) AS exists;

-- Legacy admin queries (use sparingly)
-- name: GetCredential :one
SELECT id, user_id, credential_id, public_key, sign_count, created_at
FROM user_credentials
WHERE id = @id::uuid;