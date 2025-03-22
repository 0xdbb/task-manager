-- name: GetUser :one
SELECT * FROM "user"
WHERE id = $1;

-- name: ListUsers :many
SELECT id, first_name, last_name, email 
FROM "user"
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO "user" (
   first_name, last_name, email, password, address, phone,  date_of_birth
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateUser :one
UPDATE "user"
SET first_name = $2,
    last_name = $3,
    address = $4,
    date_of_birth = $5,
    phone = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;


-- name: DeleteUser :exec
DELETE FROM "user"
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user"
where email = $1;
