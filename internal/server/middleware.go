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
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Invalid Token"))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Invalid Token"))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Invalid Token"))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, HandleError(err, http.StatusUnauthorized, "Invalid Token"))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
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
		userID := ctx.GetString("user_id") // Assume user_id is extracted from JWT or session

		if userID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}

		limiter := getRateLimiter(userID)

		if !limiter.Allow() {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Try again later."})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
