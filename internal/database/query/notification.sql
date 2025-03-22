-- name: GetNotification :one
SELECT * FROM "notification"
WHERE id = $1;

-- name: ListNotificationsForUser :many
SELECT * FROM "notification"
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateNotification :one
INSERT INTO "notification" (user_id, message, sent_at)
VALUES ($1, $2, now())
RETURNING *;

-- name: DeleteNotification :exec
DELETE FROM "notification"
WHERE id = $1;
