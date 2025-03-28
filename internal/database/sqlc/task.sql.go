// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: task.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createTask = `-- name: CreateTask :one
INSERT INTO "task" (
  user_id, type, payload, status, due_time
) VALUES (
  $1, $2, $3, 'pending', $4
)
RETURNING id, user_id, type, payload, status, result, due_time, created_at, updated_at
`

type CreateTaskParams struct {
	UserID  uuid.UUID `json:"user_id"`
	Type    string    `json:"type"`
	Payload string    `json:"payload"`
	DueTime time.Time `json:"due_time"`
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error) {
	row := q.db.QueryRow(ctx, createTask,
		arg.UserID,
		arg.Type,
		arg.Payload,
		arg.DueTime,
	)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Type,
		&i.Payload,
		&i.Status,
		&i.Result,
		&i.DueTime,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createTaskLog = `-- name: CreateTaskLog :exec
INSERT INTO "task_log" (
  task_id, worker_id, status, message
) VALUES (
  $1, $2, $3, $4
)
`

type CreateTaskLogParams struct {
	TaskID   uuid.UUID   `json:"task_id"`
	WorkerID pgtype.Text `json:"worker_id"`
	Status   string      `json:"status"`
	Message  pgtype.Text `json:"message"`
}

func (q *Queries) CreateTaskLog(ctx context.Context, arg CreateTaskLogParams) error {
	_, err := q.db.Exec(ctx, createTaskLog,
		arg.TaskID,
		arg.WorkerID,
		arg.Status,
		arg.Message,
	)
	return err
}

const deleteTask = `-- name: DeleteTask :exec
DELETE FROM "task"
WHERE id = $1
`

func (q *Queries) DeleteTask(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteTask, id)
	return err
}

const getTask = `-- name: GetTask :one
SELECT id, user_id, type, payload, status, result, due_time, created_at, updated_at FROM "task"
WHERE id = $1
`

func (q *Queries) GetTask(ctx context.Context, id uuid.UUID) (Task, error) {
	row := q.db.QueryRow(ctx, getTask, id)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Type,
		&i.Payload,
		&i.Status,
		&i.Result,
		&i.DueTime,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getTaskLog = `-- name: GetTaskLog :many

SELECT id, task_id, worker_id, status, message, created_at FROM "task_log"
WHERE task_id = $1
ORDER BY created_at DESC
`

// TASK LOGS
func (q *Queries) GetTaskLog(ctx context.Context, taskID uuid.UUID) ([]TaskLog, error) {
	rows, err := q.db.Query(ctx, getTaskLog, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TaskLog{}
	for rows.Next() {
		var i TaskLog
		if err := rows.Scan(
			&i.ID,
			&i.TaskID,
			&i.WorkerID,
			&i.Status,
			&i.Message,
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

const listTasksByUser = `-- name: ListTasksByUser :many
SELECT id, user_id, type, payload, status, result, due_time, created_at, updated_at FROM "task"
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

type ListTasksByUserParams struct {
	UserID uuid.UUID `json:"user_id"`
	Limit  int32     `json:"limit"`
	Offset int32     `json:"offset"`
}

func (q *Queries) ListTasksByUser(ctx context.Context, arg ListTasksByUserParams) ([]Task, error) {
	rows, err := q.db.Query(ctx, listTasksByUser, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Task{}
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Type,
			&i.Payload,
			&i.Status,
			&i.Result,
			&i.DueTime,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const updateTaskStatus = `-- name: UpdateTaskStatus :one
UPDATE "task"
SET status = $2,
    result = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, user_id, type, payload, status, result, due_time, created_at, updated_at
`

type UpdateTaskStatusParams struct {
	ID     uuid.UUID   `json:"id"`
	Status pgtype.Text `json:"status"`
	Result pgtype.Text `json:"result"`
}

func (q *Queries) UpdateTaskStatus(ctx context.Context, arg UpdateTaskStatusParams) (Task, error) {
	row := q.db.QueryRow(ctx, updateTaskStatus, arg.ID, arg.Status, arg.Result)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Type,
		&i.Payload,
		&i.Status,
		&i.Result,
		&i.DueTime,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
