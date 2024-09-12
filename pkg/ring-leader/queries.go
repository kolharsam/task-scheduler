package ringLeader

const (
	GET_LATEST_CREATED_TASKS string = `
		SELECT * FROM "public"."tasks" WHERE status IN ('CREATED', 'RUNNING')
		ORDER BY created_at ASC;
	`
	INSERT_TASK_STATUS_UPDATE string = `
		INSERT INTO "public"."task_status_updates_log" 
		(status, task_id, data, worker_id) VALUES (@status, @taskId, @data, @workerId)
		ON CONFLICT (task_id, status) DO NOTHING;
	`

	UPDATE_WORKER_LIVE_STATUS string = `
		INSERT INTO "public"."live_workers" (service_id, status, port, host, connected_at) 
		VALUES (@serviceId, @workerStatus, @port, @host, @connectedAt) ON CONFLICT (host, port)
		DO UPDATE SET status = @workerStatus, connected_at = @connectedAt, service_id = @serviceId;
	`

	CHECK_CURRENT_TASK_STATUS string = `
		SELECT * FROM "public"."tasks" WHERE task_id = @taskId;
	`

	// NOTE: this is run while the ring-leader sets up
	// and forgets about the workers that were present
	// in the last session
	TRUNCATE_LIVE_WORKERS string = `
		TRUNCATE TABLE "public"."live_workers";
	`
)
