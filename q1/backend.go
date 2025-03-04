package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	pb "github.com/vishnu1910/samplego/q1/protofiles/lbpb"
)

// backendServer implements the Backend service.
type backendServer struct {
	pb.UnimplementedBackendServer
	id string
}

func (s *backendServer) ProcessTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	// Simulate task processing.
	time.Sleep(100 * time.Millisecond)
	result := fmt.Sprintf("Processed task '%s' by server %s", req.Task, s.id)
	log.Printf("Processed task from client %s: %s", req.ClientId, result)
	return &pb.TaskResponse{Result: result}, nil
}

// registerWithLB registers the backend server with the Load Balancer.
func registerWithLB(lbAddr, id, address string, port int) {
	conn, err := grpc.Dial(lbAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to LB: %v", err)
	}
	defer conn.Close()
	client := pb.NewLoadBalancerClient(conn)
	req := &pb.ServerInfo{
		Id:      id,
		Address: address,
		Port:    int32(port),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.RegisterServer(ctx, req)
	if err != nil {
		log.Fatalf("Failed to register with LB: %v", err)
	}
	log.Printf("Registration response: %s", resp.Message)
}

// reportLoadPeriodically simulates and reports the server load to the LB.
func reportLoadPeriodically(lbAddr, id string) {
	conn, err := grpc.Dial(lbAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to LB for reporting load: %v", err)
	}
	defer conn.Close()
	client := pb.NewLoadBalancerClient(conn)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		loadValue := rand.Float32() * 100 // simulate a load value between 0 and 100
		req := &pb.LoadInfo{
			Id:   id,
			Load: loadValue,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := client.ReportLoad(ctx, req)
		cancel()
		if err != nil {
			log.Printf("Error reporting load: %v", err)
		}
	}
}

func main() {
	var (
		lbAddr = flag.String("lb", "localhost:50051", "Address of the Load Balancer")
		port   = flag.Int("port", 50052, "Port for the backend server")
		id     = flag.String("id", "", "Unique ID for the backend server")
	)
	flag.Parse()

	if *id == "" {
		*id = "backend-" + strconv.Itoa(rand.Intn(1000))
	}
	address := "localhost" // assuming backend runs on localhost

	// Register with the Load Balancer.
	registerWithLB(*lbAddr, *id, address, *port)

	// Start a goroutine to report load periodically.
	go reportLoadPeriodically(*lbAddr, *id)

	// Start the gRPC server for the backend service.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterBackendServer(grpcServer, &backendServer{id: *id})
	log.Printf("Backend server %s listening on port %d", *id, *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
