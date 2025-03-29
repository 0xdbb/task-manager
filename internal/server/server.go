package server

import (
	"fmt"
	"log"
	"net/http"
	"task-manager/config"
	db "task-manager/internal/database/sqlc"
	"task-manager/internal/queue"
	"task-manager/internal/token"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	engine *gin.Engine

	tokenMaker token.Maker

	queueManager *queue.QueueManager

	db *db.Service

	config *config.Config
}

func NewServer(config *config.Config, queueManager *queue.QueueManager) (*http.Server, error) {

	mode := gin.DebugMode

	dburl := config.DB_URL_DEV

	if config.PRODUCTION == "1" {
		dburl = config.DB_URL
		mode = gin.ReleaseMode

	}

	gin.SetMode(mode)

	// ----Create JWT Maker-----
	tokenMaker, err := token.NewJWTMaker(config.TOKEN_SECRET)

	if err != nil {
		return nil, fmt.Errorf("Error creating token maker %w", err)
	}

	// ----Create Database Service
	newService := db.NewService(dburl)

	// ----- NewServer -----
	NewServer := &Server{

		engine: gin.Default(),

		config: config,

		tokenMaker: tokenMaker,

		queueManager: queueManager,

		db: newService,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("StrongPassword", StrongPassword)
	}

	// -----Set CORS----
	NewServer.Cors()

	NewServer.RegisterRoutes()

	port := fmt.Sprintf(":%s", config.PORT)

	log.Printf("------Server spinning on Port %s-------\n", port)

	// Declare Server config
	server := &http.Server{
		Addr:         port,
		Handler:      NewServer.engine,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, nil
}
