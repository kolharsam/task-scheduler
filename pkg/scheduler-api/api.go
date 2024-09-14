package schedulerapi

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	constants "github.com/kolharsam/task-scheduler/pkg/scheduler-api/common"
	"github.com/kolharsam/task-scheduler/pkg/scheduler-api/handlers"
	"go.uber.org/zap"
)

var (
	retryNumber = flag.Int64("retry", 5, "--retry 5 [time between retries to connect with db]")
)

var (
	taskGetRoute      string = fmt.Sprintf("%s/*task_id", constants.TaskRoute)
	taskEventGetRoute string = fmt.Sprintf("%s/*task_id", constants.TaskEventsRoute)
)

func Run(serverPort string) {
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

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

	ctx := context.Background()
	err = db.Ping(ctx)
	maxRetries := 10

	for err != nil {
		if maxRetries <= 0 {
			break
		}
		logger.Warn("failed to connect with database...retrying in 5 seconds", zap.Error(err))
		time.Sleep(time.Second * time.Duration(*retryNumber))
		err = db.Ping(ctx)
		maxRetries--
	}
	if err != nil {
		logger.Fatal("failed to set up an active connection with the database", zap.Error(err))
	}

	logger.Info("successfully connected with the database...")

	err = handlers.TruncateWorkersTable(db)
	if err != nil {
		log.Fatalf("failure in starting up scheduler-api service %v", err)
	}

	v1.POST(constants.TaskRoute, apiCtx.TaskPostHandler)
	v1.GET(constants.TaskRoute, apiCtx.TaskGetHandler)
	v1.GET(taskGetRoute, apiCtx.TaskGetHandler)
	v1.GET(constants.TaskEventsRoute, apiCtx.GetAllTaskEventsHandler)
	v1.GET(taskEventGetRoute, apiCtx.GetAllTaskEventsHandler)
	v1.GET(constants.HealthRoute, apiCtx.StatusHandler)

	log.Default().Printf("starting scheuler-api on port[%s]...", serverPort)

	router.Run(serverPort)
}
