package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"taskflow-pro/backend/internal/config"
	"taskflow-pro/backend/internal/model"

	"gorm.io/gorm"
)

func TestRegisterAndLoginIssueUsableToken(t *testing.T) {
	users := newFakeUserStore()
	auth := NewAuthService(users, newMemoryCache(), config.JWTConfig{
		Secret: "test-secret",
		TTL:    time.Hour,
	})

	registered, err := auth.Register(context.Background(), model.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if registered.User.ID == 0 {
		t.Fatal("expected registered user id")
	}

	login, err := auth.Login(context.Background(), model.LoginRequest{
		Email:    "alice@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	claims, err := auth.ParseToken(login.Token)
	if err != nil {
		t.Fatalf("parse login token: %v", err)
	}
	if claims.UserID != registered.User.ID {
		t.Fatalf("expected token user id %d, got %d", registered.User.ID, claims.UserID)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	users := newFakeUserStore()
	auth := NewAuthService(users, newMemoryCache(), config.JWTConfig{
		Secret: "test-secret",
		TTL:    time.Hour,
	})
	_, err := auth.Register(context.Background(), model.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = auth.Login(context.Background(), model.LoginRequest{
		Email:    "alice@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestTaskStatusFlowUpdatesTodoDoingDone(t *testing.T) {
	projects := &fakeProjectStore{members: map[uint]bool{7: true}}
	tasks := newFakeTaskStore()
	svc := NewTaskService(projects, tasks, newMemoryCache())

	task, err := svc.Create(context.Background(), 99, 7, model.CreateTaskRequest{
		Title:    "write api docs",
		Priority: model.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.Status != model.TaskTodo {
		t.Fatalf("new task should start as todo, got %s", task.Status)
	}

	doing := model.TaskDoing
	task, err = svc.Update(context.Background(), task.ID, 7, model.UpdateTaskRequest{Status: &doing})
	if err != nil {
		t.Fatalf("update to doing: %v", err)
	}
	if task.Status != model.TaskDoing {
		t.Fatalf("expected doing, got %s", task.Status)
	}

	done := model.TaskDone
	task, err = svc.Update(context.Background(), task.ID, 7, model.UpdateTaskRequest{Status: &done})
	if err != nil {
		t.Fatalf("update to done: %v", err)
	}
	if task.Status != model.TaskDone {
		t.Fatalf("expected done, got %s", task.Status)
	}
}

func TestTaskStatusFlowRejectsInvalidStatus(t *testing.T) {
	projects := &fakeProjectStore{members: map[uint]bool{7: true}}
	tasks := newFakeTaskStore()
	svc := NewTaskService(projects, tasks, newMemoryCache())

	task, err := svc.Create(context.Background(), 99, 7, model.CreateTaskRequest{
		Title:    "write api docs",
		Priority: model.PriorityMedium,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	blocked := model.TaskStatus("blocked")
	_, err = svc.Update(context.Background(), task.ID, 7, model.UpdateTaskRequest{Status: &blocked})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected invalid status, got %v", err)
	}
}

func TestProjectMemberPermissionBlocksNonMembers(t *testing.T) {
	projects := &fakeProjectStore{members: map[uint]bool{7: false}}
	tasks := newFakeTaskStore()
	svc := NewTaskService(projects, tasks, newMemoryCache())

	_, err := svc.Create(context.Background(), 99, 7, model.CreateTaskRequest{
		Title:    "private task",
		Priority: model.PriorityMedium,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden for non-member, got %v", err)
	}
}

func TestOnlyProjectOwnerCanAddMember(t *testing.T) {
	users := newFakeUserStore()
	_ = users.Create(&model.User{ID: 3, Username: "bob", Email: "bob@example.com"})
	projects := &fakeProjectStore{owners: map[uint]bool{1: true, 2: false}}
	svc := NewProjectService(projects, users, newMemoryCache())

	if err := svc.AddMember(99, 1, 3); err != nil {
		t.Fatalf("owner should add member: %v", err)
	}
	if !projects.members[3] {
		t.Fatal("expected target user to become member")
	}

	err := svc.AddMember(99, 2, 3)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected non-owner forbidden, got %v", err)
	}
}

type fakeUserStore struct {
	nextID  uint
	byID    map[uint]*model.User
	byEmail map[string]*model.User
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		nextID:  1,
		byID:    make(map[uint]*model.User),
		byEmail: make(map[string]*model.User),
	}
}

func (s *fakeUserStore) Create(user *model.User) error {
	if user.ID == 0 {
		user.ID = s.nextID
		s.nextID++
	}
	copied := *user
	s.byID[user.ID] = &copied
	s.byEmail[user.Email] = &copied
	return nil
}

func (s *fakeUserStore) FindByEmail(email string) (*model.User, error) {
	user, ok := s.byEmail[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copied := *user
	return &copied, nil
}

func (s *fakeUserStore) FindByID(id uint) (*model.User, error) {
	user, ok := s.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copied := *user
	return &copied, nil
}

type fakeProjectStore struct {
	members map[uint]bool
	owners  map[uint]bool
	stats   model.ProjectStats
}

func (s *fakeProjectStore) CreateWithOwner(project *model.Project) error {
	project.ID = 99
	return nil
}

func (s *fakeProjectStore) ListByUser(userID uint) ([]model.Project, error) {
	return nil, nil
}

func (s *fakeProjectStore) FindByID(projectID uint) (*model.Project, error) {
	return &model.Project{ID: projectID, OwnerID: 1}, nil
}

func (s *fakeProjectStore) IsMember(projectID uint, userID uint) (bool, error) {
	return s.members[userID], nil
}

func (s *fakeProjectStore) IsOwner(projectID uint, userID uint) (bool, error) {
	return s.owners[userID], nil
}

func (s *fakeProjectStore) AddMember(projectID uint, userID uint) error {
	if s.members == nil {
		s.members = make(map[uint]bool)
	}
	s.members[userID] = true
	return nil
}

func (s *fakeProjectStore) Stats(projectID uint) (model.ProjectStats, error) {
	return s.stats, nil
}

type fakeTaskStore struct {
	nextID uint
	tasks  map[uint]*model.Task
}

func newFakeTaskStore() *fakeTaskStore {
	return &fakeTaskStore{nextID: 1, tasks: make(map[uint]*model.Task)}
}

func (s *fakeTaskStore) Create(task *model.Task) error {
	task.ID = s.nextID
	s.nextID++
	copied := *task
	s.tasks[task.ID] = &copied
	return nil
}

func (s *fakeTaskStore) List(projectID uint, status string, keyword string) ([]model.Task, error) {
	tasks := make([]model.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, *task)
	}
	return tasks, nil
}

func (s *fakeTaskStore) FindByID(taskID uint) (*model.Task, error) {
	task, ok := s.tasks[taskID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copied := *task
	return &copied, nil
}

func (s *fakeTaskStore) Update(task *model.Task) error {
	copied := *task
	s.tasks[task.ID] = &copied
	return nil
}

func (s *fakeTaskStore) Delete(taskID uint) error {
	delete(s.tasks, taskID)
	return nil
}

type memoryCache struct {
	values map[string]string
}

func newMemoryCache() *memoryCache {
	return &memoryCache{values: make(map[string]string)}
}

func (c *memoryCache) Get(ctx context.Context, key string) (string, error) {
	value, ok := c.values[key]
	if !ok {
		return "", errors.New("cache miss")
	}
	return value, nil
}

func (c *memoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	switch typed := value.(type) {
	case string:
		c.values[key] = typed
	case []byte:
		c.values[key] = string(typed)
	default:
		c.values[key] = ""
	}
	return nil
}

func (c *memoryCache) Del(ctx context.Context, key string) error {
	delete(c.values, key)
	return nil
}
