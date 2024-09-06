package common

const APIVersion string = "v1"

// ROUTE constants
var (
	TaskRoute = "/task"
	// serves as status endpoint of service db and other infra pieces
	HealthRoute     = "/healthz"
	TaskEventsRoute = "/events/*task_id"
)

// ROUTE constants for tests
var (
	TestTaskRoute = "/v1/task"
	// serves as status endpoint of service db and other infra pieces
	TestHealthRoute     = "/v1/healthz"
	TestTaskEventsRoute = "/v1/events/*task_id"
)
