package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pb "github.com/vishnu1910/samplego/P2/protofiles/mapreduce"

	"google.golang.org/grpc"
)

// Global job type (set via command-line flag): "wc" or "ii".
var jobType string

// Hash function to determine reduce bucket.
func ihash(s string, nReduce int) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32() % uint32(nReduce))
}

// Map function for word count: reads input file and writes intermediate files.
func doMap(mapTaskID int, inputFile string, nReduce int) error {
	// Open the input file.
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("cannot open input file %s: %v", inputFile, err)
	}
	defer file.Close()

	// Read file contents.
	scanner := bufio.NewScanner(file)
	// Create writers for each reduce bucket.
	writers := make([]*os.File, nReduce)
	for i := 0; i < nReduce; i++ {
		intermediateFile := fmt.Sprintf("mr-%d-%d.txt", mapTaskID, i)
		f, err := os.Create(intermediateFile)
		if err != nil {
			return fmt.Errorf("cannot create intermediate file %s: %v", intermediateFile, err)
		}
		writers[i] = f
		defer f.Close()
	}

	// Process each line (for simplicity, treat the whole file as text).
	for scanner.Scan() {
		// Split line into words.
		words := strings.Fields(scanner.Text())
		for _, word := range words {
			bucket := ihash(word, nReduce)
			// For word count, output "word 1"
			line := fmt.Sprintf("%s %d\n", word, 1)
			if _, err := writers[bucket].WriteString(line); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Map function for inverted index: each line from input file is tokenized into words.
func doMapInvertedIndex(mapTaskID int, inputFile string, nReduce int) error {
	// Open the input file.
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("cannot open input file %s: %v", inputFile, err)
	}
	defer file.Close()

	// Create writers for each reduce bucket.
	writers := make([]*os.File, nReduce)
	for i := 0; i < nReduce; i++ {
		intermediateFile := fmt.Sprintf("mr-%d-%d.txt", mapTaskID, i)
		f, err := os.Create(intermediateFile)
		if err != nil {
			return fmt.Errorf("cannot create intermediate file %s: %v", intermediateFile, err)
		}
		writers[i] = f
		defer f.Close()
	}

	// For inverted index, use the input file name as the filename.
	baseName := filepath.Base(inputFile)

	// Process file line by line.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		for _, word := range words {
			bucket := ihash(word, nReduce)
			// For inverted index, output "word filename"
			line := fmt.Sprintf("%s %s\n", word, baseName)
			if _, err := writers[bucket].WriteString(line); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Reduce function for word count.
func doReduce(reduceTaskID int, nMap int) error {
	// key -> total count
	counts := make(map[string]int)

	// For each map task, read the corresponding intermediate file.
	for i := 0; i < nMap; i++ {
		intermediateFile := fmt.Sprintf("mr-%d-%d.txt", i, reduceTaskID)
		f, err := os.Open(intermediateFile)
		if err != nil {
			// If file doesn't exist, skip (could be empty).
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			// Each line is "word count"
			parts := strings.Fields(scanner.Text())
			if len(parts) != 2 {
				continue
			}
			word := parts[0]
			cnt, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}
			counts[word] += cnt
		}
		f.Close()
	}

	// Write reduce output to file.
	outFile := fmt.Sprintf("mr-out-%d.txt", reduceTaskID)
	fout, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("cannot create output file %s: %v", outFile, err)
	}
	defer fout.Close()
	for word, total := range counts {
		line := fmt.Sprintf("%s %d\n", word, total)
		fout.WriteString(line)
	}
	return nil
}

// Reduce function for inverted index.
func doReduceInvertedIndex(reduceTaskID int, nMap int) error {
	// key -> set of filenames (use map to deduplicate)
	index := make(map[string]map[string]bool)

	for i := 0; i < nMap; i++ {
		intermediateFile := fmt.Sprintf("mr-%d-%d.txt", i, reduceTaskID)
		f, err := os.Open(intermediateFile)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			// Each line: "word filename"
			parts := strings.Fields(scanner.Text())
			if len(parts) != 2 {
				continue
			}
			word := parts[0]
			filename := parts[1]
			if index[word] == nil {
				index[word] = make(map[string]bool)
			}
			index[word][filename] = true
		}
		f.Close()
	}

	// Write reduce output to file.
	outFile := fmt.Sprintf("mr-out-%d.txt", reduceTaskID)
	fout, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("cannot create output file %s: %v", outFile, err)
	}
	defer fout.Close()
	for word, fileSet := range index {
		// Build a comma-separated list of filenames.
		files := make([]string, 0)
		for fname := range fileSet {
			files = append(files, fname)
		}
		line := fmt.Sprintf("%s %s\n", word, strings.Join(files, ","))
		fout.WriteString(line)
	}
	return nil
}

func main() {
	// Command-line flags for the worker.
	masterAddr := flag.String("master", "localhost:50051", "Master server address")
	workerID := flag.String("worker", "worker-1", "Worker ID")
	jobTypeFlag := flag.String("job", "wc", "Job type: 'wc' for word count, 'ii' for inverted index")
	nReduce := flag.Int("nReduce", 3, "Number of reduce tasks (must match master's setting)")
	nMap := flag.Int("nMap", 0, "Number of map tasks (will be set by master; if 0, use reported value)")
	flag.Parse()
	jobType = *jobTypeFlag

	conn, err := grpc.Dial(*masterAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to master: %v", err)
	}
	defer conn.Close()
	client := pb.NewMasterServiceClient(conn)
	
	// Register worker before requesting tasks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = client.RegisterWorker(ctx, &pb.WorkerRegisterRequest{WorkerId: *workerID})
	cancel()
	if err != nil {
		log.Fatalf("Error registering worker %s: %v", *workerID, err)
	}
	log.Printf("Worker %s registered with master.", *workerID)

	for {
		// Request a task.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		task, err := client.RequestTask(ctx, &pb.TaskRequest{WorkerId: *workerID})
		cancel()
		if err != nil {
			log.Fatalf("Error requesting task: %v", err)
		}

		switch task.TaskType {
		case pb.TaskType_MAP:
			log.Printf("Worker %s received MAP task %d with input %s", *workerID, task.TaskId, task.InputFile)
			mapTaskID := int(task.TaskId)
			// Execute map function based on job type.
			var mapErr error
			if jobType == "wc" {
				mapErr = doMap(mapTaskID, task.InputFile, *nReduce)
			} else if jobType == "ii" {
				mapErr = doMapInvertedIndex(mapTaskID, task.InputFile, *nReduce)
			} else {
				log.Fatalf("Unknown job type: %s", jobType)
			}
			if mapErr != nil {
				log.Printf("Error in map task %d: %v", mapTaskID, mapErr)
			}
			// Report task completion.
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			_, err = client.ReportTask(ctx, &pb.TaskResult{
				WorkerId:  *workerID,
				TaskId:    task.TaskId,
				Success:   (mapErr == nil),
				OutputFile: "", // Not used for map tasks.
			})
			cancel()
			if err != nil {
				log.Printf("Error reporting map task %d: %v", mapTaskID, err)
			}

		case pb.TaskType_REDUCE:
			log.Printf("Worker %s received REDUCE task %d", *workerID, task.TaskId)
			reduceTaskID := int(task.TaskId)
			// For reduce tasks, the worker uses the number of map tasks.
			// Here we assume the worker can count how many map tasks there were
			// by the number of intermediate files (length of IntermediateFiles).
			if *nMap == 0 {
				*nMap = len(task.IntermediateFiles)
			}
			var redErr error
			if jobType == "wc" {
				redErr = doReduce(reduceTaskID, *nMap)
			} else if jobType == "ii" {
				redErr = doReduceInvertedIndex(reduceTaskID, *nMap)
			} else {
				log.Fatalf("Unknown job type: %s", jobType)
			}
			outFile := fmt.Sprintf("mr-out-%d.txt", reduceTaskID)
			if redErr != nil {
				log.Printf("Error in reduce task %d: %v", reduceTaskID, redErr)
			}
			// Report task completion.
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			_, err = client.ReportTask(ctx, &pb.TaskResult{
				WorkerId:  *workerID,
				TaskId:    task.TaskId,
				Success:   (redErr == nil),
				OutputFile: outFile,
			})
			cancel()
			if err != nil {
				log.Printf("Error reporting reduce task %d: %v", reduceTaskID, err)
			}

		case pb.TaskType_WAIT:
			// No task available now; wait a bit.
			log.Printf("Worker %s: no task available; waiting...", *workerID)
			time.Sleep(2 * time.Second)

		case pb.TaskType_EXIT:
			log.Printf("Worker %s: received EXIT task. Shutting down.", *workerID)
			os.Exit(0)
		}
	}
}

