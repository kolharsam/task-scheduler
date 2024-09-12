package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/kolharsam/task-scheduler/pkg/config"
	"github.com/kolharsam/task-scheduler/pkg/worker"
)

var (
	host       = flag.String("host", "localhost", "--host localhost")
	port       = flag.Int64("port", 9001, "--port 8081")
	configFile = flag.String("config", "config.toml", "--config ../../<PATH_TO_FILE>")
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

	leaderHost := os.Getenv("RING_LEADER_HOST")
	leaderPort := os.Getenv("RING_LEADER_PORT")

	var ringLeaderPort uint32

	if leaderPort == "" {
		ringLeaderPort = 8081
	} else {
		portp, _ := strconv.Atoi(leaderPort)
		ringLeaderPort = uint32(portp)
	}

	lis, server, workerCtx, err := worker.GetListenerAndServer(*host, uint32(*port), leaderHost, ringLeaderPort, appConfig)
	if err != nil {
		log.Fatalf("failed to setup ring-leader server %v", err)
	}

	workerCtx.ConnectWithLeader()
	go workerCtx.HandleHeartbeats()

	err = server.Serve(lis)
	if err != nil {
		log.Fatalf("failure at ring-leader server at [%s:%d]", *host, *port)
	}
}
