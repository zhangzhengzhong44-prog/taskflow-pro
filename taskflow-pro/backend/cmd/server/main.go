package main

import (
	"log"

	"taskflow-pro/backend/internal/config"
	"taskflow-pro/backend/internal/database"
	"taskflow-pro/backend/internal/handler"
	"taskflow-pro/backend/internal/repository"
	"taskflow-pro/backend/internal/router"
	"taskflow-pro/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.App.Mode)

	store, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect dependencies: %v", err)
	}
	if err := database.AutoMigrate(store.DB); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	userRepo := repository.NewUserRepository(store.DB)
	projectRepo := repository.NewProjectRepository(store.DB)
	taskRepo := repository.NewTaskRepository(store.DB)
	commentRepo := repository.NewCommentRepository(store.DB)
	cache := service.NewRedisCache(store.Redis)

	authSvc := service.NewAuthService(userRepo, cache, cfg.JWT)
	projectSvc := service.NewProjectService(projectRepo, userRepo, cache)
	taskSvc := service.NewTaskService(projectRepo, taskRepo, cache)
	commentSvc := service.NewCommentService(projectRepo, taskRepo, commentRepo)

	h := handler.New(authSvc, projectSvc, taskSvc, commentSvc)
	r := router.Setup(h, authSvc, store.Redis)

	log.Printf("TaskFlow Pro API listening on :%s", cfg.App.Port)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
