package schedulerapi

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kolharsam/task-scheduler/pkg/lib"
)

func (api *APIContext) statusHandler(c *gin.Context) {
	err := api.db.Ping(context.Background())
	if err != nil {
		c.Errors = append(c.Errors, &gin.Error{
			Err: fmt.Errorf("failed to reach database [%v]", err),
		})
		return
	}

	// TODO: do similar check with ring-leader as well to update status of the whole system

	c.JSON(200, lib.JSON{
		"status": "OK",
	})
}
