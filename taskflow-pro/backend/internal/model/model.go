package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:40;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:120;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	OwnerID     uint      `gorm:"index;not null" json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectRole string

const (
	RoleOwner  ProjectRole = "owner"
	RoleMember ProjectRole = "member"
)

type ProjectMember struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	ProjectID uint        `gorm:"uniqueIndex:idx_project_user;not null" json:"project_id"`
	UserID    uint        `gorm:"uniqueIndex:idx_project_user;not null" json:"user_id"`
	Role      ProjectRole `gorm:"size:20;not null" json:"role"`
	CreatedAt time.Time   `json:"created_at"`
}

type TaskStatus string

const (
	TaskTodo  TaskStatus = "todo"
	TaskDoing TaskStatus = "doing"
	TaskDone  TaskStatus = "done"
)

type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ProjectID   uint           `gorm:"index;not null" json:"project_id"`
	CreatorID   uint           `gorm:"index;not null" json:"creator_id"`
	AssigneeID  *uint          `gorm:"index" json:"assignee_id"`
	Title       string         `gorm:"size:120;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      TaskStatus     `gorm:"size:20;index;not null" json:"status"`
	Priority    TaskPriority   `gorm:"size:20;index;not null" json:"priority"`
	DueDate     *time.Time     `json:"due_date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    uint      `gorm:"index;not null" json:"task_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=40"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=60"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=80"`
	Description string `json:"description" binding:"max=500"`
}

type AddMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

type CreateTaskRequest struct {
	Title       string       `json:"title" binding:"required,min=2,max=120"`
	Description string       `json:"description"`
	AssigneeID  *uint        `json:"assignee_id"`
	Priority    TaskPriority `json:"priority"`
	DueDate     *time.Time   `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	AssigneeID  *uint         `json:"assignee_id"`
	Status      *TaskStatus   `json:"status"`
	Priority    *TaskPriority `json:"priority"`
	DueDate     *time.Time    `json:"due_date"`
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type ProjectStats struct {
	Total int64 `json:"total"`
	Todo  int64 `json:"todo"`
	Doing int64 `json:"doing"`
	Done  int64 `json:"done"`
}

func ToUserResponse(user User) UserResponse {
	return UserResponse{ID: user.ID, Username: user.Username, Email: user.Email}
}

func IsValidTaskStatus(status TaskStatus) bool {
	return status == TaskTodo || status == TaskDoing || status == TaskDone
}

func IsValidTaskPriority(priority TaskPriority) bool {
	return priority == PriorityLow || priority == PriorityMedium || priority == PriorityHigh
}
