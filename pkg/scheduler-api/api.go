package schedulerapi

import (
	"fmt"
	"log"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	constants "github.com/kolharsam/task-scheduler/pkg/scheduler-api/common"
	"github.com/kolharsam/task-scheduler/pkg/scheduler-api/handlers"
)

var (
	taskGetRoute string = fmt.Sprintf("%s/*task_id", constants.TaskRoute)
)

func Run(serverPort string) {
	router := gin.Default()
	v1 := router.Group(constants.APIVersion)

	logger, err := lib.GetLogger()
	if err != nil {
		log.Fatalf("there was issue while setting up the logger [%v]", err)
	}

	db, err := lib.GetDBConnectionPool()

	if err != nil {
		log.Fatalf("failed to set up connection pool with the database [%v]", err)
	}

	defer db.Close()

	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	apiCtx := handlers.NewAPIContext(db, logger)

	v1.POST(constants.TaskRoute, apiCtx.TaskPostHandler)
	v1.GET(taskGetRoute, apiCtx.TaskGetHandler)
	v1.GET(constants.HealthRoute, apiCtx.StatusHandler)
	v1.GET(constants.TaskEventsRoute, apiCtx.GetAllTaskEventsHandler)

	log.Default().Printf("starting scheuler-api on port[%s]...", serverPort)
	router.Run(serverPort)
}
