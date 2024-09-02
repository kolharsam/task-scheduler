package common

const APIVersion string = "v1"

// ROUTE constants
var (
	TaskRoute = "/task"
	// serves as status endpoint of service db and other infra pieces
	HealthRoute     = "/healthz"
	TaskEventsRoute = "/events/*task_id"
)
