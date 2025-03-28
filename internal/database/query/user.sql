-- name: GetUser :one
SELECT * FROM "user"
WHERE id = $1;

-- name: ListUsers :many
SELECT id,  email 
FROM "user"
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO "user" (
    name, email, password, role
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateUserRole :one
UPDATE "user"
SET role = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;


-- name: DeleteUser :exec
DELETE FROM "user"
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user"
where email = $1;
