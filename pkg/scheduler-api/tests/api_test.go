package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	constants "github.com/kolharsam/task-scheduler/pkg/scheduler-api/common"
	"github.com/kolharsam/task-scheduler/pkg/scheduler-api/handlers"
	"github.com/stretchr/testify/assert"
)

var (
	taskGetRoute string = fmt.Sprintf("%s/*task_id", constants.TaskRoute)
)

func setupTestRouter() *gin.Engine {
	router := gin.Default()
	v1 := router.Group(constants.APIVersion)

	logger, err := lib.GetLogger()
	if err != nil {
		log.Fatalf("there was issue while setting up the logger [%v]", err)
	}
	cs := "postgres://postgres:postgres@localhost:9090/postgres_test"

	db, err := pgxpool.New(context.Background(), cs)

	if err != nil {
		log.Fatalf("failed to set up connection pool with the database [%v]", err)
	}

	apiCtx := handlers.NewAPIContext(db, logger)

	v1.POST(constants.TaskRoute, apiCtx.TaskPostHandler)
	v1.GET(taskGetRoute, apiCtx.TaskGetHandler)
	v1.GET(constants.HealthRoute, apiCtx.StatusHandler)
	v1.GET(constants.TaskEventsRoute, apiCtx.GetAllTaskEventsHandler)

	return router
}

func TestHealthzRoute(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, constants.TestHealthRoute, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"status":"OK"}`, w.Body.String())
	// Summary of the test:
	// 1. set up the router
	// 2. make a call to the /v1/healthz route
	// 3. ensure that the response recorded is Status OK (200)
}

func TestInsertTask(t *testing.T) {
	router := setupTestRouter()
	body := map[string]interface{}{
		"command": "ls",
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, constants.TestTaskRoute, b)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	// Summary of the test:
	// 1. set up the router
	// 2. make a call to the /v1/task route to insert a new task
	// 3. ensure that the response recorded is Status Created (201)
}

type taskIdResponse struct {
	TaskId string `json:"task_id"`
}

func TestGetTasks(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"command": "ls",
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, constants.TestTaskRoute, b)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	var v taskIdResponse
	json.Unmarshal(w.Body.Bytes(), &v)

	w = httptest.NewRecorder()
	testRoute := fmt.Sprintf("%s/%s", constants.TestTaskRoute, v.TaskId)
	req, _ = http.NewRequest(http.MethodGet, testRoute, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// Summary of the test:
	// 1. set up the router
	// 2. make a call to the /v1/task route to insert a new task
	// 3. read the response from the same, with the latest id of the inserted task
	// 4. make GET call to /v1/task route to ensure that there a success response
}

// type taskResponse struct {
// 	TaskId string `json:"task_id"`
// }

// type mapResponse struct {
// 	Data []taskResponse `json:"data"`
// }

// TODO: fix this test
// func TestInsertManyTasksGetManyTasks(t *testing.T) {
// 	router := setupTestRouter()
// 	ids := []string{}

// 	body := map[string]interface{}{
// 		"command": "ls",
// 	}
// 	b := new(bytes.Buffer)
// 	json.NewEncoder(b).Encode(body)

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest(http.MethodPost, constants.TestTaskRoute, b)
// 	req.Header.Set("Content-Type", "application/json")
// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, 201, w.Code)
// 	var v1 taskIdResponse
// 	json.Unmarshal(w.Body.Bytes(), &v1)
// 	ids = append(ids, v1.TaskId)

// 	body = map[string]interface{}{
// 		"command": "ls",
// 	}
// 	b = new(bytes.Buffer)
// 	json.NewEncoder(b).Encode(body)
// 	w = httptest.NewRecorder()
// 	req, _ = http.NewRequest(http.MethodPost, constants.TestTaskRoute, b)
// 	req.Header.Set("Content-Type", "application/json")
// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, 201, w.Code)
// 	var v2 taskIdResponse
// 	json.Unmarshal(w.Body.Bytes(), &v2)
// 	ids = append(ids, v2.TaskId)

// 	w = httptest.NewRecorder()
// 	testRoute := fmt.Sprintf("%s/%s", constants.TestTaskRoute, ids[0])
// 	req, _ = http.NewRequest(http.MethodGet, testRoute, nil)
// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, 200, w.Code)

// 	w = httptest.NewRecorder()
// 	testRoute = fmt.Sprintf("%s/%s", constants.TestTaskRoute, ids[1])
// 	req, _ = http.NewRequest(http.MethodGet, testRoute, nil)
// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, 200, w.Code)

// 	w = httptest.NewRecorder()
// 	req, _ = http.NewRequest(http.MethodGet, constants.TestTaskRoute, nil)
// 	router.ServeHTTP(w, req)

// 	var respMap mapResponse
// 	dec := json.NewDecoder(w.Result().Request.Response.Body)
// 	err := dec.Decode(&respMap)
// 	assert.Equal(t, 200, w.Code)
// 	assert.Nil(t, err)
// 	assert.Equal(t, len(respMap.Data), 2)
// 	// Summary of the test:
// 	// 1. set up the router
// 	// 2. make a call to the /v1/task route to insert a new task (x2)
// 	// 3. read the response from the same, with the latest id of the inserted task (x2)
// 	// 4. make GET call to /v1/task (given id from prev. req)
// 	//    route to ensure that there a success response (x2)
// 	// 5. make GET call to /v1/task without a task id,
// 	//    and see that we've recieved an array of responses (= 2)
// }

// TODO: add tests for the task events API
