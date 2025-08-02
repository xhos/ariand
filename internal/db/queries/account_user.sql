-- name: AddAccountCollaborator :one
INSERT INTO account_users (account_id, user_id)
SELECT @account_id::bigint, @collaborator_user_id::uuid
FROM accounts a
WHERE a.id = @account_id::bigint 
  AND a.owner_id = @owner_user_id::uuid  -- only owners can add collaborators
  AND @collaborator_user_id::uuid != @owner_user_id::uuid  -- can't add yourself
ON CONFLICT DO NOTHING
RETURNING account_id, user_id, added_at;

-- name: RemoveAccountCollaborator :execrows
DELETE FROM account_users
WHERE account_id = @account_id::bigint
  AND user_id = @collaborator_user_id::uuid
  AND EXISTS (
    SELECT 1 FROM accounts a 
    WHERE a.id = @account_id::bigint 
      AND a.owner_id = @owner_user_id::uuid
  );

-- name: ListAccountCollaborators :many
SELECT u.id, u.email, u.display_name, au.added_at
FROM account_users au
JOIN users u ON u.id = au.user_id
JOIN accounts a ON a.id = au.account_id
WHERE au.account_id = @account_id::bigint
  AND (a.owner_id = @requesting_user_id::uuid OR au.user_id = @requesting_user_id::uuid)
ORDER BY u.email;

-- name: ListUserCollaborations :many
SELECT 
  a.id AS account_id, 
  a.name AS account_name, 
  a.bank, 
  au.added_at,
  u.email AS owner_email,
  u.display_name AS owner_name
FROM account_users au
JOIN accounts a ON au.account_id = a.id
JOIN users u ON a.owner_id = u.id
WHERE au.user_id = @user_id::uuid
ORDER BY au.added_at DESC;

-- name: CheckAccountCollaborator :one
SELECT EXISTS(
  SELECT 1 FROM account_users au
  JOIN accounts a ON au.account_id = a.id
  WHERE au.account_id = @account_id::bigint
    AND au.user_id = @user_id::uuid
    OR a.owner_id = @user_id::uuid
) AS is_collaborator;

-- name: GetAccountCollaboratorCount :one
SELECT COUNT(*) AS collaborator_count
FROM account_users
WHERE account_id = @account_id::bigint;

-- name: RemoveUserFromAllAccounts :execrows
DELETE FROM account_users
WHERE user_id = @user_id::uuid;

-- name: TransferAccountOwnership :execrows
UPDATE accounts
SET owner_id = @new_owner_id::uuid
WHERE id = @account_id::bigint 
  AND owner_id = @current_owner_id::uuid
  AND EXISTS (
    SELECT 1 FROM account_users 
    WHERE account_id = @account_id::bigint 
      AND user_id = @new_owner_id::uuid
  );

-- name: LeaveAccountCollaboration :execrows
DELETE FROM account_users
WHERE account_id = @account_id::bigint
  AND user_id = @user_id::uuid;