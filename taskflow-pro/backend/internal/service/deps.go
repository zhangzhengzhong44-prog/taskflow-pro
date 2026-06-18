package service

import (
	"context"
	"time"

	"taskflow-pro/backend/internal/model"

	"github.com/redis/go-redis/v9"
)

type UserStore interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
}

type ProjectStore interface {
	CreateWithOwner(project *model.Project) error
	ListByUser(userID uint) ([]model.Project, error)
	FindByID(projectID uint) (*model.Project, error)
	IsMember(projectID uint, userID uint) (bool, error)
	IsOwner(projectID uint, userID uint) (bool, error)
	AddMember(projectID uint, userID uint) error
	Stats(projectID uint) (model.ProjectStats, error)
}

type TaskStore interface {
	Create(task *model.Task) error
	List(projectID uint, status string, keyword string) ([]model.Task, error)
	FindByID(taskID uint) (*model.Task, error)
	Update(task *model.Task) error
	Delete(taskID uint) error
}

type CommentStore interface {
	Create(comment *model.Comment) error
	ListByTask(taskID uint) ([]model.Comment, error)
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
