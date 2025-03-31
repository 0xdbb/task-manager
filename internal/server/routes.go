package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @BasePath  /api/v1
func (s *Server) RegisterRoutes() {
	v1 := s.engine.Group("/api/v1")
	{
		s.SwaggerRoute(v1)
		s.TaskRoutes(v1)
		s.AdminUserRoutes(v1)
		s.AuthRoutes(v1)
	}
}

func (s *Server) AuthRoutes(v1 *gin.RouterGroup) {
	AuthRouteGroup := v1.Group("/auth")
	{
		AuthRouteGroup.POST("/login", s.Login)
		AuthRouteGroup.POST("/register", s.Register)
		AuthRouteGroup.POST("/tokens/renew_access", s.RenewAccessToken)
	}
}

func (s *Server) AdminUserRoutes(v1 *gin.RouterGroup) {
	AdminUserRouteGroup := v1.Group("/user").
		Use(AuthMiddleware(s.tokenMaker)).
		Use(RateLimitMiddleware())
	{
		AdminUserRouteGroup.GET("/", s.GetUsers)
		AdminUserRouteGroup.GET("/:id", s.GetUser)
		AdminUserRouteGroup.PATCH("/:id/role", s.UpdateUserRole)
		AdminUserRouteGroup.DELETE("/:id", s.DeleteUser)
	}
}

func (s *Server) TaskRoutes(v1 *gin.RouterGroup) {
	taskRouteGroup := v1.Group("/task").
		Use(AuthMiddleware(s.tokenMaker)).
		Use(RateLimitMiddleware())
	{
		taskRouteGroup.GET("/", s.GetTasks)
		taskRouteGroup.GET("/:id", s.GetTask)
		taskRouteGroup.POST("/", s.CreateTask)
		taskRouteGroup.PATCH("/:id", s.UpdateTaskStatus)
	}
}

func (s *Server) SwaggerRoute(v1 *gin.RouterGroup) {
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	v1.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusPermanentRedirect, "/api/v1/swagger/index.html")
	})
}

func (s *Server) Cors() {

	origins := strings.Split(s.config.ALLOWED_ORIGINS, ",")

	s.engine.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}
