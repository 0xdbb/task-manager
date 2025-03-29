package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "task-manager/internal/database/sqlc"
)

// @BasePath /api/v1

// @Summary		Get Users
// @Description	Get a list of users
// @Tags			users
// @Accept			json
// @Produce		json
// @Security BearerAuth
// @Param			page_size	query		int	false	"Number of users per page"
// @Param			page_id		query		int	false	"Page number"
// @Success		200		{array}		UserResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/user [get]
func (h *Server) GetUsers(ctx *gin.Context) {
	if !isAdmin(ctx) {
		return
	}

	var req UsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
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
// @Security BearerAuth
// @Param id path string true "User ID (UUID format)"
// @Success		200	{object}	UserResponse
// @Failure		400	{object}	ErrorResponse
// @Failure		404	{object}	ErrorResponse
// @Failure		500	{object}	ErrorResponse
// @Router			/user/{id} [get]
func (h *Server) GetUser(ctx *gin.Context) {
	if !isAdmin(ctx) {
		return
	}

	id := ctx.Param("id")

	userID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid ID"))
		return
	}

	user, err := h.db.GetUser(ctx, userID)
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

// @Summary  Update User Role
// @Tags     users
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "User ID (UUID format)"
// @Param    role body     UpdateUserRoleRequest true  "New Role"
// @Success  200  {object}  Message
// @Failure  400  {object}  ErrorResponse
// @Failure  404  {object}  ErrorResponse
// @Failure  500  {object}  ErrorResponse
// @Router   /user/{id}/role [patch]
func (h *Server) UpdateUserRole(ctx *gin.Context) {
	if !isAdmin(ctx) {
		return
	}
	var req UpdateUserRoleRequest

	id := ctx.Param("id")

	userID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid ID"))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request payload"))
		return
	}

	// Update User Role
	_, err = h.db.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:   userID,
		Role: db.UserRole(req.Role),
	})

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, HandleError(err, http.StatusNotFound, "User not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error updating role"))
		return
	}

	ctx.JSON(http.StatusOK, HandleMessage("User role updated successfully"))
}

// @Summary  Delete User
// @Tags     users
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "User ID (UUID format)"
// @Success  200  {object}  Message
// @Failure  400  {object}  ErrorResponse
// @Failure  404  {object}  ErrorResponse
// @Failure  500  {object}  ErrorResponse
// @Router   /user/{id} [delete]
func (h *Server) DeleteUser(ctx *gin.Context) {
	if !isAdmin(ctx) {
		return
	}
	id := ctx.Param("id")

	userID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid ID"))
		return
	}

	err = h.db.DeleteUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, HandleError(err, http.StatusNotFound, "User not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error deleting user"))
		return
	}

	ctx.JSON(http.StatusOK, HandleMessage("User deleted successfully"))
}

// Convert DB User to API Response
func newUserResponse(user db.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
