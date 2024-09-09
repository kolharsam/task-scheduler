package main

import (
	"flag"
	"log"

	ringLeader "github.com/kolharsam/task-scheduler/pkg/ring-leader"
)

var (
	host = flag.String("host", "localhost", "--host localhost")
	port = flag.Int64("port", 8081, "--port 8081")
)

func main() {
	flag.Parse()
	lis, server, err := ringLeader.GetListenerAndServer(*host, int32(*port))
	if err != nil {
		log.Fatalf("failed to setup ring-leader server %v", err)
	}
	err = server.Serve(lis)
	if err != nil {
		log.Fatalf("failure at ring-leader server at [%s:%d]", *host, *port)
	}
}
