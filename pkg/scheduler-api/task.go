package schedulerapi

import (
	"github.com/gin-gonic/gin"
	"github.com/kolharsam/task-scheduler/pkg/lib"
)

func (api *APIContext) taskPostHandler(c *gin.Context) {
	c.JSON(200, lib.JSON{
		"status": "OK",
	})
}

func (api *APIContext) taskGetHandler(c *gin.Context) {
	c.JSON(200, lib.JSON{
		"status": "OK",
	})
}
