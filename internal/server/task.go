package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "task-manager/internal/database/sqlc"
)

// Task Types
type CreateTaskRequest struct {
	Title       string    `json:"title" binding:"required" example:"Image Processing"`
	Type        string    `json:"type" binding:"required" example:"Image Processing"`
	Description string    `json:"description" binding:"required" example:"Image Processing"`
	UserID      uuid.UUID `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Priority    string    `json:"priority" binding:"required" example:"HIGH"`
	Payload     string    `json:"payload" binding:"required" example:"{\"recipient\":\"user@example.com\",\"subject\":\"Welcome\",\"body\":\"Thanks for signing up!\"}"`
	DueTime     time.Time `json:"due_date" binding:"required" example:"2025-03-30T12:00:00Z"`
}

type UpdateTaskRequest struct {
	Status string `json:"status" example:"completed"`
	Result string `json:"result" example:"2025-04-01T12:00:00Z"`
}

type TaskRequest struct {
	ID uuid.UUID `uri:"id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type TasksRequest struct {
	PageSize int32     `form:"page_size" binding:"required,min=1" example:"10"`
	PageID   int32     `form:"page_id" binding:"required,min=1" example:"1"`
	UserID   uuid.UUID `form:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// @Summary		Get all created Tasks
// @Description	Get a list of all tasks with pagination. Supports filtering by passing `user_id` as a query parameter.
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Security BearerAuth
// @Param			page_size	query	int	true	"Page Size"
// @Param			page_id		query	int	true	"Page Number"
// @Success		200			{array}		db.Task
// @Failure		400			{object}	ErrorResponse
// @Failure		500			{object}	ErrorResponse
// @Router			/task [get]
func (h *Server) GetTasks(ctx *gin.Context) {
	if !isAdmin(ctx) {
		return
	}

	var req TasksRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	arg := db.ListAllTasksParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	tasks, err := h.db.ListAllTasks(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error retrieving tasks"))
		return
	}

	ctx.JSON(http.StatusOK, tasks)
}

// @Summary		Get Task
// @Description	Get task by ID
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Security BearerAuth
// @Param id path string true "User ID (UUID format)"
// @Success		200	{object}	db.Task
// @Failure		400	{object}	ErrorResponse
// @Failure		404	{object}	ErrorResponse
// @Failure		500	{object}	ErrorResponse
// @Router			/task/{id} [get]
func (h *Server) GetTask(ctx *gin.Context) {
	if !isAdmin(ctx) {
		return
	}

	var taskReq TaskRequest
	if err := ctx.ShouldBindUri(&taskReq); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	task, err := h.db.GetTask(ctx, taskReq.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, HandleError(err, http.StatusNotFound, "Task not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error retrieving task"))
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// @Summary		Create Task
// @Description	Create a new task
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Security BearerAuth
// @Param			request	body		CreateTaskRequest	true	"Create Task Request"
// @Success		201		{object}	db.Task
// @Failure		400		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/task [post]
func (h *Server) CreateTask(ctx *gin.Context) {
	if !isAdmin(ctx) {
		return
	}

	var task CreateTaskRequest
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	// TODO: Publish to queue

	taskArg := db.CreateTaskParams{
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Type:        "DATA_PROCESSING", // or REPORT_GENERATION or DATA_LABELING or RESULT_REVIEW
		Payload:     task.Payload,
		DueTime:     task.DueTime,
		Priority:    "High",
	}

	createdTask, err := h.db.CreateTask(ctx, taskArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating task"))
		return
	}

	ctx.JSON(http.StatusCreated, createdTask)
}

// @Summary		Update Task Status
// @Description	Update the status of an existing task
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Security BearerAuth
// @Param id path string true "User ID (UUID format)"
// @Param			request	body		UpdateTaskRequest	true	"Update Task Request"
// @Success		200		{object}	db.Task
// @Failure		400		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/task/{id}/status [patch]
func (h *Server) UpdateTaskStatus(ctx *gin.Context) {
	if !isAdmin(ctx) {
		return
	}

	var taskReq TaskRequest
	var update UpdateTaskRequest

	if err := ctx.ShouldBindUri(&taskReq); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	taskArg := db.UpdateTaskStatusParams{
		ID: taskReq.ID,
		Status: db.NullTaskStatus{
			TaskStatus: db.TaskStatus(update.Status),
			Valid:      true,
		},
		Result: toPgTypeText(update.Result),
	}

	task, err := h.db.UpdateTaskStatus(ctx, taskArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error updating task"))
		return
	}

	ctx.JSON(http.StatusOK, task)
}
func toPgTypeText(text string) pgtype.Text {
	return pgtype.Text{
		String: text,
		Valid:  text != "",
	}
}
