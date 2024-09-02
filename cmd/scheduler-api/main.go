package main

import (
	"flag"
	"log"

	"github.com/joho/godotenv"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	schedulerapi "github.com/kolharsam/task-scheduler/pkg/scheduler-api"
)

var (
	port = flag.String("port", "8080", "--port=8080")
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to set up environment variables [%v]", err)
	}

	serverPort := lib.MakePortString(*port)

	schedulerapi.Run(serverPort)
}
