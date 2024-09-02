package common

const GET_ALL_TASKS = `
	SELECT * FROM "public"."tasks";
`

const INSERT_ONE_TASK = `
	INSERT INTO "public"."tasks" (command) VALUES (@command)
`

const GET_ONE_TASK = `
	SELECT * FROM "public"."tasks" where task_id = @taskId;
`

const GET_ALL_TASK_EVENTS = `
	SELECT * FROM "public"."task_status_updates_log" where task_id = @taskId;
`
