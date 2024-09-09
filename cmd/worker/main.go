package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	pb "github.com/kolharsam/task-scheduler/pkg/grpc-api"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	"github.com/kolharsam/task-scheduler/pkg/worker"
	"go.uber.org/zap"
)

var (
	//--ring-leader-host ring-leader --ring-leader-port 8081
	ringLeaderHost = flag.String("ring-leader-host", "ring_leader", "--ring-leader-host")
	ringLeaderPort = flag.Int64("ring-leader-port", 8081, "--ring-leader-port 8081")
)

func main() {
	flag.Parse()
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

	connection, err := worker.SetupConnectionWithLeader(leaderHost, port)
	maxRetries := 10
	for err != nil {
		if maxRetries <= 0 {
			break
		}
		logger.Warn("failed to connect with ring-leader....reconnecting in some time",
			zap.Error(err),
			zap.Int("retries_left", maxRetries),
		)
		time.Sleep(time.Second * 5)
		connection, err = worker.SetupConnectionWithLeader(*ringLeaderHost, *ringLeaderPort)
		maxRetries--
	}
	if err != nil || connection == nil {
		logger.Fatal("failed to connect with the ring-leader", zap.Error(err))
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
