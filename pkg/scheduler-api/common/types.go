package common

import (
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type Task struct {
	TaskID  uuid.UUID `json:"task_id"`
	Command string    `json:"command"`
	Status  string    `json:"status"`
	// NOTE: Status is one of 'CREATED' | 'RUNNING' | 'COMPLETED'
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	CompletedAt *string `json:"completed_at"`
}

type TaskStatusUpdateLog struct {
	TaskStatusRecordID uuid.UUID `json:"task_status_record_id"`
	TaskID             uuid.UUID `json:"task_id"`
	Status             string    `json:"status"`
	// NOTE: Status is one of
	// 'CREATED',
	// 'INITIATED',
	// 'RUNNING',
	// 'SUCCESS',
	// 'FAILED'
	Data      gin.H  `json:"data"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
