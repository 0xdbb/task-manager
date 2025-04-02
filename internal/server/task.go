package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "task-manager/internal/database/sqlc"
	"task-manager/internal/token"
	"task-manager/util"
)

var (
	taskQueue   = "task_queue"
	priorityMap = map[string]uint8{
		"HIGH":   10,
		"MEDIUM": 5,
		"LOW":    0,
	}
)

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
func (s *Server) GetTasks(ctx *gin.Context) {
	if !isUserRoleAllowed(ctx, db.UserRoleADMIN) {
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

	tasks, err := s.db.ListAllTasks(ctx, arg)
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
func (s *Server) GetTask(ctx *gin.Context) {
	if !isUserRoleAllowed(ctx, db.UserRoleSTANDARD) {
		return
	}

	var taskReq TaskRequest
	if err := ctx.ShouldBindUri(&taskReq); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}
	taskID, err := uuid.Parse(taskReq.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	task, err := s.db.GetTask(ctx, taskID)
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

// @Summary		Long Poll Task Status
// @Description	Wait until a task's status changes from PENDING before responding
// @Tags			tasks
// @Accept			json
// @Produce		json
// @Security BearerAuth
// @Param id path string true "Task ID (UUID format)"
// @Success		200	{object}	Message
// @Failure		400	{object}	ErrorResponse
// @Failure		404	{object}	ErrorResponse
// @Failure		500	{object}	ErrorResponse
// @Router			/task/{id}/status [get]
func (s *Server) LongPollTaskStatus(ctx *gin.Context) {
	var taskReq TaskRequest
	if err := ctx.ShouldBindUri(&taskReq); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	taskID, err := uuid.Parse(taskReq.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid task ID"))
		return
	}

	timeout := 60 // seconds
	for i := 0; i < timeout; i++ {
		task, err := s.db.GetTask(ctx, taskID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, HandleError(err, http.StatusNotFound, "Task not found"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error retrieving task"))
			return
		}

		// If the status is no longer PENDING, return immediately
		if task.Status.TaskStatus != db.TaskStatusPENDING {
			ctx.JSON(http.StatusOK, Message{
				TaskID:  task.ID.String(),
				Message: fmt.Sprintf("Task status updated - %v", task.Status),
				Result: task.Result.String,
			})
			return
		}

		// Wait for 1 second before checking again
		time.Sleep(1 * time.Second)
	}

	// If timeout reached and task is still pending, return timeout response
	ctx.JSON(http.StatusRequestTimeout, HandleError(nil, http.StatusRequestTimeout, "Task is still pending, try again later"))
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
func (s *Server) CreateTask(ctx *gin.Context) {
	if !isUserRoleAllowed(ctx, db.UserRoleSTANDARD) {
		return
	}

	// Apply rate limiting
	userID := ctx.MustGet(authorizationPayloadKey).(*token.Payload).UserID

	limiter := getRateLimiter(userID.String())

	if !limiter.Allow() {
		ctx.JSON(http.StatusTooManyRequests, HandleError(nil, http.StatusTooManyRequests, "too many requests. please slow down."))
		return
	}

	var task CreateTaskRequest
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	dueTime, err := util.ParseTimeString(task.DueTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err, http.StatusBadRequest, "Invalid time format"))
		return
	}

	taskArg := db.CreateTaskParams{
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Type:        task.Type, // DATA_PROCESSING or REPORT_GENERATION
		Payload:     task.Payload,
		DueTime:     dueTime,
		Priority:    db.TaskPriority(task.Priority),
	}

	createdTask, err := s.db.CreateTask(ctx, taskArg)
	if err != nil {
		if db.ErrorCode(err) == db.ForeignKeyViolation {
			ctx.JSON(http.StatusForbidden, HandleError(err, http.StatusForbidden, "UserID does not exist"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error creating user"))
		return
	}

	sendTaskPayload := SendTaskPayload{
		ID:          createdTask.ID,
		Title:       createdTask.Title,
		Type:        createdTask.Type,
		Description: createdTask.Description,
		UserID:      createdTask.UserID,
		Priority:    task.Priority,
		Payload:     createdTask.Payload,
		DueTime:     createdTask.DueTime,
	}

	taskBytes, err := json.Marshal(sendTaskPayload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error serializing task request"))
		return
	}

	err = s.queueManager.Publish(taskQueue, taskBytes, priorityMap[task.Priority])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err, http.StatusInternalServerError, "Error Publishing Task to queue"))
		return
	}
	log.Printf(" [x] Sent %s", task.Payload)

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
func (s *Server) UpdateTaskStatus(ctx *gin.Context) {
	if !isUserRoleAllowed(ctx, db.UserRoleADMIN) {
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

	taskArg := db.UpdateTaskStatusParams{}

	task, err := s.db.UpdateTaskStatus(ctx, taskArg)
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
