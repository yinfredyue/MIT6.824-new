package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	Idle       = "Idle"
	InProgress = "In progress"
	Completed  = "Completed"
)

type Task struct {
	TaskType  string
	State     string
	StartTime time.Time

	// Used by Map tasks
	FileName string
}

type Master struct {
	mutex sync.Mutex

	nMap        int    // Number of map tasks
	nReduce     int    // Number of reduce tasks
	mapDone     bool   // if all map tasks completed
	reduceDone  bool   // if all reduce tasks completed
	mapTasks    []Task // map task info
	reduceTasks []Task // reduce task info
}

func taskAvailable(task Task) bool {
	isIdle := task.State == Idle
	isStarting := task.StartTime.IsZero()
	isInProgress := task.State == InProgress
	timeLimit := task.StartTime.Add(time.Duration(10 * time.Second))
	timeLimitReached := time.Now().After(timeLimit)
	return isIdle || isStarting || (isInProgress && timeLimitReached)
}

func (m *Master) GiveTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.mapDone && m.reduceDone {
		reply.TaskType = Exit
		return nil
	}

	// Give map task
	if !m.mapDone {
		uncompleted := false
		for i := 0; i < len(m.mapTasks); i++ {
			if m.mapTasks[i].State == Completed {
				continue
			}

			uncompleted = true
			if taskAvailable(m.mapTasks[i]) {
				// Fill in reply
				reply.TaskType = Map
				reply.MapTaskNumber = i
				reply.FileName = m.mapTasks[i].FileName
				reply.NReduce = m.nReduce

				// Update data structure
				m.mapTasks[i].State = InProgress
				m.mapTasks[i].StartTime = time.Now()

				return nil
			}
		}

		// All map tasks assigned
		if uncompleted {
			// Tell worker to wait
			reply.TaskType = Wait
			return nil
		}

		m.mapDone = true
	}

	// Give reduce task
	uncompleted := false
	for i := 0; i < len(m.reduceTasks); i++ {
		if m.reduceTasks[i].State == Completed {
			continue
		}

		uncompleted = true
		if taskAvailable(m.reduceTasks[i]) {
			// Fill in the reply
			reply.TaskType = Reduce
			reply.ReduceTaskNumber = i
			reply.NMap = m.nMap

			// Update data structure
			m.reduceTasks[i].State = InProgress
			m.reduceTasks[i].StartTime = time.Now()

			return nil
		}
	}

	if uncompleted {
		reply.TaskType = Wait
		return nil
	}

	m.reduceDone = true
	reply.TaskType = Exit
	return nil
}

func (m *Master) TaskDone(args *TaskDoneArgs, reply *TaskDoneReply) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Always ok to update the state
	if args.TaskType == Map {
		if m.mapTasks[args.TaskNumber].State == Completed {
			return nil
		}

		m.mapTasks[args.TaskNumber].State = Completed

		// Atomic rename
		for i := 0; i < m.nReduce; i++ {
			name := fmt.Sprintf("mr-%v-%v", args.TaskNumber, i)
			os.Rename(args.FileNames[i], name)
		}
	} else {
		if m.reduceTasks[args.TaskNumber].State == Completed {
			return nil
		}

		m.reduceTasks[args.TaskNumber].State = Completed

		// Atomic rename
		name := fmt.Sprintf("mr-out-%v", args.TaskNumber)
		os.Rename(args.FileNames[0], name)
	}

	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.mapDone && m.reduceDone
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	m.nMap = len(files)
	m.nReduce = nReduce
	m.mapDone = false
	m.reduceDone = false

	m.mapTasks = make([]Task, m.nMap)
	for i := 0; i < len(files); i++ {
		m.mapTasks[i] = Task{
			TaskType: Map,
			State:    Idle,
			FileName: files[i],
		}
	}

	m.reduceTasks = make([]Task, m.nReduce)
	for i := 0; i < m.nReduce; i++ {
		m.reduceTasks[i] = Task{
			TaskType: Reduce,
			State:    Idle,
		}
	}

	m.server()
	return &m
}
