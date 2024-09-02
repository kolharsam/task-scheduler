package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	queries "github.com/kolharsam/task-scheduler/pkg/scheduler-api/common"
	types "github.com/kolharsam/task-scheduler/pkg/scheduler-api/common"
)

type AddTaskRequest struct {
	Command string `json:"command" binding:"required"`
}

func (api *APIContext) TaskPostHandler(c *gin.Context) {
	var taskRequestBody AddTaskRequest
	if err := c.ShouldBindJSON(&taskRequestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := queries.INSERT_ONE_TASK
	args := pgx.NamedArgs{
		"command": taskRequestBody.Command,
	}
	_, err := api.db.Exec(c.Request.Context(), query, args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": "task created",
	})
}

func (api *APIContext) TaskGetHandler(c *gin.Context) {
	taskId := c.Param("task_id")
	query := queries.GET_ALL_TASKS
	args := pgx.NamedArgs{}

	if taskId != "" {
		query = queries.GET_ONE_TASK
		args = pgx.NamedArgs{
			"taskId": taskId,
		}
	}

	rows, err := api.db.Query(c.Request.Context(), query, args)

	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("failed to fetch tasks from db [%s]", err.Error()),
		})
		return
	}

	collectedRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Task])

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("failed to read rows[tasks] from db [%s]", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": collectedRows,
	})
}

func (api *APIContext) GetAllTaskEventsHandler(c *gin.Context) {
	taskId := c.Param("task_id")

	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ":task_id must be provided",
		})
		return
	}

	rows, err := api.db.Query(c.Request.Context(), queries.GET_ALL_TASK_EVENTS, pgx.NamedArgs{
		"taskId": taskId,
	})

	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("failed to fetch tasks from db [%s]", err.Error()),
		})
		return
	}

	collectedRows, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[types.TaskStatusUpdateLog],
	)

	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("failed to read rows[task_status_updates_log] from db [%s]", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": collectedRows,
	})
}
