package worker

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/google/uuid"
	"github.com/kolharsam/task-scheduler/pkg/config"
	pb "github.com/kolharsam/task-scheduler/pkg/grpc-api"
	"github.com/kolharsam/task-scheduler/pkg/lib"
)

type leaderInfo struct {
	host string
	port uint32
}

type workerContext struct {
	pb.UnimplementedWorkerServer
	logger              *zap.Logger
	serviceId           string
	workerHost          string
	workerPort          uint32
	leaderInfo          leaderInfo
	isConnectedToLeader bool
	mu                  sync.Mutex
	appConfig           *config.TaskSchedulerConfig
}

func setupConnectionWithLeader(host string, port uint32) (pb.RingLeaderClient, error) {
	ringLeaderTarget := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(ringLeaderTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ring leader at %s: %w", ringLeaderTarget, err)
	}
	return pb.NewRingLeaderClient(conn), nil
}

func (wc *workerContext) RunTask(taskRequest *pb.TaskRequest, stream grpc.ServerStreamingServer[pb.TaskUpdate]) error {
	wc.logger.Info("new task request is here...", zap.String("task_id", taskRequest.TaskId))

	err := stream.Send(&pb.TaskUpdate{
		Timestamp:     time.Now().Format(time.RFC3339),
		TaskId:        taskRequest.TaskId,
		State:         pb.TaskState_RUNNING,
		ErrorDetails:  "",
		CommandResult: "",
	})

	if err != nil {
		wc.logger.Error("failed to send update to leader...", zap.Error(err))
		return err
	}

	cmd := exec.Command(taskRequest.Command)
	var stderr strings.Builder
	var stdout strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil || stderr.String() != "" {
		wc.logger.Error("failed to run command on worker...",
			zap.String("task_id", taskRequest.TaskId),
			zap.String("worker_id", wc.serviceId),
			zap.Error(err),
		)
		stream.Send(&pb.TaskUpdate{
			Timestamp:     time.Now().Format(time.RFC3339),
			TaskId:        taskRequest.TaskId,
			State:         pb.TaskState_ERROR,
			ErrorDetails:  stdout.String(),
			CommandResult: "",
		})
		return err
	}

	err = stream.Send(&pb.TaskUpdate{
		Timestamp:     time.Now().Format(time.RFC3339),
		TaskId:        taskRequest.TaskId,
		State:         pb.TaskState_SUCCESS,
		ErrorDetails:  "",
		CommandResult: stdout.String(),
	})

	if err != nil {
		wc.logger.Error("failed to send SUCCESS message to leader...", zap.Error(err),
			zap.String("task_id", taskRequest.TaskId))
		return err
	}

	return nil
}

func (wc *workerContext) ConnectWithLeader() {
	sleepNumber := time.Duration(
		wc.appConfig.WorkerConfig.Connections.TimeBetweenRetries,
	)

	for {
		ringLeaderClient, err := setupConnectionWithLeader(wc.leaderInfo.host, wc.leaderInfo.port)

		if err != nil {
			wc.logger.Warn("failed to set up client to connect with leader...", zap.Error(err))
			time.Sleep(sleepNumber * time.Second)
			continue
		}

		ack, err := ringLeaderClient.Connect(context.Background(), &pb.ConnectRequest{
			ServiceId:   wc.serviceId,
			ServiceHost: wc.workerHost,
			Port:        wc.workerPort,
			TimeStamp:   time.Now().Format(time.RFC3339),
		})

		if err != nil {
			wc.logger.Warn("failed to get ack from ring-leader",
				zap.Error(err),
				zap.String("ring-leader-host", wc.leaderInfo.host),
				zap.Uint32("ring-leader-port", wc.leaderInfo.port),
			)
			time.Sleep(sleepNumber * time.Second)
			continue
		}

		wc.logger.Info("connected with leader...", zap.Any("ring-leader-host", ack.GetHost()))
		wc.mu.Lock()
		wc.isConnectedToLeader = true
		wc.mu.Unlock()
		return
	}
}

func (wc *workerContext) HandleHeartbeats() {
	backoff := time.Second
	maxBackoff := time.Duration(wc.appConfig.WorkerConfig.BackoffMax) * time.Minute

	for {
		ringLeaderClient, err := setupConnectionWithLeader(wc.leaderInfo.host, wc.leaderInfo.port)
		if err != nil {
			wc.logger.Warn("failed to set up client to connect with leader...", zap.Error(err))
			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)
		}

		stream, err := ringLeaderClient.Hearbeat(context.Background())
		if err != nil {
			wc.logger.Warn("failed to set up heartbeats with leader...", zap.Error(err))
			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		backoff = time.Second // Reset time if we connect properly

		go func() {
			for {
				beat, err := stream.Recv()
				if err == io.EOF {
					wc.logger.Warn("heartbeat stream closed by leader")
					return
				}
				if err != nil {
					wc.logger.Warn("failed to recv ack for heartbeat",
						zap.Error(err),
						zap.Any("worker_info", map[string]interface{}{
							"worker_id":   wc.serviceId,
							"worker_port": wc.workerPort,
							"leader_info": wc.leaderInfo,
						}))
					return
				}

				if beat.LeaderStatus != pb.LeaderStatus_ACTIVE {
					wc.logger.Warn("there seems to be an issue at the leader...", zap.Any("worker_id", wc.serviceId))
				}
			}
		}()

		ticker := time.NewTicker(
			time.Duration(wc.appConfig.WorkerConfig.HeartbeatInterval) * time.Second,
		)
		defer ticker.Stop()

		for range ticker.C {
			err := stream.Send(&pb.HeartbeatFromWorker{
				ServiceId: wc.serviceId,
				Timestamp: time.Now().Format(time.RFC3339),
				Host:      wc.workerHost,
				Port:      wc.workerPort,
			})

			if err != nil {
				wc.logger.Warn("failed to send a heartbeat to leader...",
					zap.String("worker_id", wc.serviceId),
					zap.Error(err))
				wc.mu.Lock()
				wc.isConnectedToLeader = false
				wc.mu.Unlock()
				break
			}
		}

		// NOTE: try to re-establish the connection with the leader
		wc.ConnectWithLeader()
	}
}

func newServer(logger *zap.Logger, serviceId string, host string, port uint32, leaderHost string, leaderPort uint32, config *config.TaskSchedulerConfig) *workerContext {
	return &workerContext{
		logger:              logger,
		serviceId:           serviceId,
		workerHost:          host,
		workerPort:          port,
		leaderInfo:          leaderInfo{host: leaderHost, port: leaderPort},
		isConnectedToLeader: false,
		appConfig:           config,
	}
}

func GetListenerAndServer(host string, port uint32, ringLeaderHost string, ringLeaderPort uint32, config *config.TaskSchedulerConfig) (net.Listener, *grpc.Server, *workerContext, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, nil, nil, err
	}

	logger, err := lib.GetLogger()
	if err != nil {
		log.Fatalf("failed to initiate logger for worker [%v]", err)
		return nil, nil, nil, err
	}

	serviceId := uuid.New()

	grpcServer := grpc.NewServer()

	workerCtx := newServer(logger, serviceId.String(), host, port, ringLeaderHost, ringLeaderPort, config)

	pb.RegisterWorkerServer(grpcServer, workerCtx)

	workerCtx.logger.Info("initiating server...", zap.Any("worker_info", workerCtx))

	return listener, grpcServer, workerCtx, nil
}
