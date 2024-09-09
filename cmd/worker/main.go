package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	pb "github.com/kolharsam/task-scheduler/pkg/grpc-api"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	"github.com/kolharsam/task-scheduler/pkg/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	leaderHost := os.Getenv("RING_LEADER_HOST")
	leaderPort := os.Getenv("RING_LEADER_PORT")
	var port int64
	if leaderPort == "" {
		port = 8081
	} else {
		portp, _ := strconv.Atoi(leaderPort)
		port = int64(portp)
	}

	logger, err := lib.GetLogger()
	if err != nil {
		log.Fatalf("failed to set up logger for worker service %v", err)
	}

	var connection *grpc.ClientConn
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		connection, err = worker.SetupConnectionWithLeader(leaderHost, port)
		if err == nil {
			break
		}
		logger.Warn("Failed to connect with ring-leader. Retrying...",
			zap.Error(err),
			zap.Int("attempt", i+1),
			zap.Int("max_retries", maxRetries),
		)
		time.Sleep(time.Second * 5) // Wait 5 seconds before retrying
	}

	if err != nil || connection == nil {
		logger.Fatal("Failed to connect with the ring-leader after multiple attempts", zap.Error(err))
	}

	logger.Info("connected with ring-leader")

	defer connection.Close()

	client := pb.NewSchedulerClient(connection)
	serviceId, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("failed to set up worker_id [%v]", err)
	}

	workerContext := &worker.WorkerContext{
		RingLeaderClient: client,
		Logger:           logger,
		ServiceId:        serviceId,
	}

	go workerContext.ManageHeartBeats()
	err = workerContext.RunWorker()
	if err != nil {
		logger.Fatal("error has occurred while running worker",
			zap.Error(err),
			zap.String("worker_id", serviceId.String()))
	}
}
