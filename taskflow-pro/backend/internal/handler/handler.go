package handler

import (
	"errors"
	"net/http"
	"strconv"

	"taskflow-pro/backend/internal/middleware"
	"taskflow-pro/backend/internal/model"
	"taskflow-pro/backend/internal/repository"
	"taskflow-pro/backend/internal/response"
	"taskflow-pro/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	auth     *service.AuthService
	projects *service.ProjectService
	tasks    *service.TaskService
	comments *service.CommentService
}

func New(auth *service.AuthService, projects *service.ProjectService, tasks *service.TaskService, comments *service.CommentService) *Handler {
	return &Handler{auth: auth, projects: projects, tasks: tasks, comments: comments}
}

func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "注册参数不正确")
		return
	}
	data, err := h.auth.Register(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, data)
}

func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "登录参数不正确")
		return
	}
	data, err := h.auth.Login(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Me(c *gin.Context) {
	user, err := h.auth.Me(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, user)
}

func (h *Handler) CreateProject(c *gin.Context) {
	var req model.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "项目参数不正确")
		return
	}
	project, err := h.projects.Create(middleware.CurrentUserID(c), req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, project)
}

func (h *Handler) ListProjects(c *gin.Context) {
	projects, err := h.projects.List(middleware.CurrentUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, projects)
}

func (h *Handler) AddProjectMember(c *gin.Context) {
	projectID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req model.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "成员参数不正确")
		return
	}
	if err := h.projects.AddMember(projectID, middleware.CurrentUserID(c), req.UserID); err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, gin.H{"added": true})
}

func (h *Handler) ProjectStats(c *gin.Context) {
	projectID, ok := parseID(c, "id")
	if !ok {
		return
	}
	stats, err := h.projects.Stats(c.Request.Context(), projectID, middleware.CurrentUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, stats)
}

func (h *Handler) CreateTask(c *gin.Context) {
	projectID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req model.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "任务参数不正确")
		return
	}
	task, err := h.tasks.Create(c.Request.Context(), projectID, middleware.CurrentUserID(c), req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, task)
}

func (h *Handler) ListTasks(c *gin.Context) {
	projectID, ok := parseID(c, "id")
	if !ok {
		return
	}
	tasks, err := h.tasks.List(projectID, middleware.CurrentUserID(c), c.Query("status"), c.Query("keyword"))
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, tasks)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	taskID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req model.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "任务参数不正确")
		return
	}
	task, err := h.tasks.Update(c.Request.Context(), taskID, middleware.CurrentUserID(c), req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, task)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	taskID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.tasks.Delete(c.Request.Context(), taskID, middleware.CurrentUserID(c)); err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) CreateComment(c *gin.Context) {
	taskID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req model.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "评论内容不能为空")
		return
	}
	comment, err := h.comments.Create(taskID, middleware.CurrentUserID(c), req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, comment)
}

func (h *Handler) ListComments(c *gin.Context) {
	taskID, ok := parseID(c, "id")
	if !ok {
		return
	}
	comments, err := h.comments.List(taskID, middleware.CurrentUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, comments)
}

func parseID(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, "ID 参数不正确")
		return 0, false
	}
	return uint(id), true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		response.Fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrEmailExists):
		response.Fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrForbidden):
		response.Fail(c, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrInvalidStatus), errors.Is(err, service.ErrInvalidPriority):
		response.Fail(c, http.StatusBadRequest, err.Error())
	case repository.IsNotFound(err), errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, http.StatusNotFound, "数据不存在")
	default:
		response.Fail(c, http.StatusInternalServerError, "服务器内部错误")
	}
}
