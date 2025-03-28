package server

import (
	"time"

	"github.com/google/uuid"
)

// Users types
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required" example:"John"`
	Password string `json:"password" binding:"required,min=6,StrongPassword" example:"password123"`
	Email    string `json:"email" binding:"required" example:"john.doe@example.com"`
}

type UserRequest struct {
	ID uuid.UUID `uri:"id" binding:"min=0" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type UsersRequest struct {
	PageSize int32 `query:"page_size" binding:"required,min=1" example:"10"`
	PageID   int32 `query:"page_id" binding:"required,min=1" example:"1"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type UserLoginResponse struct {
	SessionID             uuid.UUID    `json:"session_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	AccessToken           string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at" example:"2025-02-05T13:15:08Z"`
	RefreshToken          string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at" example:"2025-02-06T13:15:08Z"`
	User                  UserResponse `json:"user"`
}

// User Response
type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name      string    `json:"name" binding:"required" example:"John"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Role      string    `json:"role" example:"ADMIN"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-02T12:00:00Z"`
}
