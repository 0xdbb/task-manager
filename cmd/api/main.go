// @title           task-manager API
// @version         1.0
// @description     API documentation for CheapStores Service
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"task-manager/config"
	_ "task-manager/docs/swagger" // Import generated Swagger docs
	"task-manager/internal/queue"
	"task-manager/internal/server"
	"time"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("-----Server exiting-----")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("config error: %s", err))
	}

	newQueue, err := queue.NewQueueManager(config.RMQ_ADDRESS)
	if err != nil {
		log.Fatal(fmt.Sprintf("config error: %s", err))
	}

	server, err := server.NewServer(config, newQueue)
	if err != nil {
		panic(fmt.Sprintf("config error: %s", err))
	}

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
