package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

var mapf func(string, string) []KeyValue
var reducef func(string, []string) string

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapfunc func(string, string) []KeyValue,
	reducefunc func(string, []string) string) {

	mapf = mapfunc
	reducef = reducefunc

	for {
		reply, ok := requestTask()
		if !ok {
			return
		}
		// if reply.TaskType == Map {
		// 	log.Printf("RequestTaskReply: {%s, %v, %s}\n", reply.TaskType, reply.MapTaskNumber, reply.FileName)
		// } else if reply.TaskType == Reduce {
		// 	log.Printf("RequestTaskReply: {%s, %v}\n", reply.TaskType, reply.ReduceTaskNumber)
		// } else {
		// 	log.Printf("RequestTaskReply: {%s}\n", reply.TaskType)
		// }

		switch reply.TaskType {
		case Map:
			handleMapTask(reply)
		case Reduce:
			handleReduceTask(reply)
		case Wait:
			time.Sleep(time.Second)
		case Exit:
			return
		default:
			log.Fatalf("Unrecognized task type!")
		}
	}
}

func requestTask() (RequestTaskReply, bool) {
	args := RequestTaskArgs{}
	reply := RequestTaskReply{}
	ok := call("Master.GiveTask", &args, &reply)
	return reply, ok
}

func handleReduceTask(reply RequestTaskReply) {
	// Read from intermediate files
	intermediate := []KeyValue{}
	for i := 0; i < reply.NMap; i++ {
		fileName := fmt.Sprintf("mr-%d-%d", i, reply.ReduceTaskNumber)
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("Cannot open %s", fileName)
		}

		decoder := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := decoder.Decode(&kv); err != nil {
				break
			} else {
			}
			intermediate = append(intermediate, kv)
		}
	}

	// Sort intermediate KVs
	sort.Sort(ByKey(intermediate))

	// Write to output file
	tempOutputFileName := fmt.Sprintf("mr-out-%d-*", reply.ReduceTaskNumber)
	tempOutputFile, err := os.Create(tempOutputFileName)
	if err != nil {
		log.Fatalf("Cannot create %s", tempOutputFileName)
	}

	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)
		fmt.Fprintf(tempOutputFile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	tempOutputFile.Close()
	notifyTaskDone(Reduce, reply.ReduceTaskNumber, []string{tempOutputFileName})
}

func handleMapTask(reply RequestTaskReply) {
	// Read file content
	file, err := os.Open(reply.FileName)
	if err != nil {
		log.Fatalf("cannot open %v", reply.FileName)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", reply.FileName)
	}
	file.Close()

	// Call Map function
	kva := mapf(reply.FileName, string(content))

	// Pre-open all needed files to save time
	pwd, _ := os.Getwd()
	files := make([](*os.File), reply.NReduce)
	fileNames := make([]string, reply.NReduce)
	for i := 0; i < len(files); i++ {
		tempInterFileName := fmt.Sprintf("mr-%v-%v-*", reply.MapTaskNumber, i)
		tempInterFile, err := ioutil.TempFile(pwd, tempInterFileName)
		if err != nil {
			log.Fatalf("Cannot create temp inter file: %v\n", tempInterFileName)
		}
		files[i] = tempInterFile
		fileNames[i] = tempInterFile.Name()
	}

	// Store intermediate result
	for _, kv := range kva {
		reduceIndex := ihash(kv.Key) % reply.NReduce
		interFile := files[reduceIndex]

		// Write intermediate keys to file
		encoder := json.NewEncoder(interFile)
		err = encoder.Encode(kv)
		if err != nil {
			log.Fatalf("Cannot encode %v: %v", kv, err)
		}
	}

	// Close all intermediate files
	for i := 0; i < len(files); i++ {
		err := files[i].Close()
		if err != nil {
			log.Fatalf("Cannot close %v: %v", files[i].Name(), err)
		}
	}

	// Notify master, master would rename temporary files
	notifyTaskDone(Map, reply.MapTaskNumber, fileNames)
}

func notifyTaskDone(taskType string, mapTaskNumber int, fileNames []string) {
	args := TaskDoneArgs{taskType, mapTaskNumber, fileNames}
	reply := TaskDoneReply{}
	ok := call("Master.TaskDone", &args, &reply)
	if !ok {
		log.Println("Call Master.TaskDone returns FALSE")
	}
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
