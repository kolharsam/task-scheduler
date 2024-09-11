package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/kolharsam/task-scheduler/pkg/worker"
)

var (
	host = flag.String("host", "", "--host localhost")
	port = flag.Int64("port", 0, "--port 8081")
)

func main() {
	flag.Parse()

	leaderHost := os.Getenv("RING_LEADER_HOST")
	leaderPort := os.Getenv("RING_LEADER_PORT")
	var ringLeaderPort uint32
	if leaderPort == "" {
		ringLeaderPort = 8081
	} else {
		portp, _ := strconv.Atoi(leaderPort)
		ringLeaderPort = uint32(portp)
	}

	lis, server, workerCtx, err := worker.GetListenerAndServer(*host, uint32(*port), leaderHost, ringLeaderPort)
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
