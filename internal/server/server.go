package server

import (
	"fmt"
	"log"
	"net/http"
	"task-manager/config"
	db "task-manager/internal/database/sqlc"
	"task-manager/internal/server/handler"
	"task-manager/internal/server/token"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	engine *gin.Engine

	tokenMaker token.Maker

	config *config.Config

	handlers *handler.Handler
}

func (s *Server) setHandlers(service *db.Service, config *config.Config, tokenMaker token.Maker) {

	serveHandler := handler.NewHandler(service, config, tokenMaker)

	s.handlers = serveHandler
}

func NewServer(config *config.Config) (*http.Server, error) {

	mode := gin.DebugMode

	dburl := config.DB_URL_DEV

	if config.PRODUCTION == "1" {
		dburl = config.DB_URL
		mode = gin.ReleaseMode

	}

	gin.SetMode(mode)

	tokenMaker, err := token.NewJWTMaker(config.TOKEN_SECRET)

	if err != nil {
		return nil, fmt.Errorf("Error creating token maker %w", err)
	}

	NewServer := &Server{

		engine: gin.Default(),

		config: config,

		tokenMaker: tokenMaker,
	}

	newService := db.NewService(dburl)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("StrongPassword", handler.StrongPassword)
	}

	// NewServer.engine.Use(helmet.Default())
	NewServer.setHandlers(newService, NewServer.config, NewServer.tokenMaker)

	NewServer.Cors()

	NewServer.RegisterRoutes()

	port := fmt.Sprintf(":%s", config.PORT)

	log.Printf("Server spinning on Port %s.....\n", port)

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
