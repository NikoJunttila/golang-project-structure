-- name: GetFoo :one
SELECT * FROM foos
WHERE id = ? LIMIT 1;

-- name: ListFoos :many
SELECT * FROM foos
ORDER BY message;

-- name: InsertFoo :many
INSERT INTO foos (
    message
) VALUES (?) RETURNING *;

-- name: UpdateFoo :one
UPDATE foos 
SET message = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? 
RETURNING *;