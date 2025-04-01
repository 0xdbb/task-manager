package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"task-manager/internal/token"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Rate limiter store (per user)
var rateLimiters sync.Map

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"

	// Rate limit settings (5 requests per minute per user)
	requestLimit  = 5
	timeInterval  = time.Minute
	burstCapacity = 2 // Allow small bursts
)

// AuthMiddleware creates a gin middleware for authorization
func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
return func(ctx *gin.Context) {
		ctx.Set("tokenMaker", tokenMaker) // Store the token maker in the context

		payload, err := ExtractTokenPayload(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Invalid Token"))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}}


// ExtractTokenPayload extracts the token payload from the request context
func ExtractTokenPayload(ctx *gin.Context) (*token.Payload, error) {
	authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

	if len(authorizationHeader) == 0 {
		return nil, errors.New("authorization header is not provided")
	}

	fields := strings.Fields(authorizationHeader)
	if len(fields) < 2 {
		return nil, errors.New("invalid authorization header format")
	}

	authorizationType := strings.ToLower(fields[0])
	if authorizationType != authorizationTypeBearer {
		return nil, fmt.Errorf("unsupported authorization type %s", authorizationType)
	}

	accessToken := fields[1]
	payload, err := ctx.MustGet("tokenMaker").(token.Maker).VerifyToken(accessToken)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	return payload, nil
}

// getRateLimiter retrieves or creates a rate limiter for a given user
func getRateLimiter(userID string) *rate.Limiter {
	limiter, exists := rateLimiters.Load(userID)
	if !exists {
		rl := rate.NewLimiter(rate.Every(timeInterval/time.Duration(requestLimit)), burstCapacity)
		rateLimiters.Store(userID, rl)
		return rl
	}
	return limiter.(*rate.Limiter)
}

// Middleware for rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload, err := ExtractTokenPayload(ctx)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Unauthorized"))
			ctx.Abort()
			return
		}

		limiter := getRateLimiter(payload.UserID.String()) // Using extracted user ID

		if !limiter.Allow() {
			ctx.JSON(http.StatusTooManyRequests, HandleError(nil, http.StatusTooManyRequests, "Too many requests. Try again later."))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

