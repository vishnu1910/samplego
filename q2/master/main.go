package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/vishnu1910/samplego/q2/mapreduce/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TaskStatus represents the state of a task.
type TaskStatus int

const (
	Idle TaskStatus = iota
	InProgress
	Completed
)

// Task holds details for a map or reduce task.
type Task struct {
	ID        int
	TaskType  pb.TaskType
	InputFile string   // For map tasks.
	Status    TaskStatus
	// For reduce tasks, we compute intermediate file names later.
}

type MasterServer struct {
	pb.UnimplementedMasterServiceServer

	mu          sync.Mutex
	phase       string // "map", "reduce", or "done"
	mapTasks    []Task
	reduceTasks []Task

	numReduce int
	// Total number of map tasks (used to compute intermediate file names).
	numMap int
}

func NewMasterServer(jobType string, numMap, numReduce int, inputFiles []string) *MasterServer {
	// Initialize map tasks.
	mapTasks := make([]Task, 0)
	// For simplicity, we assume one map task per input file.
	// (If numMap is less than len(inputFiles), you could combine files.)
	for i, file := range inputFiles {
		mapTasks = append(mapTasks, Task{
			ID:        i,
			TaskType:  pb.TaskType_MAP,
			InputFile: file,
			Status:    Idle,
		})
	}

	// Create reduce tasks (their input will be computed from intermediate file names).
	reduceTasks := make([]Task, numReduce)
	for i := 0; i < numReduce; i++ {
		reduceTasks[i] = Task{
			ID:       i,
			TaskType: pb.TaskType_REDUCE,
			Status:   Idle,
		}
	}

	return &MasterServer{
		phase:       "map",
		mapTasks:    mapTasks,
		reduceTasks: reduceTasks,
		numReduce:   numReduce,
		numMap:      len(mapTasks),
	}
}

// RequestTask is called by a worker to request a task.
func (m *MasterServer) RequestTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskAssignment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// In the map phase: assign idle map tasks.
	if m.phase == "map" {
		for i, task := range m.mapTasks {
			if task.Status == Idle {
				m.mapTasks[i].Status = InProgress
				log.Printf("Assigning MAP task %d (%s) to worker %s", task.ID, task.InputFile, req.WorkerId)
				return &pb.TaskAssignment{
					TaskType:  pb.TaskType_MAP,
					TaskId:    int32(task.ID),
					InputFile: task.InputFile,
				}, nil
			}
		}
		// If all map tasks are in-progress but not yet complete, ask worker to wait.
		allDone := true
		for _, task := range m.mapTasks {
			if task.Status != Completed {
				allDone = false
				break
			}
		}
		if allDone {
			// Move to reduce phase.
			m.phase = "reduce"
			log.Println("All map tasks completed. Moving to reduce phase.")
		} else {
			return &pb.TaskAssignment{TaskType: pb.TaskType_WAIT}, nil
		}
	}

	// In the reduce phase: assign idle reduce tasks.
	if m.phase == "reduce" {
		for i, task := range m.reduceTasks {
			if task.Status == Idle {
				m.reduceTasks[i].Status = InProgress
				// Compute intermediate file names for this reduce task.
				interFiles := make([]string, 0, m.numMap)
				for j := 0; j < m.numMap; j++ {
					// Convention: "mr-mapID-reduceID.txt"
					filename := fmt.Sprintf("mr-%d-%d.txt", j, task.ID)
					interFiles = append(interFiles, filename)
				}
				log.Printf("Assigning REDUCE task %d to worker %s", task.ID, req.WorkerId)
				return &pb.TaskAssignment{
					TaskType:          pb.TaskType_REDUCE,
					TaskId:            int32(task.ID),
					IntermediateFiles: interFiles,
				}, nil
			}
		}
		// Check if all reduce tasks are done.
		allDone := true
		for _, task := range m.reduceTasks {
			if task.Status != Completed {
				allDone = false
				break
			}
		}
		if allDone {
			m.phase = "done"
			log.Println("All reduce tasks completed. Job finished.")
		} else {
			return &pb.TaskAssignment{TaskType: pb.TaskType_WAIT}, nil
		}
	}

	// If phase is done, tell workers to exit.
	return &pb.TaskAssignment{TaskType: pb.TaskType_EXIT}, nil
}

// ReportTask is called by a worker to report task completion.
func (m *MasterServer) ReportTask(ctx context.Context, result *pb.TaskResult) (*emptypb.Empty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.phase == "map" {
		// Mark map task as complete.
		for i, task := range m.mapTasks {
			if task.ID == int(result.TaskId) {
				m.mapTasks[i].Status = Completed
				log.Printf("MAP task %d completed by worker %s", task.ID, result.WorkerId)
				break
			}
		}
	} else if m.phase == "reduce" {
		// Mark reduce task as complete.
		for i, task := range m.reduceTasks {
			if task.ID == int(result.TaskId) {
				m.reduceTasks[i].Status = Completed
				log.Printf("REDUCE task %d completed by worker %s, output: %s", task.ID, result.WorkerId, result.OutputFile)
				break
			}
		}
	}
	return &emptypb.Empty{}, nil
}

func main() {
	// Command-line flags for the master.
	jobType := flag.String("job", "wc", "Job type: 'wc' for word count, 'ii' for inverted index")
	numReduce := flag.Int("nReduce", 3, "Number of reduce tasks")
	flag.Parse()

	inputFiles := flag.Args()
	if len(inputFiles) == 0 {
		fmt.Println("Usage: master -job=<wc|ii> -nReduce=<num> <inputfiles...>")
		os.Exit(1)
	}

	// For this assignment, the number of map tasks equals the number of input files.
	numMap := len(inputFiles)
	master := NewMasterServer(*jobType, numMap, *numReduce, inputFiles)

	// Start the gRPC server.
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMasterServiceServer(grpcServer, master)

	// Run the server in a separate goroutine.
	go func() {
		log.Println("Master server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Periodically check for job completion.
	for {
		time.Sleep(2 * time.Second)
		master.mu.Lock()
		if master.phase == "done" {
			master.mu.Unlock()
			break
		}
		master.mu.Unlock()
	}
	log.Println("Job completed. Shutting down master server.")
	grpcServer.GracefulStop()
}

