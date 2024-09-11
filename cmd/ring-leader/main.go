package main

import (
	"flag"
	"log"

	"github.com/kolharsam/task-scheduler/pkg/lib"
	ringLeader "github.com/kolharsam/task-scheduler/pkg/ring-leader"
)

var (
	host = flag.String("host", "localhost", "--host localhost")
	port = flag.Int64("port", 8081, "--port 8081")
)

func main() {
	flag.Parse()
	lis, server, serverCtx, err := ringLeader.GetListenerAndServer(*host, uint32(*port))
	if err != nil {
		log.Fatalf("failed to setup ring-leader server %v", err)
	}

	go serverCtx.CheckHearbeats()

	taskQueue := make(chan *lib.Task, 1024)
	go serverCtx.FetchTasks(taskQueue)
	go serverCtx.RunTasks(taskQueue)

	err = server.Serve(lis)
	if err != nil {
		log.Fatalf("failure at ring-leader server at [%s:%d]", *host, *port)
	}
}
