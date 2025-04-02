package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-manager/internal/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	userID uuid.UUID,
	role string,
	duration time.Duration,
) {
	token, payload, err := tokenMaker.CreateToken(userID, role, duration)
	require.NotEmpty(t, payload)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	userID := uuid.New()
	role := "worker" // Example role

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, userID, role, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No authorization header set
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuthorizationType",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unsupported", userID, role, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				request.Header.Set(authorizationHeaderKey, "invalidformat")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, userID, role, -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			tokenMaker, err := token.NewJWTMaker("12345678901234567890123456789012")
			require.NoError(t, err)

			// Test endpoint
			router.GET(
				"/protected",
				AuthMiddleware(tokenMaker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"message": "authorized"})
				},
			)

			// Create request
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, "/protected", nil)
			require.NoError(t, err)

			// Set up auth
			tc.setupAuth(t, request, tokenMaker)

			// Serve request
			router.ServeHTTP(recorder, request)

			// Check response
			tc.checkResponse(t, recorder)
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	userID := uuid.New()
	role := "ADMIN"

	testCases := []struct {
		name          string
		prepare       func(t *testing.T, tokenMaker token.Maker) *http.Request
		requests      int // Number of requests to make
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "WithinRateLimit",
			prepare: func(t *testing.T, tokenMaker token.Maker) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/limited", nil)
				require.NoError(t, err)
				addAuthorization(t, req, tokenMaker, authorizationTypeBearer, userID, role, time.Minute)
				return req
			},
			requests: 1,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "ExceedRateLimit",
			prepare: func(t *testing.T, tokenMaker token.Maker) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/limited", nil)
				require.NoError(t, err)
				addAuthorization(t, req, tokenMaker, authorizationTypeBearer, userID, role, time.Minute)
				return req
			},
			requests: requestLimit + 1,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusTooManyRequests, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			prepare: func(t *testing.T, tokenMaker token.Maker) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/limited", nil)
				require.NoError(t, err)
				// No authorization header
				return req
			},
			requests: 1,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			tokenMaker, err := token.NewJWTMaker("12345678901234567890123456789012")
			require.NoError(t, err)

			// The key change is here - we need to set the tokenMaker in the context chain
			router.GET(
				"/limited",
				func(ctx *gin.Context) {
					ctx.Set("tokenMaker", tokenMaker)
					ctx.Next()
				},
				RateLimitMiddleware(),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
				},
			)

			// Make the specified number of requests
			for i := 0; i < tc.requests; i++ {
				recorder := httptest.NewRecorder()
				request := tc.prepare(t, tokenMaker)
				router.ServeHTTP(recorder, request)

				// Only check the last response
				if i == tc.requests-1 {
					tc.checkResponse(t, recorder)
				}
			}
		})
	}
}
func TestExtractTokenPayload(t *testing.T) {
	userID := uuid.New()
	role := "worker" // Example role

	testCases := []struct {
		name          string
		setupRequest  func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, payload *token.Payload, err error)
	}{
		{
			name: "ValidToken",
			setupRequest: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, userID, role, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.NoError(t, err)
				require.Equal(t, userID, payload.UserID)
				require.Equal(t, role, payload.Role)
			},
		},
		{
			name: "NoAuthorizationHeader",
			setupRequest: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No header set
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "authorization header is not provided")
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupRequest: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				request.Header.Set(authorizationHeaderKey, "invalidformat")
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid authorization header format")
			},
		},
		{
			name: "UnsupportedAuthorizationType",
			setupRequest: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unsupported", userID, role, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "unsupported authorization type")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test context
			ctx := &gin.Context{}
			request, err := http.NewRequest(http.MethodGet, "/test", nil)
			require.NoError(t, err)

			// Setup token maker and request
			tokenMaker, err := token.NewJWTMaker("12345678901234567890123456789012")
			require.NoError(t, err)
			tc.setupRequest(t, request, tokenMaker)
			ctx.Request = request
			ctx.Set("tokenMaker", tokenMaker)

			// Call the function
			payload, err := ExtractTokenPayload(ctx)

			// Check response
			tc.checkResponse(t, payload, err)
		})
	}
}

