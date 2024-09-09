package ringLeader

import "fmt"

type RingLeaderError struct {
	Op  string
	Err error
}

func (e *RingLeaderError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Usage
// if err != nil {
//     return &RingLeaderError{Op: "fetch_tasks", Err: err}
// }
