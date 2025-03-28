package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "task-manager/internal/database/sqlc"
	"task-manager/util"
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

type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name      string    `json:"name" binding:"required" example:"John"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-02T12:00:00Z"`
}

func newUserResponse(user db.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// @BasePath /api/v1

// @Summary		Get Users
// @Description	Get a list of users
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			request	body		UsersRequest	true	"User Request"
// @Success		200		{array}		UserResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/user [get]
func (h *Server) GetUsers(ctx *gin.Context) {
	var req UsersRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request - perhaps page_id and page_size are missing in body"))
		return
	}

	arg := db.ListUsersParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	users, err := h.db.ListUsers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// @Summary		Get User
// @Description	Get user by ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id	path		uuid.UUID	true	"User ID"
// @Success		200	{object}	UserResponse
// @Failure		400	{object}	ErrorResponse
// @Failure		404	{object}	ErrorResponse
// @Failure		500	{object}	ErrorResponse
// @Router			/user/{id} [get]
func (h *Server) GetUser(ctx *gin.Context) {
	var userReq UserRequest

	if err := ctx.ShouldBindUri(&userReq); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	user, err := h.db.GetUser(ctx, userReq.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, HandleError(err, http.StatusNotFound, "User does not exist"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error retrieving user"))
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user))
}

// @Summary		Register User
// @Description	Register a new user
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			request	body		CreateUserRequest	true	"Create User Request"
// @Success		200		{object}	Message
// @Failure		400		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/user/register [post]
func (h *Server) Register(ctx *gin.Context) {
	var user CreateUserRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	passwordHash, err := util.HashPassword(user.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error hashing password"))
		return
	}

	userArg := db.CreateUserParams{
		Name:     user.Name,
		Email:    user.Email,
		Password: passwordHash,
	}

	_, err = h.db.CreateUser(ctx, userArg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, HandleError(err, http.StatusForbidden, "User with specified email already exists"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating user"))
		return
	}
	ctx.JSON(http.StatusOK, HandleMessage("User created successfully"))
}

// @Summary		Login User
// @Description	Login user with email and password
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			request	body		UserLoginRequest	true	"User Login Request"
// @Success		200		{object}	UserLoginResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		404		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/user/login [post]
func (h *Server) Login(ctx *gin.Context) {
	var userLoginReq UserLoginRequest

	if err := ctx.ShouldBindJSON(&userLoginReq); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	user, err := h.db.GetUserByEmail(ctx, userLoginReq.Email)

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, HandleError(err, http.StatusNotFound, "Invalid email or password"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error retrieving user"))
		return
	}

	err = util.VerifyPassword(user.Password, userLoginReq.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, HandleError(err, http.StatusNotFound, "Invalid email or password"))
		return
	}

	accessToken, accessPayload, err := h.tokenMaker.CreateToken(user.ID, h.config.ACCESS_TOKEN_DURATION)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating access token"))
		return
	}
	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(
		user.ID,
		h.config.REFRESH_TOKEN_DURATION,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating refresh token"))
		return
	}

	session, err := h.db.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating session"))
		return
	}

	rsp := UserLoginResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, rsp)
}
