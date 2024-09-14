package lib

import (
	"time"

	"github.com/google/uuid"
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
	// 'RUNNING',
	// 'SUCCESS',
	// 'FAILED'
	Data      map[string]interface{} `json:"data"`
	WorkerId  *string                `json:"worker_id"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type LiveWorkers struct {
	WorkerId  int8   `json:"worker_id"`
	ServiceId string `json:"service_id"`
	Status    string `json:"worker_status"`
	// NOTE: Status is one of
	// 'RUNNING',
	// 'ERRORED',
	Port        uint32    `json:"port"`
	Host        string    `json:"host"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ConnectedAt time.Time `json:"connected_at"`
}

type WorkerStatus = string

const (
	RUNNING_WORKER_STATUS WorkerStatus = "RUNNING"
	ERRORED_WORKER_STATUS WorkerStatus = "ERRORED"
)
