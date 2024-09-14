package ringLeader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	omap "github.com/elliotchance/orderedmap/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kolharsam/task-scheduler/pkg/config"
	pb "github.com/kolharsam/task-scheduler/pkg/grpc-api"
	"github.com/kolharsam/task-scheduler/pkg/lib"
)

type workerId = string

type taskWorkerInfo struct {
	ServiceId     string    `json:"service_id"`
	ServiceHost   string    `json:"service_host"`
	Port          uint32    `json:"port"`
	LastHeartBeat time.Time `json:"last_heartbeat"`
}

type taskWorkers struct {
	mtx     sync.RWMutex
	workers *omap.OrderedMap[workerId, *taskWorkerInfo]
	next    uint32
}

// nextWorkerForTask distributes the tasks amongst the connected workers
// by using the round-robin algorithm. The omap isn't the best data structure
// to get this done efficiently. Improvements will be made in the future versions
func nextWorkerForTask(ts *taskWorkers) *taskWorkerInfo {
	ts.mtx.Lock()
	n := atomic.AddUint32(&ts.next, 1)
	ts.mtx.Unlock()

	if int(n) > ts.workers.Len() {
		ts.mtx.Lock()
		atomic.StoreUint32(&ts.next, 1)
		ts.mtx.Unlock()
		n = 1
	}

	nxtWorkerIndex := (int(n) - 1) % ts.workers.Len()

	iter := 0
	for el := ts.workers.Front(); el != nil; el = el.Next() {
		if iter == nxtWorkerIndex {
			return el.Value
		}
		iter++
	}
	return nil
}

type ringLeaderServer struct {
	pb.UnimplementedRingLeaderServer
	db            *pgxpool.Pool
	activeServers *taskWorkers
	logger        *zap.Logger
	leaderHost    string
	leaderPort    uint32
	appConfig     *config.TaskSchedulerConfig
}

type connectionRequest struct {
	serviceId   string
	serviceHost string
	port        uint32
	timeStamp   string
}

func (tsi *taskWorkerInfo) updateHeartbeatTimestamp(timeStamp string) error {
	tm, err := time.Parse(time.RFC3339, timeStamp)
	if err != nil {
		return err
	}
	tsi.LastHeartBeat = tm
	return nil
}

func (ts *taskWorkers) addNewService(connectRequest connectionRequest) error {
	tm, err := time.Parse(time.RFC3339, connectRequest.timeStamp)

	if err != nil {
		return err
	}

	tsi := taskWorkerInfo{
		ServiceId:     connectRequest.serviceId,
		ServiceHost:   connectRequest.serviceHost,
		Port:          connectRequest.port,
		LastHeartBeat: tm,
	}

	ts.workers.Set(connectRequest.serviceId, &tsi)

	return nil
}

func newWorkerServiceClient(host string, port uint32) (pb.WorkerClient, error) {
	workerTarget := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(workerTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return pb.NewWorkerClient(conn), nil
}

func (ts *taskWorkers) updateServiceHeartbeat(serviceId, timestamp string) error {
	if taskWorker, ok := ts.workers.Get(serviceId); ok {
		if err := taskWorker.updateHeartbeatTimestamp(timestamp); err != nil {
			return err
		}
	}
	return nil
}

func (ts *taskWorkers) removeService(serviceId string) *taskWorkerInfo {
	val, ok := ts.workers.Get(serviceId)
	if !ok {
		return nil
	}
	ts.workers.Delete(serviceId)
	return val
}

func (rls *ringLeaderServer) Hearbeat(stream grpc.BidiStreamingServer[pb.HeartbeatFromWorker, pb.HeartbeatFromLeader]) error {
	for {
		beat, err := stream.Recv()
		if err == io.EOF {
			rls.activeServers.mtx.Lock()
			rls.activeServers.removeService(beat.ServiceId)
			rls.activeServers.mtx.Unlock()
			return nil
		}

		if err != nil {
			rls.logger.Error("failed to receive heartbeat from worker...")
			return err
		}

		workerId := beat.ServiceId
		beatTime := beat.Timestamp

		rls.activeServers.mtx.Lock()
		rls.activeServers.updateServiceHeartbeat(workerId, beatTime)
		rls.activeServers.mtx.Unlock()

		_, err = rls.db.Exec(context.Background(), UPDATE_WORKER_LIVE_STATUS, pgx.NamedArgs{
			"serviceId":    beat.ServiceId,
			"workerStatus": lib.RUNNING_WORKER_STATUS,
			"host":         beat.Host,
			"port":         beat.Port,
			"connectedAt":  time.Now().Format(time.RFC3339),
		})

		if err != nil {
			rls.logger.Warn("failed to update the live status of db...",
				zap.String("worker_status", lib.RUNNING_WORKER_STATUS),
				zap.String("worker_id", beat.ServiceId),
				zap.Error(err))
		}

		rls.logger.Info("updated the worker status from heartbeat...",
			zap.String("worker_id", workerId),
		)

		stream.Send(&pb.HeartbeatFromLeader{
			Timestamp:    time.Now().Format(time.RFC3339),
			LeaderStatus: pb.LeaderStatus_ACTIVE,
		})
	}
}

func (rls *ringLeaderServer) CheckHearbeats() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for range ticker.C {
		if rls.activeServers.workers.Len() == 0 {
			continue
		}

		rls.activeServers.mtx.RLock()

		for el := rls.activeServers.workers.Front(); el != nil; el = el.Next() {
			if time.Since(el.Value.LastHeartBeat) >= (time.Second * 15) {
				rls.logger.Warn("worker seems to be down...",
					zap.String("worker_id", el.Value.ServiceId))
				_, err := rls.db.Exec(context.Background(), UPDATE_WORKER_LIVE_STATUS, pgx.NamedArgs{
					"serviceId":    el.Value.ServiceId,
					"workerStatus": lib.ERRORED_WORKER_STATUS,
					"host":         el.Value.ServiceHost,
					"port":         el.Value.Port,
				})
				if err != nil {
					rls.logger.Warn("failed to update db about ERRORED status of worker...",
						zap.String("worker_id", el.Value.ServiceId))
				}
			}
		}

		rls.activeServers.mtx.RUnlock()

	}
}

func (rls *ringLeaderServer) Connect(ctx context.Context, connReq *pb.ConnectRequest) (*pb.ConnectAck, error) {
	connectRequest := connectionRequest{
		serviceId:   connReq.GetServiceId(),
		serviceHost: connReq.GetServiceHost(),
		port:        connReq.GetPort(),
		timeStamp:   connReq.GetTimeStamp(),
	}

	rls.activeServers.mtx.Lock()
	err := rls.activeServers.addNewService(connectRequest)
	rls.activeServers.mtx.Unlock()

	if err != nil {
		return nil, err
	}

	_, err = rls.db.Exec(context.Background(), UPDATE_WORKER_LIVE_STATUS, pgx.NamedArgs{
		"serviceId":    connectRequest.serviceId,
		"workerStatus": lib.RUNNING_WORKER_STATUS,
		"host":         connectRequest.serviceHost,
		"port":         connectRequest.port,
		"connected_at": time.Now().Format(time.RFC3339),
	})

	if err != nil {
		// NOTE: it is okay if we fail here and should not stop the execution
		// of the rest of the function
		rls.logger.Warn("failed to write about new worker to db....", zap.Error(err))
	}

	rls.logger.Info("connected with new worker...",
		zap.String("worker_host", connectRequest.serviceHost),
		zap.Uint32("worker_port", connectRequest.port),
		zap.String("worker_id", connectRequest.serviceId),
	)

	return &pb.ConnectAck{
		Host:      rls.leaderHost,
		Port:      rls.leaderPort,
		TimeStamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (rls *ringLeaderServer) handleTask(taskWorker *taskWorkerInfo, task *lib.Task) {
	if taskWorker == nil || task == nil {
		rls.logger.Warn("task could not be handled...")
		return
	}

	rows, err := rls.db.Query(context.Background(), CHECK_CURRENT_TASK_STATUS, pgx.NamedArgs{
		"taskId": task.TaskID.String(),
	})

	if err != nil {
		rls.logger.Warn("failed to fetch the latest info on the given task...",
			zap.Error(err),
			zap.String("task_id", task.TaskID.String()),
			zap.String("worker_id", taskWorker.ServiceId),
		)
		return
	}

	dbTask, err := pgx.CollectRows(rows, pgx.RowToStructByName[lib.Task])

	if err != nil {
		rls.logger.Warn("failed to convert the latest info on the given task...",
			zap.Error(err),
			zap.String("task_id", task.TaskID.String()),
			zap.String("worker_id", taskWorker.ServiceId),
		)
	}

	if dbTask[0].Status != task.Status {
		rls.logger.Warn("task has a different status than status in DB...",
			zap.String("task_id", task.TaskID.String()),
		)
		return
	}

	// NOTE: this might cause a delay in the execution of the tasks because a client
	// is being set up for each and every task. this option seemed better than causing
	// the entire server to crash when one client was shared across all of the tasks
	workerClient, err := newWorkerServiceClient(taskWorker.ServiceHost, taskWorker.Port)

	if err != nil {
		rls.logger.Error("failed to set up client for task...",
			zap.String("task_id", task.TaskID.String()),
			zap.Error(err))
		return
	}

	stream, err := workerClient.RunTask(context.Background(), &pb.TaskRequest{
		Command:     task.Command,
		TaskId:      task.TaskID.String(),
		RequestTime: time.Now().Format(time.RFC3339),
	})

	if err != nil {
		rls.logger.Error("failed to set up task stream with worker",
			zap.String("worker_id", taskWorker.ServiceId),
			zap.String("task_id", task.TaskID.String()),
			zap.Error(err))
		return
	}

	for {
		taskUpdate, err := stream.Recv()

		if err == io.EOF {
			stream.CloseSend()
			return
		}

		if err != nil {
			rls.logger.Error("failed to read update from worker...",
				zap.String("worker_id", taskWorker.ServiceId),
				zap.String("task_id", task.TaskID.String()),
				zap.Error(err),
			)
			// FIXME: this might end up retrying things and we might
			// want that to happen only a certain number of times
			continue
		}

		rls.logger.Info("task update incoming...",
			zap.String("task_id", taskUpdate.TaskId),
			zap.String("task_status", taskUpdate.State.String()),
			zap.String("worker_id", taskWorker.ServiceId),
		)

		data := map[string]interface{}{
			"stderr": taskUpdate.ErrorDetails,
			"stdout": taskUpdate.CommandResult,
		}

		dataBytes, err := json.Marshal(data)

		if err != nil {
			rls.logger.Warn("failed to send data to db...", zap.Error(err))
		}

		updateArgs := pgx.NamedArgs{
			"status":   taskUpdate.State.String(),
			"taskId":   taskUpdate.TaskId,
			"data":     dataBytes,
			"workerId": taskWorker.ServiceId,
		}

		_, err = rls.db.Exec(context.Background(), INSERT_TASK_STATUS_UPDATE, updateArgs)

		if err != nil {
			rls.logger.Warn("failed to write update to db...",
				zap.String("task_id", task.TaskID.String()),
				zap.String("worker_id", taskWorker.ServiceId),
				zap.String("status", taskUpdate.State.String()),
				zap.Error(err),
			)
			continue
		}

		rls.logger.Info("task update written to db...",
			zap.String("task_id", taskUpdate.TaskId),
		)
	}
}

func (rls *ringLeaderServer) FetchTasks(taskQueue chan<- *lib.Task) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for range ticker.C {
		if rls.activeServers.workers.Len() == 0 {
			continue
		}

		taskRows, err := rls.db.Query(context.Background(), GET_LATEST_CREATED_TASKS)
		if err != nil {
			rls.logger.Error("failed to fetch the latest tasks...", zap.Error(err))
		}
		tasks, err := pgx.CollectRows(taskRows, pgx.RowToAddrOfStructByName[lib.Task])
		if err != nil {
			rls.logger.Error("failed to read rows from database...", zap.Error(err))
		}

		if len(tasks) == 0 {
			rls.logger.Info("no new tasks found...")
			continue
		}

		taskRows.Close()

		for _, task := range tasks {
			taskQueue <- task
		}
	}

	close(taskQueue)
}

func (rls *ringLeaderServer) RunTasks(taskQueue <-chan *lib.Task) {
	for task := range taskQueue {
		taskWorkerInfo := nextWorkerForTask(rls.activeServers)
		go rls.handleTask(taskWorkerInfo, task)
	}
}

func newServer(host string, port uint32, db *pgxpool.Pool, logger *zap.Logger, config *config.TaskSchedulerConfig) *ringLeaderServer {
	s := &ringLeaderServer{
		activeServers: &taskWorkers{
			workers: omap.NewOrderedMap[string, *taskWorkerInfo](),
		},
		db:         db,
		logger:     logger,
		leaderHost: host,
		leaderPort: port,
		appConfig:  config,
	}
	return s
}

func GetListenerAndServer(host string, port uint32, config *config.TaskSchedulerConfig) (net.Listener, *grpc.Server, *ringLeaderServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, nil, nil, err
	}

	logger, err := lib.GetLogger()

	if err != nil {
		log.Fatalf("failed to initiate logger for ring-leader [%v]", err)
		return nil, nil, nil, err
	}

	connectionString := lib.GetDBConnectionString()
	maxRetries := config.RingLeaderConfig.Connections.MaxRetries
	contextDB := context.Background()
	db, err := pgxpool.New(contextDB, connectionString)

	for err != nil {
		if maxRetries <= 0 {
			break
		}
		time.Sleep(
			time.Duration(config.RingLeaderConfig.Connections.TimeBetweenRetries) * time.Second,
		)
		logger.Info("trying to connect to db...",
			zap.Int("max_retries", maxRetries),
			zap.Error(err),
			zap.String("connection_string", connectionString))
		db, err = pgxpool.New(contextDB, connectionString)
		maxRetries--
	}

	if err != nil || db == nil {
		log.Fatalf("failed to connect with the database [%v]", err)
		return nil, nil, nil, err
	} else {
		logger.Info("connected with database....", zap.String("connection_string", connectionString))
	}

	ctx := context.Background()
	err = db.Ping(ctx)
	maxRetries = config.RingLeaderConfig.Connections.MaxRetries

	for err != nil {
		if maxRetries <= 0 {
			break
		}
		time.Sleep(
			time.Duration(config.RingLeaderConfig.Connections.TimeBetweenRetries) * time.Second,
		)
		logger.Info("performing ping on db...", zap.Int("max_retries", maxRetries), zap.Error(err))
		err = db.Ping(ctx)
		maxRetries--
	}

	if err == nil {
		logger.Info("successfully connected with database...")
	} else {
		logger.Fatal("failed to connect with the database...", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	serverCtx := newServer(host, port, db, logger, config)
	pb.RegisterRingLeaderServer(grpcServer, serverCtx)
	return listener, grpcServer, serverCtx, nil
}
