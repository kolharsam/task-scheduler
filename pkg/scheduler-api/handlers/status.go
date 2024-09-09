package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	"github.com/kolharsam/task-scheduler/pkg/scheduler-api/common"
)

type statusResponse struct {
	DBStatus      string            `json:"db_status"`
	WorkersStatus map[string]string `json:"workers_status"`
	Error         string            `json:"error"`
}

func (api *APIContext) StatusHandler(c *gin.Context) {
	err := api.db.Ping(context.Background())
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		resp := statusResponse{
			DBStatus:      "NO",
			WorkersStatus: map[string]string{},
			Error:         err.Error(),
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "NO",
			"data":   resp,
		})
		return
	}

	rows, err := api.db.Query(c.Request.Context(), common.GET_ALL_WORKER_STATUS)
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		resp := statusResponse{
			DBStatus:      "OK",
			WorkersStatus: map[string]string{},
			Error:         err.Error(),
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "NO",
			"data":   resp,
		})
		return
	}

	collectedRows, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[lib.LiveWorkers],
	)

	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{Err: err})
		resp := statusResponse{
			DBStatus:      "OK",
			WorkersStatus: map[string]string{},
			Error:         err.Error(),
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "NO",
			"data":   resp,
		})
		return
	}

	workerStatusMap := make(map[string]string)

	for _, workerData := range collectedRows {
		workerStatusMap[workerData.ServiceId] = workerData.Status
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
		"data": statusResponse{
			DBStatus:      "OK",
			WorkersStatus: workerStatusMap,
			Error:         "",
		},
	})
}
