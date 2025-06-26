-- name: CreateUser :one
INSERT INTO users (
  id,
  lookup_id,
  email,
  password_hash,
  name,
  avatar_url,
  provider,
  provider_id,
  email_verified,
  created_at,
  updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ?;

-- name: GetUserBylookupID :one
SELECT * FROM users
WHERE lookup_id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: GetUserByProviderID :one
SELECT * FROM users
WHERE provider = ? AND provider_id = ?;

-- name: UpdateUser :one
UPDATE users
SET
  email = ?,
  password_hash = ?,
  name = ?,
  avatar_url = ?,
  provider = ?,
  provider_id = ?,
  email_verified = ?,
  created_at = ?,
  updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;

-- name: UpdateUserSecret :exec
UPDATE users SET secret = ? WHERE id = ?;
