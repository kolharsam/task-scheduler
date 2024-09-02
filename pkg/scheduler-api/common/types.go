package common

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type Task struct {
	TaskID  uuid.UUID `json:"task_id"`
	Command string    `json:"command"`
	Status  string    `json:"status"`
	// NOTE: Status is one of 'CREATED' | 'RUNNING' | 'COMPLETED'
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type TaskStatusUpdateLog struct {
	TaskStatusRecordID uuid.UUID `json:"task_status_record_id"`
	TaskID             uuid.UUID `json:"task_id"`
	Status             string    `json:"status"`
	// NOTE: Status is one of
	// 'INITIATED',
	// 'RUNNING',
	// 'SUCCESS',
	// 'FAILED'
	StatusOrder int64     `json:"-"`
	Data        *gin.H    `json:"data"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
