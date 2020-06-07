package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

const (
	Map    = "Map"
	Reduce = "Reduce"
	Wait   = "Wait" // no task currently available
	Exit   = "Exit" // all tasks done
)

type RequestTaskArgs struct {
}

type RequestTaskReply struct {
	TaskType string

	// Used by map task
	MapTaskNumber int
	FileName      string
	NReduce       int

	// Used by reduce task
	ReduceTaskNumber int
	NMap             int
}

type TaskDoneArgs struct {
	TaskType   string
	TaskNumber int
	FileNames  []string // Single element for reduce task
}

type TaskDoneReply struct {
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
