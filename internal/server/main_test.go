package server

import (
	"net/http"
	"os"
	"task-manager/internal/config"
	db "task-manager/internal/database/sqlc"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Service) *http.Server {

	config, err := config.LoadConfig()
	require.NoError(t, err)

	server, err := NewServer(config, nil)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
