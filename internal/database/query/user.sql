-- name: GetUser :one
SELECT * FROM "user"
WHERE id = $1;

-- name: ListUsers :many
SELECT id,  email 
FROM "user"
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO "user" (
     name,email, password 
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateUser :one
UPDATE "user"
SET name = $2,
    type = $3,
    phone = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;


-- name: DeleteUser :exec
DELETE FROM "user"
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user"
where email = $1;
