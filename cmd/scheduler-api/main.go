package main

import (
	"flag"
	"log"

	"github.com/kolharsam/task-scheduler/pkg/lib"
	schedulerapi "github.com/kolharsam/task-scheduler/pkg/scheduler-api"
)

var (
	port = flag.String("port", "8080", "--port=8080")
)

func main() {
	server, err := schedulerapi.NewServer()
	if err != nil {
		log.Fatalf("failed to run scheduler-api service [%v]", err)
		return
	}

	portStr := lib.MakePortString(*port)

	log.Default().Printf("starting scheuler-api on port[%s]...", *port)
	err = server.Run(portStr)

	if err != nil {
		log.Fatalf("there was issue while starting the server [%v]", err)
	}
}
