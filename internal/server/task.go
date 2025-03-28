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
	Type       string    `json:"title" binding:"required" example:"Complete project"`
	Payload string    `json:"payload" example:"Example payload"`
	DueDate     time.Time `json:"due_date" example:"2025-03-30T12:00:00Z"`
}

type UpdateTaskRequest struct {
	Status string `json:"status" example:"completed"`
	Result string `json:"result" example:"2025-04-01T12:00:00Z"`
}

type TaskRequest struct {
	ID uuid.UUID `uri:"id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type TasksRequest struct {
	PageSize int32 `form:"page_size" binding:"required,min=1" example:"10"`
	PageID   int32 `form:"page_id" binding:"required,min=1" example:"1"`
}

// TaskResponse represents the response structure for tasks
type TaskResponse struct {
	ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title       string    `json:"title" example:"Complete project"`
	Description string    `json:"description" example:"Finish the pending project by Friday"`
	Status      string    `json:"status" example:"pending"`
	DueDate     time.Time `json:"due_date" example:"2025-03-30T12:00:00Z"`
	CreatedAt   time.Time `json:"created_at" example:"2025-03-25T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2025-03-26T12:00:00Z"`
}

func newTaskResponse(task db.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

// @Summary		Get Tasks
// @Description	Get a list of tasks with pagination
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Param			page_size	query	int	true	"Page Size"
// @Param			page_id		query	int	true	"Page Number"
// @Success		200			{array}		TaskResponse
// @Failure		400			{object}	ErrorResponse
// @Failure		500			{object}	ErrorResponse
// @Router			/task [get]
func (h *Server) GetTasks(ctx *gin.Context) {
	var req TasksRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request - missing page_id or page_size"))
		return
	}

	arg := db.ListTasksParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	tasks, err := h.db.ListTasks(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error retrieving tasks"))
		return
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, newTaskResponse(task))
	}

	ctx.JSON(http.StatusOK, taskResponses)
}

// @Summary		Get Task
// @Description	Get task by ID
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Param			id	path		uuid.UUID	true	"Task ID"
// @Success		200	{object}	TaskResponse
// @Failure		400	{object}	ErrorResponse
// @Failure		404	{object}	ErrorResponse
// @Failure		500	{object}	ErrorResponse
// @Router			/task/{id} [get]
func (h *Server) GetTask(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, newTaskResponse(task))
}

// @Summary		Create Task
// @Description	Create a new task
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Param			request	body		CreateTaskRequest	true	"Create Task Request"
// @Success		201		{object}	TaskResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/task [post]
func (h *Server) CreateTask(ctx *gin.Context) {
	var task CreateTaskRequest

	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	// TODO: Apply authorization

	taskArg := db.CreateTaskParams{
		// UserID:  ,
		// Type:  ,
		// Payload: "",
		DueTime: task.DueDate,
	}

	createdTask, err := h.db.CreateTask(ctx, taskArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating task"))
		return
	}

	ctx.JSON(http.StatusCreated, newTaskResponse(createdTask))
}

// @Summary		Update Task Status
// @Description	Update the status of an existing task
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Param			id		path		uuid.UUID			true	"Task ID"
// @Param			request	body		UpdateTaskRequest	true	"Update Task Request"
// @Success		200		{object}	TaskResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		500		{object}	ErrorResponse
// @Router			/task/{id} [put]
func (h *Server) UpdateTaskStatus(ctx *gin.Context) {
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
		ID:     taskReq.ID,
		Status: toPgTypeText(update.Status),
		Result: toPgTypeText(update.Result),
	}

	task, err := h.db.UpdateTaskStatus(ctx, taskArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error updating task"))
		return
	}

	ctx.JSON(http.StatusOK, newTaskResponse(task))
}

func toPgTypeText(text string) pgtype.Text {
	return pgtype.Text{
		String: text,
		Valid:  text != "",
	}

}
