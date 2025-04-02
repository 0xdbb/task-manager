package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"task-manager/internal/token"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Rate limiter store (per user)
var rateLimiters sync.Map

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"

	// Rate limit settings (10 requests per minute)
	requestsPerMinute = 10
	burstCapacity     = 10 // Same as limit for strict control
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
	}
}

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
		// Create a limiter that allows 10 requests per minute with a burst of 10
		rl := rate.NewLimiter(rate.Limit(requestsPerMinute)/60, burstCapacity)
		rateLimiters.Store(userID, rl)
		return rl
	}
	return limiter.(*rate.Limiter)
}

// RateLimitMiddleware should be placed BEFORE AuthMiddleware in your chain
func RateLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Try to extract user ID without full auth
		userID := "unauthorized" // Default for unauthenticated requests

		// Lightweight extraction of just the token (no verification yet)
		if authHeader := ctx.GetHeader(authorizationHeaderKey); authHeader != "" {
			fields := strings.Fields(authHeader)
			if len(fields) == 2 && strings.ToLower(fields[0]) == authorizationTypeBearer {
				if tokenMaker, exists := ctx.Get("tokenMaker"); exists {
					if payload, err := tokenMaker.(token.Maker).VerifyToken(fields[1]); err == nil {
						userID = payload.UserID.String()
					}
				}
			}
		}

		limiter := getRateLimiter(userID)
		if !limiter.Allow() {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": fmt.Sprintf("Limit: %d requests per minute", requestsPerMinute),
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
