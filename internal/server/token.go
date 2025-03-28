package server

import (
	"errors"
	"fmt"
	"net/http"
	db "task-manager/internal/database/sqlc"
	"time"

	"github.com/gin-gonic/gin"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

// @Summary      Renew Access Token
// @Description  Generates a new access token using a valid refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body renewAccessTokenRequest true "Refresh Token Request"
// @Success      200  {object}  renewAccessTokenResponse
// @Failure      400  {object}  ErrorResponse "Invalid request"
// @Failure      401  {object}  ErrorResponse "Unauthorized or Invalid token"
// @Failure      404  {object}  ErrorResponse "Session not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /auth/renew [post]
func (h *Server) RenewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	refreshPayload, err := h.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Invalid token"))
		return
	}

	session, err := h.db.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, HandleError(err, http.StatusNotFound, "Session not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error retrieving session"))
		return
	}

	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		ctx.JSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Blocked session"))
		return
	}

	if session.UserID != refreshPayload.UserID {
		err := fmt.Errorf("incorrect session user")
		ctx.JSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Invalid session user"))
		return
	}

	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		ctx.JSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Invalid session token"))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		ctx.JSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Session expired"))
		return
	}

	accessToken, accessPayload, err := h.tokenMaker.CreateToken(
		refreshPayload.UserID,
		refreshPayload.Role,
		h.config.ACCESS_TOKEN_DURATION,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating access token"))
		return
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}
