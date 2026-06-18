package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"taskflow-pro/backend/internal/model"
)

var (
	ErrForbidden       = errors.New("没有权限")
	ErrInvalidStatus   = errors.New("任务状态不合法")
	ErrInvalidPriority = errors.New("任务优先级不合法")
)

type ProjectService struct {
	projects ProjectStore
	users    UserStore
	cache    Cache
}

func NewProjectService(projects ProjectStore, users UserStore, cache Cache) *ProjectService {
	return &ProjectService{projects: projects, users: users, cache: cache}
}

func (s *ProjectService) Create(userID uint, req model.CreateProjectRequest) (*model.Project, error) {
	project := &model.Project{Name: req.Name, Description: req.Description, OwnerID: userID}
	return project, s.projects.CreateWithOwner(project)
}

func (s *ProjectService) List(userID uint) ([]model.Project, error) {
	return s.projects.ListByUser(userID)
}

func (s *ProjectService) AddMember(projectID uint, operatorID uint, userID uint) error {
	isOwner, err := s.projects.IsOwner(projectID, operatorID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrForbidden
	}
	if _, err := s.users.FindByID(userID); err != nil {
		return err
	}
	return s.projects.AddMember(projectID, userID)
}

func (s *ProjectService) Stats(ctx context.Context, projectID uint, userID uint) (model.ProjectStats, error) {
	if err := s.ensureMember(projectID, userID); err != nil {
		return model.ProjectStats{}, err
	}

	key := fmt.Sprintf("project:%d:stats", projectID)
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, key)
		if err == nil {
			var stats model.ProjectStats
			if json.Unmarshal([]byte(cached), &stats) == nil {
				return stats, nil
			}
		}
	}

	stats, err := s.projects.Stats(projectID)
	if err != nil {
		return model.ProjectStats{}, err
	}
	data, _ := json.Marshal(stats)
	if s.cache != nil {
		_ = s.cache.Set(ctx, key, data, 2*time.Minute)
	}
	return stats, nil
}

func (s *ProjectService) ensureMember(projectID uint, userID uint) error {
	isMember, err := s.projects.IsMember(projectID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrForbidden
	}
	return nil
}

type TaskService struct {
	projects ProjectStore
	tasks    TaskStore
	cache    Cache
}

func NewTaskService(projects ProjectStore, tasks TaskStore, cache Cache) *TaskService {
	return &TaskService{projects: projects, tasks: tasks, cache: cache}
}

func (s *TaskService) Create(ctx context.Context, projectID uint, userID uint, req model.CreateTaskRequest) (*model.Task, error) {
	if err := s.ensureMember(projectID, userID); err != nil {
		return nil, err
	}
	priority := req.Priority
	if priority == "" {
		priority = model.PriorityMedium
	}
	if !model.IsValidTaskPriority(priority) {
		return nil, ErrInvalidPriority
	}

	task := &model.Task{
		ProjectID:   projectID,
		CreatorID:   userID,
		AssigneeID:  req.AssigneeID,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskTodo,
		Priority:    priority,
		DueDate:     req.DueDate,
	}
	if err := s.tasks.Create(task); err != nil {
		return nil, err
	}
	s.invalidateStats(ctx, projectID)
	return task, nil
}

func (s *TaskService) List(projectID uint, userID uint, status string, keyword string) ([]model.Task, error) {
	if err := s.ensureMember(projectID, userID); err != nil {
		return nil, err
	}
	if status != "" && !model.IsValidTaskStatus(model.TaskStatus(status)) {
		return nil, ErrInvalidStatus
	}
	return s.tasks.List(projectID, status, keyword)
}

func (s *TaskService) Update(ctx context.Context, taskID uint, userID uint, req model.UpdateTaskRequest) (*model.Task, error) {
	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureMember(task.ProjectID, userID); err != nil {
		return nil, err
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.Status != nil {
		if !model.IsValidTaskStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		task.Status = *req.Status
	}
	if req.Priority != nil {
		if !model.IsValidTaskPriority(*req.Priority) {
			return nil, ErrInvalidPriority
		}
		task.Priority = *req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if err := s.tasks.Update(task); err != nil {
		return nil, err
	}
	s.invalidateStats(ctx, task.ProjectID)
	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, taskID uint, userID uint) error {
	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return err
	}
	if err := s.ensureMember(task.ProjectID, userID); err != nil {
		return err
	}
	if err := s.tasks.Delete(taskID); err != nil {
		return err
	}
	s.invalidateStats(ctx, task.ProjectID)
	return nil
}

func (s *TaskService) ensureMember(projectID uint, userID uint) error {
	isMember, err := s.projects.IsMember(projectID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrForbidden
	}
	return nil
}

func (s *TaskService) invalidateStats(ctx context.Context, projectID uint) {
	if s.cache != nil {
		_ = s.cache.Del(ctx, fmt.Sprintf("project:%d:stats", projectID))
	}
}

type CommentService struct {
	projects ProjectStore
	tasks    TaskStore
	comments CommentStore
}

func NewCommentService(projects ProjectStore, tasks TaskStore, comments CommentStore) *CommentService {
	return &CommentService{projects: projects, tasks: tasks, comments: comments}
}

func (s *CommentService) Create(taskID uint, userID uint, req model.CreateCommentRequest) (*model.Comment, error) {
	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	isMember, err := s.projects.IsMember(task.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrForbidden
	}
	comment := &model.Comment{TaskID: taskID, UserID: userID, Content: req.Content}
	return comment, s.comments.Create(comment)
}

func (s *CommentService) List(taskID uint, userID uint) ([]model.Comment, error) {
	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	isMember, err := s.projects.IsMember(task.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrForbidden
	}
	return s.comments.ListByTask(taskID)
}
