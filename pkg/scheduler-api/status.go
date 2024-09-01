package schedulerapi

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/kolharsam/task-scheduler/pkg/lib"
)

func (api *APIContext) statusHandler(c *gin.Context) {
	var dbCheckRes string

	queryRes, err := api.db.Query(context.Background(), "SELECT 1;")
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{
			Err: errors.New("failed to reach database"),
		})
		return
	}

	err = queryRes.Scan(&dbCheckRes)
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{
			Err: errors.New("failed to read data from database"),
		})
		return
	}

	if dbCheckRes != "" {
		c.Errors = append(c.Errors, &gin.Error{
			Err: errors.New("invalid response obtained from database"),
		})
	}

	// TODO: do similar check with ring-leader as well to update status of the whole system

	c.JSON(200, lib.JSON{
		"status": "OK",
	})
}
