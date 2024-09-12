package main

import (
	"flag"
	"log"

	"github.com/kolharsam/task-scheduler/pkg/config"
	"github.com/kolharsam/task-scheduler/pkg/lib"
	ringLeader "github.com/kolharsam/task-scheduler/pkg/ring-leader"
)

var (
	host       = flag.String("host", "localhost", "--host localhost")
	port       = flag.Int64("port", 8081, "--port 8081")
	configFile = flag.String("config", "config.toml", "--config ../../<path-to-config>")
)

func main() {
	flag.Parse()

	appConfig, err := config.ParseConfig(*configFile)

	if err != nil && appConfig != nil {
		log.Println("config file not provided...switching to defaults...")
	} else if err != nil && appConfig == nil {
		log.Fatalf("there's an issue with the config file provided...[%v]", err)
	}

	log.Println("applied config successfully...")

	lis, server, serverCtx, err := ringLeader.GetListenerAndServer(*host, uint32(*port), appConfig)
	if err != nil {
		log.Fatalf("failed to setup ring-leader server %v", err)
	}

	go serverCtx.CheckHearbeats()

	taskQueue := make(chan *lib.Task, appConfig.RingLeaderConfig.TaskQueueSize)
	go serverCtx.FetchTasks(taskQueue)
	go serverCtx.RunTasks(taskQueue)

	err = server.Serve(lis)
	if err != nil {
		log.Fatalf("failure at ring-leader server at [%s:%d]", *host, *port)
	}
}
