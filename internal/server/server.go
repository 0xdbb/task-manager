package server

import (
	"fmt"
	"log"
	"net/http"
	"task-manager/internal/config"
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

	dburl := config.DbUrlDev

	if config.Production == "1" {
		dburl = config.DbUrl
		mode = gin.ReleaseMode

	}

	gin.SetMode(mode)

	// ----Create JWT Maker-----
	tokenMaker, err := token.NewJWTMaker(config.TokenSecret)

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

	port := fmt.Sprintf(":%s", config.Port)

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
