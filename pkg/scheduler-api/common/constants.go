package common

// ROUTE constants
var (
	TaskRoute = "/v1/task"
	// serves as status endpoint of service db and other infra pieces
	HealthRoute        = "/v1/healthz"
	AllTaskEventsRoute = "/v1/events/:task_id"
)
