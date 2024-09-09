package worker

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/google/uuid"
	pb "github.com/kolharsam/task-scheduler/pkg/grpc-api"
)

type WorkerContext struct {
	Logger           *zap.Logger
	RingLeaderClient pb.SchedulerClient
	ServiceId        uuid.UUID
}

func SetupConnectionWithLeader(host string, port int64) (*grpc.ClientConn, error) {
	ringLeaderTarget := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.NewClient(ringLeaderTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ring leader at %s: %w", ringLeaderTarget, err)
	}
	return conn, nil
}

func (ws *WorkerContext) ManageHeartBeats() {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	stream, err := ws.RingLeaderClient.HeartBeat(context.Background())
	if err != nil {
		ws.Logger.Sugar().Fatalf(
			"failed to initiate communication with ring-leader via heartbeats [%v]",
			err,
		)
	}

	for range ticker.C {
		err := stream.Send(&pb.HeartbeatRequest{
			ServiceId:          ws.ServiceId.String(),
			HeartBeatTimestamp: time.Now().Format(time.RFC3339),
		})

		if err != nil {
			ws.Logger.Fatal("failed to send heartbeat request to server")
		}
	}

	stream.CloseSend()
}

func (ws *WorkerContext) RunWorker() error {
	stream, err := ws.RingLeaderClient.RunTask(context.Background(), &pb.TaskReadyRequest{
		ServerId:  ws.ServiceId.String(),
		Timestamp: time.Now().Format(time.RFC3339),
	})

	if err != nil {
		return fmt.Errorf("failed to initiate run-tasks with ring-leader [%v]", err)
	}

	for {
		task, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			ws.Logger.Error(
				"failed to read task from leader",
				zap.String("service_id",
					ws.ServiceId.String()),
				zap.Error(err),
			)
			return err
		}
		go ws.handleTask(task)
	}
	return nil
}

func (ws *WorkerContext) handleTask(task *pb.TaskRequest) {
	stream, err := ws.RingLeaderClient.ReadTaskUpdates(context.Background())

	if err != nil {
		ws.Logger.Error(
			"failed to set up stream to send updates from worker",
			zap.String("service_id", ws.ServiceId.String()),
			zap.Error(err),
		)
	}

	err = stream.Send(&pb.TaskUpdate{
		TaskId:       task.TaskId,
		State:        pb.TaskState_INIT,
		Timestamp:    time.Now().Format(time.RFC3339),
		ErrorDetails: "",
	})

	if err != nil {
		ws.Logger.Error(
			"failed to send INIT task update from worker",
			zap.String("service_id", ws.ServiceId.String()),
			zap.String("task_id", task.TaskId),
			zap.Error(err),
		)
	}

	c := exec.Command(task.Command)
	ch := make(chan error)
	go func(st grpc.ClientStreamingClient[pb.TaskUpdate, pb.TaskCompleteResponse]) {
		err = st.Send(&pb.TaskUpdate{
			TaskId:       task.TaskId,
			State:        pb.TaskState_RUNNING,
			Timestamp:    time.Now().Format(time.RFC3339),
			ErrorDetails: "",
		})
		if err != nil {
			ws.Logger.Error(
				"failed to send RUNNING task update from worker",
				zap.String("service_id", ws.ServiceId.String()),
				zap.String("task_id", task.TaskId),
				zap.Error(err),
			)
			ch <- err
		}
		ch <- c.Run()
	}(stream)

	err = <-ch
	if err != nil {
		e := stream.Send(&pb.TaskUpdate{
			Timestamp:    time.Now().Format(time.RFC3339),
			TaskId:       task.TaskId,
			State:        pb.TaskState_ERROR,
			ErrorDetails: err.Error(),
		})
		if e != nil {
			ws.Logger.Error(
				"failed to send ERROR task update from worker",
				zap.String("service_id", ws.ServiceId.String()),
				zap.String("task_id", task.TaskId),
				zap.Error(err),
			)
		}
	} else {
		e := stream.Send(&pb.TaskUpdate{
			Timestamp:    time.Now().Format(time.RFC3339),
			TaskId:       task.TaskId,
			State:        pb.TaskState_SUCCESS,
			ErrorDetails: err.Error(),
		})
		if e != nil {
			ws.Logger.Error(
				"failed to send SUCCESS task update from worker",
				zap.String("service_id", ws.ServiceId.String()),
				zap.String("task_id", task.TaskId),
				zap.Error(err),
			)
		}
	}

	stream.CloseAndRecv()
}
