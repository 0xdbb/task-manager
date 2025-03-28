-- name: GetTask :one
SELECT * FROM "task"
WHERE id = $1;

-- name: ListTasksByUser :many
SELECT * FROM "task"
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTask :one
INSERT INTO "task" (
  user_id, type, payload, status
) VALUES (
  $1, $2, $3, 'pending'
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


-- TASK LOGS

-- name: GetTaskLog :many
SELECT * FROM "task_log"
WHERE task_id = $1
ORDER BY created_at DESC;

-- name: CreateTaskLog :exec
INSERT INTO "task_log" (
  task_id, worker_id, status, message
) VALUES (
  $1, $2, $3, $4
);
