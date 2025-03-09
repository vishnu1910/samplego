package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "github.com/vishnu1910/samplego/q1/protofiles/lbpb"
)

func main() {
	var (
		lbAddr   = flag.String("lb", "localhost:50051", "Address of the Load Balancer")
		clientId = flag.String("id", "client-1", "Client ID")
		task     = flag.String("task", "Sample Task", "Task to process")
	)
	flag.Parse()

	// Connect to the Load Balancer.
	lbConn, err := grpc.Dial(*lbAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to LB: %v", err)
	}
	defer lbConn.Close()
	lbClient := pb.NewLoadBalancerClient(lbConn)

	// Request the best server.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.ClientRequest{
		ClientId: *clientId,
		Task:     *task,
	}
	serverInfo, err := lbClient.GetBestServer(ctx, req)
	if err != nil {
		log.Fatalf("Error getting best server: %v", err)
	}
	backendAddr := fmt.Sprintf("%s:%d", serverInfo.Address, serverInfo.Port)
	log.Printf("Client %s got best server: %s at %s", *clientId, serverInfo.Id, backendAddr)

	// Connect to the selected backend server.
	backendConn, err := grpc.Dial(backendAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to backend server: %v", err)
	}
	defer backendConn.Close()
	backendClient := pb.NewBackendClient(backendConn)

	// Send the task request.
	taskReq := &pb.TaskRequest{
		ClientId: *clientId,
		Task:     *task,
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	taskResp, err := backendClient.ProcessTask(ctx2, taskReq)
	if err != nil {
		log.Fatalf("Error processing task: %v", err)
	}
	log.Printf("Task response: %s", taskResp.Result)
}
