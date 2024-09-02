package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *APIContext) StatusHandler(c *gin.Context) {
	err := api.db.Ping(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Errorf("failed to reach database [%v]", err.Error()),
		})
		return
	}

	// TODO: do similar check with ring-leader as well to update status of the whole system

	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}
