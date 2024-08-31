package ringleader

import (
	"fmt"

	ringLeader "github.com/kolharsam/task-scheduler/pkg/ring-leader"
)

func main() {
	fmt.Println(ringLeader.Leader())
}
