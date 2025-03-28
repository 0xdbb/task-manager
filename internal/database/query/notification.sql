-- name: ListUserNotifications :many
SELECT * FROM "notification"
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateNotification :exec
INSERT INTO "notification" (
  user_id, task_id, message
) VALUES (
  $1, $2, $3
);

-- name: MarkNotificationAsRead :exec
UPDATE "notification"
SET read = true
WHERE id = $1;

-- name: DeleteNotification :exec
DELETE FROM "notification"
WHERE id = $1;
