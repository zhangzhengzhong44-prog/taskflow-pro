package repository

import (
	"errors"

	"taskflow-pro/backend/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) CreateWithOwner(project *model.Project) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(project).Error; err != nil {
			return err
		}
		member := model.ProjectMember{ProjectID: project.ID, UserID: project.OwnerID, Role: model.RoleOwner}
		return tx.Create(&member).Error
	})
}

func (r *ProjectRepository) ListByUser(userID uint) ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Table("projects").
		Select("projects.*").
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userID).
		Order("projects.updated_at DESC").
		Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) FindByID(projectID uint) (*model.Project, error) {
	var project model.Project
	if err := r.db.First(&project, projectID).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) IsMember(projectID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ProjectRepository) IsOwner(projectID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND role = ?", projectID, userID, model.RoleOwner).
		Count(&count).Error
	return count > 0, err
}

func (r *ProjectRepository) AddMember(projectID uint, userID uint) error {
	member := model.ProjectMember{ProjectID: projectID, UserID: userID, Role: model.RoleMember}
	return r.db.Create(&member).Error
}

func (r *ProjectRepository) Stats(projectID uint) (model.ProjectStats, error) {
	var stats model.ProjectStats
	if err := r.db.Model(&model.Task{}).Where("project_id = ?", projectID).Count(&stats.Total).Error; err != nil {
		return stats, err
	}
	if err := r.db.Model(&model.Task{}).Where("project_id = ? AND status = ?", projectID, model.TaskTodo).Count(&stats.Todo).Error; err != nil {
		return stats, err
	}
	if err := r.db.Model(&model.Task{}).Where("project_id = ? AND status = ?", projectID, model.TaskDoing).Count(&stats.Doing).Error; err != nil {
		return stats, err
	}
	if err := r.db.Model(&model.Task{}).Where("project_id = ? AND status = ?", projectID, model.TaskDone).Count(&stats.Done).Error; err != nil {
		return stats, err
	}
	return stats, nil
}

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) List(projectID uint, status string, keyword string) ([]model.Task, error) {
	var tasks []model.Task
	query := r.db.Where("project_id = ?", projectID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	err := query.Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) FindByID(taskID uint) (*model.Task, error) {
	var task model.Task
	if err := r.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

func (r *TaskRepository) Delete(taskID uint) error {
	result := r.db.Delete(&model.Task{}, taskID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) ListByTask(taskID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Where("task_id = ?", taskID).Order("created_at ASC").Find(&comments).Error
	return comments, err
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
