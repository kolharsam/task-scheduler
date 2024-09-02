package common

// ROUTE constants
var (
	TaskRoute = "/task"
	// serves as status endpoint of service db and other infra pieces
	HealthRoute        = "/healthz"
	AllTaskEventsRoute = "/task/events/:task_id"
)
