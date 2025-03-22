// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: user.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO "user" (
   first_name, last_name, email, password, address, phone,  date_of_birth
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, first_name, last_name, email, password, phone, address, date_of_birth, created_at, updated_at
`

type CreateUserParams struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"date_of_birth"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.FirstName,
		arg.LastName,
		arg.Email,
		arg.Password,
		arg.Address,
		arg.Phone,
		arg.DateOfBirth,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.Password,
		&i.Phone,
		&i.Address,
		&i.DateOfBirth,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM "user"
WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const getUser = `-- name: GetUser :one
SELECT id, first_name, last_name, email, password, phone, address, date_of_birth, created_at, updated_at FROM "user"
WHERE id = $1
`

func (q *Queries) GetUser(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.Password,
		&i.Phone,
		&i.Address,
		&i.DateOfBirth,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, first_name, last_name, email, password, phone, address, date_of_birth, created_at, updated_at FROM "user"
where email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.Password,
		&i.Phone,
		&i.Address,
		&i.DateOfBirth,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, first_name, last_name, email 
FROM "user"
ORDER BY id
LIMIT $1 OFFSET $2
`

type ListUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListUsersRow struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]ListUsersRow, error) {
	rows, err := q.db.Query(ctx, listUsers, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUsersRow{}
	for rows.Next() {
		var i ListUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
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

const updateUser = `-- name: UpdateUser :one
UPDATE "user"
SET first_name = $2,
    last_name = $3,
    address = $4,
    date_of_birth = $5,
    phone = $6,
    updated_at = now()
WHERE id = $1
RETURNING id, first_name, last_name, email, password, phone, address, date_of_birth, created_at, updated_at
`

type UpdateUserParams struct {
	ID          uuid.UUID `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Address     string    `json:"address"`
	DateOfBirth string    `json:"date_of_birth"`
	Phone       string    `json:"phone"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.ID,
		arg.FirstName,
		arg.LastName,
		arg.Address,
		arg.DateOfBirth,
		arg.Phone,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.Password,
		&i.Phone,
		&i.Address,
		&i.DateOfBirth,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
