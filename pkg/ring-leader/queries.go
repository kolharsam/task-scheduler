package ringLeader

const (
	GET_LATEST_CREATED_TASKS string = `
		SELECT * FROM "public"."tasks" WHERE status IN ('CREATED')
		ORDER BY created_at ASC;
	`
	INSERT_TASK_STATUS_UPDATE string = `
		INSERT INTO "public"."task_status_updates_log" 
		(status, task_id) VALUES (@status, @taskId);
	`

	UPDATE_WORKER_LIVE_STATUS string = `
		INSERT INTO "public"."live_workers" (service_id, status) 
		VALUES (@serviceId, @workerStatus) ON CONFLICT (service_id)
		DO UPDATE SET status = @workerStatus;
	`
)
