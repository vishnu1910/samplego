package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/vishnu1910/samplego/q2/protofiles/mrpb"
	"google.golang.org/grpc"

	// Import the generated package; ensure the import path matches your setup.
)

// Define task statuses.
type mapTaskStatus int

const (
	pending mapTaskStatus = iota
	inProgress
	completed
)

type MapTaskInfo struct {
	task              *mrpb.MapTask
	status            mapTaskStatus
	intermediateFiles []string // files produced by the map task
}

type reduceTaskStatus int

const (
	rPending reduceTaskStatus = iota
	rInProgress
	rCompleted
)

type ReduceTaskInfo struct {
	task       *mrpb.ReduceTask
	status     reduceTaskStatus
	outputFile string
}

// Master implements the MasterService.
type Master struct {
	mrpb.UnimplementedMasterServiceServer
	mu          sync.Mutex
	mapTasks    []*MapTaskInfo
	reduceTasks []*ReduceTaskInfo
	numMap      int
	numReduce   int
	jobType     mrpb.JobType

	// Flag indicating all map tasks are done.
	mapTasksDone bool
}

func (m *Master) RequestMapTask(ctx context.Context, req *mrpb.MapTaskRequest) (*mrpb.MapTaskResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If all map tasks have been completed, return Assigned=false.
	if m.mapTasksDone {
		return &mrpb.MapTaskResponse{Assigned: false}, nil
	}

	for _, taskInfo := range m.mapTasks {
		if taskInfo.status == pending {
			taskInfo.status = inProgress
			log.Printf("Assigning Map Task %d to worker %s", taskInfo.task.TaskId, req.Worker.WorkerId)
			return &mrpb.MapTaskResponse{Assigned: true, Task: taskInfo.task}, nil
		}
	}
	// No task is currently available.
	return &mrpb.MapTaskResponse{Assigned: false}, nil
}

func (m *Master) ReportMapTask(ctx context.Context, req *mrpb.ReportMapTaskRequest) (*mrpb.ReportMapTaskResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, taskInfo := range m.mapTasks {
		if taskInfo.task.TaskId == req.TaskId && taskInfo.status == inProgress {
			taskInfo.status = completed
			taskInfo.intermediateFiles = req.IntermediateFiles
			log.Printf("Map Task %d completed by worker %s", req.TaskId, req.WorkerId)
			break
		}
	}
	// Check if all map tasks are completed.
	allDone := true
	for _, taskInfo := range m.mapTasks {
		if taskInfo.status != completed {
			allDone = false
			break
		}
	}
	if allDone {
		m.mapTasksDone = true
		// Prepare reduce tasks.
		reduceTasks := make([]*ReduceTaskInfo, m.numReduce)
		for i := 0; i < m.numReduce; i++ {
			reduceTask := &mrpb.ReduceTask{
				TaskId:            int32(i), // using reduce index as task id
				ReduceIndex:       int32(i),
				IntermediateFiles: []string{},
				JobType:           m.jobType,
			}
			reduceTasks[i] = &ReduceTaskInfo{
				task:   reduceTask,
				status: rPending,
			}
		}
		// For each map task, assume it produced exactly numReduce intermediate files.
		for _, mt := range m.mapTasks {
			for i, file := range mt.intermediateFiles {
				reduceTasks[i].task.IntermediateFiles = append(reduceTasks[i].task.IntermediateFiles, file)
			}
		}
		m.reduceTasks = reduceTasks
		log.Printf("All Map tasks completed. Prepared %d Reduce tasks.", m.numReduce)
	}
	return &mrpb.ReportMapTaskResponse{Status: "OK"}, nil
}

func (m *Master) RequestReduceTask(ctx context.Context, req *mrpb.ReduceTaskRequest) (*mrpb.ReduceTaskResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Only start assigning reduce tasks once map tasks are done.
	if !m.mapTasksDone {
		return &mrpb.ReduceTaskResponse{Assigned: false}, nil
	}

	for _, rt := range m.reduceTasks {
		if rt.status == rPending {
			rt.status = rInProgress
			log.Printf("Assigning Reduce Task %d (partition %d) to worker %s", rt.task.TaskId, rt.task.ReduceIndex, req.Worker.WorkerId)
			return &mrpb.ReduceTaskResponse{Assigned: true, Task: rt.task}, nil
		}
	}
	return &mrpb.ReduceTaskResponse{Assigned: false}, nil
}

func (m *Master) ReportReduceTask(ctx context.Context, req *mrpb.ReportReduceTaskRequest) (*mrpb.ReportReduceTaskResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rt := range m.reduceTasks {
		if rt.task.TaskId == req.TaskId && rt.status == rInProgress {
			rt.status = rCompleted
			rt.outputFile = req.OutputFile
			log.Printf("Reduce Task %d completed by worker %s, output: %s", req.TaskId, req.WorkerId, req.OutputFile)
			break
		}
	}
	return &mrpb.ReportReduceTaskResponse{Status: "OK"}, nil
}

func main() {
	var (
		port     = flag.Int("port", 50060, "Port for Master")
		numReduce = flag.Int("nreduce", 3, "Number of reduce tasks")
		jobTypeStr = flag.String("job", "wordcount", "Job type: wordcount or invertedindex")
	)
	flag.Parse()

	// Input files are specified as remaining command-line arguments.
	inputFiles := flag.Args()
	if len(inputFiles) == 0 {
		log.Fatalf("No input files provided")
	}
	numMap := len(inputFiles)

	var jobType mrpb.JobType
	if *jobTypeStr == "wordcount" {
		jobType = mrpb.JobType_WORDCOUNT
	} else {
		jobType = mrpb.JobType_INVERTEDINDEX
	}

	master := &Master{
		numMap:    numMap,
		numReduce: *numReduce,
		jobType:   jobType,
	}
	// Create a map task for each input file.
	master.mapTasks = make([]*MapTaskInfo, numMap)
	for i, file := range inputFiles {
		task := &mrpb.MapTask{
			TaskId:    int32(i),
			InputFile: file,
			JobType:   jobType,
			NumReduce: int32(*numReduce),
		}
		master.mapTasks[i] = &MapTaskInfo{
			task:   task,
			status: pending,
		}
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	mrpb.RegisterMasterServiceServer(grpcServer, master)
	log.Printf("Master server started on port %d", *port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

