package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	db "task-manager/internal/database/sqlc"
	"task-manager/internal/token"
	"task-manager/util"
)

var ADMIN = "ADMIN"

// @BasePath /api/v1

// @Summary		Register User
// @Description	Register a new user
// @Tags		auth
// @Accept			json
// @Produce		json
// @Param			request	body		CreateUserRequest	true	"Create User Request"
// @Success		200		{object}	Message
// @Failure 400 {object} ErrorResponse "Bad request due to invalid query params"
// @Failure 401 {object} ErrorResponse "Unauthorized access"
// @Failure 403 {object} ErrorResponse "Forbidden: Admins only"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router			/auth/register [post]
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
		Role:     db.UserRole(user.Role),
	}

	_, err = h.db.CreateUser(ctx, userArg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, HandleError(err, http.StatusForbidden, "Email already registered"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating user"))
		return
	}
	ctx.JSON(http.StatusOK, HandleMessage("User created successfully"))
}

// @Summary		Login User
// @Description	Login user with email and password
// @Tags		auth
// @Accept			json
// @Produce		json
// @Param			request	body		UserLoginRequest	true	"User Login Request"
// @Success		200		{object}	UserLoginResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		404		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/auth/login [post]
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

	accessToken, accessPayload, err := h.tokenMaker.CreateToken(user.ID, string(user.Role), h.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating access token"))
		return
	}
	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(
		user.ID,
		string(user.Role),
		h.config.RefreshTokenDuration,
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
		ExpiresAt:    refreshPayload.ExpireAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating session"))
		return
	}

	rsp := UserLoginResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpireAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpireAt,
		User:                  newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, rsp)
}

func isAdmin(ctx *gin.Context, strict bool) bool {
	authPayload, exists := ctx.Get(authorizationPayloadKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, HandleError(nil, http.StatusUnauthorized, "Unauthorized: Missing auth payload"))
		ctx.Abort()
		return false
	}

	role, ok := authPayload.(*token.Payload)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, HandleError(nil, http.StatusUnauthorized, "Unauthorized: Invalid token payload"))
		ctx.Abort()
		return false
	}

	// Direct role comparison instead of map lookup
	switch db.UserRole(role.Role) {
	case db.UserRoleADMIN:
		return true
	case db.UserRoleSTANDARD:
		if strict {
			break // If strict, STANDARD is not allowed
		}
		return true
	}

	ctx.JSON(http.StatusForbidden, HandleError(nil, http.StatusForbidden, "Forbidden: Unauthorized"))
	ctx.Abort()
	return false
}


