-- name: GetTask :one
SELECT * FROM "task"
WHERE id = $1;

-- name: ListAllTasks :many
SELECT * FROM "task"
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListTasksByUser :many
SELECT * FROM "task"
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTask :one
INSERT INTO "task" (
  user_id, title, description, result, type, payload, status, due_time, priority
) VALUES (
  $1, $2, $3, $4, $5, $6, 'PENDING', $7, $8
)
RETURNING *;

-- name: UpdateTaskStatus :one
UPDATE "task"
SET status = $2,
    result = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM "task"
WHERE id = $1;
