package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"mrpb"
)

// A simple hash function to partition keys.
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// mapF implements the map function for word count.
// It reads the file content, splits it into words, and counts occurrences.
func mapF(filename string, contents string) map[string]int {
	results := make(map[string]int)
	words := strings.Fields(contents)
	for _, w := range words {
		results[w]++
	}
	return results
}

// reduceF implements the reduce function for word count.
// It sums up counts for each word.
func reduceF(intermediate map[string][]int) map[string]int {
	final := make(map[string]int)
	for word, counts := range intermediate {
		sum := 0
		for _, c := range counts {
			sum += c
		}
		final[word] = sum
	}
	return final
}

// processMapTask handles the map task assigned to this worker.
func processMapTask(task *mrpb.MapTask, workerID string, masterAddr string) {
	log.Printf("Worker %s processing Map Task %d", workerID, task.TaskId)
	// Read input file.
	contentBytes, err := ioutil.ReadFile(task.InputFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", task.InputFile, err)
	}
	content := string(contentBytes)
	// For this example, we implement word count.
	kvs := mapF(task.InputFile, content)
	numReduce := int(task.NumReduce)
	// Partition the output by computing hash(word) mod numReduce.
	partitions := make(map[int][]string)
	for word, count := range kvs {
		partition := int(hash(word)) % numReduce
		line := fmt.Sprintf("%s %d", word, count)
		partitions[partition] = append(partitions[partition], line)
	}
	// Write each partition to an intermediate file named "mr-<MapTaskID>-<reduceIndex>".
	intermediateFiles := make([]string, numReduce)
	for i := 0; i < numReduce; i++ {
		filename := fmt.Sprintf("mr-%d-%d", task.TaskId, i)
		f, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Error creating intermediate file %s: %v", filename, err)
		}
		for _, line := range partitions[i] {
			fmt.Fprintln(f, line)
		}
		f.Close()
		intermediateFiles[i] = filename
	}
	// Report completion to the Master.
	conn, err := grpc.Dial(masterAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Master: %v", err)
	}
	defer conn.Close()
	client := mrpb.NewMasterServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &mrpb.ReportMapTaskRequest{
		WorkerId:         workerID,
		TaskId:           task.TaskId,
		IntermediateFiles: intermediateFiles,
	}
	resp, err := client.ReportMapTask(ctx, req)
	if err != nil {
		log.Fatalf("Error reporting map task: %v", err)
	}
	log.Printf("Map Task %d reported completion: %s", task.TaskId, resp.Status)
}

// processReduceTask handles the reduce task assigned to this worker.
func processReduceTask(task *mrpb.ReduceTask, workerID string, masterAddr string) {
	log.Printf("Worker %s processing Reduce Task %d (partition %d)", workerID, task.TaskId, task.ReduceIndex)
	// For each intermediate file, read key-value pairs and aggregate counts.
	intermediate := make(map[string][]int)
	for _, filename := range task.IntermediateFiles {
		f, err := os.Open(filename)
		if err != nil {
			log.Printf("Error opening intermediate file %s: %v", filename, err)
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.Fields(line)
			if len(parts) != 2 {
				continue
			}
			word := parts[0]
			count, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}
			intermediate[word] = append(intermediate[word], count)
		}
		f.Close()
	}
	// Apply reduce function.
	finalResult := reduceF(intermediate)
	// Write final output to file "mr-out-<reduceIndex>".
	outFilename := fmt.Sprintf("mr-out-%d", task.ReduceIndex)
	outFile, err := os.Create(outFilename)
	if err != nil {
		log.Fatalf("Error creating output file %s: %v", outFilename, err)
	}
	for word, count := range finalResult {
		fmt.Fprintf(outFile, "%s %d\n", word, count)
	}
	outFile.Close()
	// Report completion to the Master.
	conn, err := grpc.Dial(masterAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Master: %v", err)
	}
	defer conn.Close()
	client := mrpb.NewMasterServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &mrpb.ReportReduceTaskRequest{
		WorkerId:   workerID,
		TaskId:     task.TaskId,
		OutputFile: outFilename,
	}
	resp, err := client.ReportReduceTask(ctx, req)
	if err != nil {
		log.Fatalf("Error reporting reduce task: %v", err)
	}
	log.Printf("Reduce Task %d reported completion: %s", task.TaskId, resp.Status)
}

func main() {
	var (
		masterAddr = flag.String("master", "localhost:50060", "Address of the Master")
		workerID   = flag.String("id", "", "Unique Worker ID")
	)
	flag.Parse()
	if *workerID == "" {
		*workerID = fmt.Sprintf("worker-%d", time.Now().UnixNano())
	}
	log.Printf("Starting Worker %s", *workerID)
	// Worker continuously requests tasks from the Master.
	for {
		// First try to get a map task.
		conn, err := grpc.Dial(*masterAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect to Master: %v", err)
		}
		client := mrpb.NewMasterServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		mapResp, err := client.RequestMapTask(ctx, &mrpb.MapTaskRequest{
			Worker: &mrpb.WorkerInfo{WorkerId: *workerID},
		})
		cancel()
		conn.Close()
		if err != nil {
			log.Printf("Error requesting map task: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if mapResp.Assigned {
			processMapTask(mapResp.Task, *workerID, *masterAddr)
			continue
		} else {
			// If no map task is assigned, try to request a reduce task.
			conn, err = grpc.Dial(*masterAddr, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("Failed to connect to Master: %v", err)
			}
			client = mrpb.NewMasterServiceClient(conn)
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			reduceResp, err := client.RequestReduceTask(ctx, &mrpb.ReduceTaskRequest{
				Worker: &mrpb.WorkerInfo{WorkerId: *workerID},
			})
			cancel()
			conn.Close()
			if err != nil {
				log.Printf("Error requesting reduce task: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			if reduceResp.Assigned {
				processReduceTask(reduceResp.Task, *workerID, *masterAddr)
				continue
			} else {
				// No task available at the moment; sleep briefly and try again.
				log.Printf("Worker %s: No tasks available, sleeping...", *workerID)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

