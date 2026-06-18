package database

import (
	"context"
	"fmt"
	"time"

	"taskflow-pro/backend/internal/config"
	"taskflow-pro/backend/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func Connect(cfg config.Config) (*Store, error) {
	db, err := connectMySQL(cfg.Database)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	return &Store{DB: db, Redis: rdb}, nil
}

func connectMySQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for attempt := 1; attempt <= 20; attempt++ {
		db, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			return db, nil
		}
		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("connect mysql: %w", err)
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.ProjectMember{},
		&model.Task{},
		&model.Comment{},
	)
}
