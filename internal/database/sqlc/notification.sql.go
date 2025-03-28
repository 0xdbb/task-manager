// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: notification.sql

package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createNotification = `-- name: CreateNotification :exec
INSERT INTO "notification" (
  user_id, task_id, message
) VALUES (
  $1, $2, $3
)
`

type CreateNotificationParams struct {
	UserID  pgtype.UUID `json:"user_id"`
	TaskID  pgtype.UUID `json:"task_id"`
	Message pgtype.Text `json:"message"`
}

func (q *Queries) CreateNotification(ctx context.Context, arg CreateNotificationParams) error {
	_, err := q.db.Exec(ctx, createNotification, arg.UserID, arg.TaskID, arg.Message)
	return err
}

const deleteNotification = `-- name: DeleteNotification :exec
DELETE FROM "notification"
WHERE id = $1
`

func (q *Queries) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteNotification, id)
	return err
}

const listUserNotifications = `-- name: ListUserNotifications :many
SELECT id, user_id, task_id, message, sent, created_at FROM "notification"
WHERE user_id = $1
ORDER BY created_at DESC
`

func (q *Queries) ListUserNotifications(ctx context.Context, userID pgtype.UUID) ([]Notification, error) {
	rows, err := q.db.Query(ctx, listUserNotifications, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Notification{}
	for rows.Next() {
		var i Notification
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.TaskID,
			&i.Message,
			&i.Sent,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const markNotificationAsRead = `-- name: MarkNotificationAsRead :exec
UPDATE "notification"
SET read = true
WHERE id = $1
`

func (q *Queries) MarkNotificationAsRead(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, markNotificationAsRead, id)
	return err
}
