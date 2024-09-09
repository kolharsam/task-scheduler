package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	"github.com/kolharsam/task-scheduler/pkg/scheduler-api/common"
)

type AddTaskRequest struct {
	Command string `json:"command" binding:"required"`
}

func (api *APIContext) TaskPostHandler(c *gin.Context) {
	var taskRequestBody AddTaskRequest
	if err := c.ShouldBindJSON(&taskRequestBody); err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := common.INSERT_ONE_TASK
	args := pgx.NamedArgs{
		"command": taskRequestBody.Command,
	}

	var field uuid.UUID
	err := api.db.QueryRow(c.Request.Context(), query, args).Scan(&field)
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert task to db", "err": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"task_id": field.String(),
		},
	})
}

func (api *APIContext) TaskGetHandler(c *gin.Context) {
	taskIdStr := c.Param("task_id")
	taskId, err := uuid.Parse(taskIdStr)
	query := common.GET_ALL_TASKS
	args := pgx.NamedArgs{}

	if taskIdStr != "" && err == nil {
		query = common.GET_ONE_TASK
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

	collectedRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[lib.Task])

	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
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
	taskIdStr := c.Param("task_id")
	taskId, err := uuid.Parse(taskIdStr)
	var query string
	var args pgx.NamedArgs

	if err != nil && taskIdStr != "" {
		err = errors.New("task_id must be provided or an incorrect one has been provided")
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err != nil && taskIdStr == "" {
		query = common.GET_ALL_EVENTS
		args = pgx.NamedArgs{}
	} else {
		query = common.GET_ALL_EVENTS_ONE_TASK
		args = pgx.NamedArgs{
			"taskId": taskId.String(),
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

	collectedRows, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[lib.TaskStatusUpdateLog],
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

func TruncateWorkersTable(db *pgxpool.Pool) error {
	_, err := db.Query(context.Background(), common.TRUNCATE_LIVE_WORKERS)
	if err != nil {
		return err
	}
	return nil
}
