package router

import (
	"net/http"
	"time"

	"taskflow-pro/backend/internal/handler"
	"taskflow-pro/backend/internal/middleware"
	"taskflow-pro/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Setup(h *handler.Handler, auth *service.AuthService, redis *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(redis))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)

		protected := api.Group("")
		protected.Use(middleware.Auth(auth))
		{
			protected.GET("/auth/me", h.Me)
			protected.GET("/projects", h.ListProjects)
			protected.POST("/projects", h.CreateProject)
			protected.POST("/projects/:id/members", h.AddProjectMember)
			protected.GET("/projects/:id/stats", h.ProjectStats)
			protected.GET("/projects/:id/tasks", h.ListTasks)
			protected.POST("/projects/:id/tasks", h.CreateTask)
			protected.PUT("/tasks/:id", h.UpdateTask)
			protected.DELETE("/tasks/:id", h.DeleteTask)
			protected.GET("/tasks/:id/comments", h.ListComments)
			protected.POST("/tasks/:id/comments", h.CreateComment)
		}
	}

	return r
}
