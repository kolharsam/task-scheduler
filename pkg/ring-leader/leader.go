package ringLeader

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/kolharsam/task-scheduler/pkg/grpc-api"
	"github.com/kolharsam/task-scheduler/pkg/lib"
)

type taskServerInfo struct {
	ServiceId     string    `json:"service_id"`
	LastHeartBeat time.Time `json:"last_heartbeat_time"`
}

type taskServers map[string]taskServerInfo

type ringLeaderServer struct {
	pb.UnimplementedSchedulerServer
	db            *pgxpool.Pool
	mtx           sync.RWMutex
	activeServers taskServers
	logger        *zap.Logger
}

func (ts taskServers) addNewService(lastHeartbeatRecorded, serviceId string) bool {
	if _, ok := ts[serviceId]; ok {
		return false
	}
	tm, _ := time.Parse(time.RFC3339, lastHeartbeatRecorded)
	ts[serviceId] = taskServerInfo{
		ServiceId:     serviceId,
		LastHeartBeat: tm,
	}
	return true
}

func (ts taskServers) updateServiceHeartbeat(serviceId, timestamp string) error {
	tm, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return err
	}
	if _, ok := ts[serviceId]; !ok {
		return fmt.Errorf("service [%s] failed to update heartbeat", serviceId)
	}
	ts[serviceId] = taskServerInfo{
		ServiceId:     serviceId,
		LastHeartBeat: tm,
	}
	return nil
}

func (ts taskServers) removeService(serviceId string) *taskServerInfo {
	val, ok := ts[serviceId]
	if !ok {
		return nil
	}
	delete(ts, serviceId)
	return &val
}

func serialize(rq *pb.HeartbeatRequest) string {
	return fmt.Sprintf("port: %s, server_id: %s", rq.ServiceId, rq.HeartBeatTimestamp)
}

// HeartbeatHandler lists all features contained within the given bounding Rectangle.
func (s *ringLeaderServer) HeartbeatHandler(stream pb.Scheduler_HeartBeatServer) error {
	for {
		beat, err := stream.Recv()

		if err == io.EOF {
			s.mtx.Lock()
			removedService := s.activeServers.removeService(beat.ServiceId)
			s.mtx.Unlock()
			if removedService != nil {
				ct := time.Now()
				return stream.SendAndClose(&pb.HeartbeatResponse{
					ServiceId:     s.activeServers[beat.ServiceId].ServiceId,
					SentTimestamp: ct.Format(time.RFC3339),
				})
			}
		}

		if err != nil {
			s.logger.Error("failed to read bytes from client", zap.Error(err))
			return nil
		}
		s.mtx.Lock()
		check := s.activeServers.addNewService(beat.HeartBeatTimestamp, beat.ServiceId)
		s.mtx.Unlock()
		if !check {
			s.mtx.Lock()
			err := s.activeServers.updateServiceHeartbeat(beat.ServiceId, beat.HeartBeatTimestamp)
			s.mtx.Unlock()
			if err != nil {
				log.Printf("failed to record the heartbeat from %v", serialize(beat))
			}
		}
	}
}

func (s *ringLeaderServer) ReadTaskUpdates(stream pb.Scheduler_ReadTaskUpdatesServer) error {
	for {
		taskUpdate, err := stream.Recv()

		if err == io.EOF {
			currentTime := time.Now()
			return stream.SendAndClose(&pb.TaskCompleteResponse{
				Timestamp: currentTime.Format(time.RFC3339),
			})
		}

		if err != nil {
			s.logger.Error("failed to read bytes from client", zap.Error(err))
			return err
		}

		taskUpdateArgs := pgx.NamedArgs{
			"status": taskUpdate.State.String(),
			"taskId": taskUpdate.TaskId,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err = s.db.Exec(ctx, INSERT_TASK_STATUS_UPDATE, taskUpdateArgs)
		if err != nil {
			return fmt.Errorf("failed to insert status update log on to db %v", err)
		}
		s.logger.Info("status updated for task",
			zap.String("status_updated", taskUpdate.State.String()),
			zap.String("task_id", taskUpdate.TaskId),
		)
		if taskUpdate.ErrorDetails != "" {
			s.logger.Info("reason: task failed",
				zap.String("task_id", taskUpdate.TaskId),
				zap.String("error_details", taskUpdate.ErrorDetails))
		}
	}
}

func (s *ringLeaderServer) RunTask(taskReadyReq *pb.TaskReadyRequest, stream pb.Scheduler_RunTaskServer) error {
	// TODO: if a new task is present -> send that to the client
	ticker := time.NewTicker(time.Second * 3) // TODO: make the duration configurable
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		newTasks, err := s.db.Query(ctx, GET_LATEST_CREATED_TASKS)
		if err != nil {
			s.logger.Error("failed to fetch the latest tasks from the db", zap.Error(err))
			continue
		}
		defer newTasks.Close()

		rows, err := pgx.CollectRows(newTasks, pgx.RowToStructByName[lib.Task])
		if err != nil {
			s.logger.Error("failed to convert db result for tasks to go structs", zap.Error(err))
			continue
		}

		if len(rows) == 0 {
			s.logger.Info("no new tasks observed")
			continue
		}

		for _, row := range rows {
			// TODO: make a decision here of which worker to send the task to
			err := stream.Send(&pb.TaskRequest{
				Command:     row.Command,
				TaskId:      row.TaskID.String(),
				RequestTime: row.CreatedAt.Format(time.RFC3339),
			})

			if err != nil {
				s.logger.Error("failed to send task to client", zap.Error(err))
				return err
			}
		}
	}

	return nil
}

func (rls *ringLeaderServer) checkHeartbeats() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for range ticker.C {
		if len(rls.activeServers) == 0 {
			continue
		}
		rls.mtx.RLock()
		for workerServiceId, workerInfo := range rls.activeServers {
			timeSinceLastBeat := time.Since(workerInfo.LastHeartBeat)
			workerStatus := lib.RUNNING_WORKER_STATUS

			if timeSinceLastBeat > (time.Second * 10) {
				workerStatus = lib.ERRORED_WORKER_STATUS
			}

			args := pgx.NamedArgs{
				"workerId":     workerInfo.ServiceId,
				"port":         workerServiceId,
				"workerStatus": workerStatus,
			}

			_, err := rls.db.Query(context.Background(), INSERT_TASK_STATUS_UPDATE, args)
			rls.logger.Warn("a worker doesn't seem to be functioning as expected", zap.Error(err))
		}
		rls.mtx.RUnlock()
	}

}

func newServer(db *pgxpool.Pool, logger *zap.Logger) *ringLeaderServer {
	s := &ringLeaderServer{
		activeServers: make(taskServers),
		db:            db,
		logger:        logger,
	}
	go s.checkHeartbeats()
	return s
}

func GetListenerAndServer(host string, port int32) (net.Listener, *grpc.Server, error) {
	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, nil, err
	}

	logger, err := lib.GetLogger()

	if err != nil {
		log.Fatalf("failed to initiate logger for ring-leader [%v]", err)
		return nil, nil, err
	}

	db, err := lib.GetDBConnectionPool()

	if err != nil {
		log.Fatalf("failed to connect with the database [%v]", err)
		return nil, nil, err
	}

	ctx := context.Background()
	err = db.Ping(ctx)
	maxRetries := 10

	for err != nil {
		if maxRetries <= 0 {
			break
		}
		time.Sleep(4 * time.Second)
		logger.Info("performing ping on db...", zap.Int("max_retries", maxRetries), zap.Error(err))
		err = db.Ping(ctx)
		maxRetries--
	}

	if db != nil && err == nil {
		logger.Info("successfully connected with database...")
	} else {
		logger.Fatal("failed to connect with the database...", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSchedulerServer(grpcServer, newServer(db, logger))
	return listener, grpcServer, nil
}
