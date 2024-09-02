package schedulerapi

import (
	"log"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/kolharsam/task-scheduler/pkg/lib"
)

func Run(serverPort string) {
	router := gin.Default()

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

	apiCtx := NewAPIContext(db, logger)

	router.POST(task_route, apiCtx.taskPostHandler)
	router.GET(task_route, apiCtx.taskGetHandler)
	router.GET(health_route, apiCtx.statusHandler)

	log.Default().Printf("starting scheuler-api on port[%s]...", serverPort)
	router.Run(serverPort)
}
