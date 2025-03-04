package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	pb "github.com/vishnu1910/samplego/q1/protofiles/lbpb"
)

// serverRecord holds registered server details.
type serverRecord struct {
	id      string
	address string
	port    int32
	load    float32
}

// lbServer implements the LoadBalancer service.
type lbServer struct {
	pb.UnimplementedLoadBalancerServer
	mu       sync.Mutex
	servers  []serverRecord
	rrIndex  int
	lbPolicy string
}

func (s *lbServer) RegisterServer(ctx context.Context, info *pb.ServerInfo) (*pb.RegistrationAck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec := serverRecord{
		id:      info.Id,
		address: info.Address,
		port:    info.Port,
		load:    0.0, // initial load
	}
	s.servers = append(s.servers, rec)
	log.Printf("Registered server: %s at %s:%d", info.Id, info.Address, info.Port)
	return &pb.RegistrationAck{Message: "Server registered successfully"}, nil
}

func (s *lbServer) ReportLoad(ctx context.Context, info *pb.LoadInfo) (*pb.StatusAck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, rec := range s.servers {
		if rec.id == info.Id {
			s.servers[i].load = info.Load
			log.Printf("Updated load for server %s: %f", rec.id, info.Load)
			return &pb.StatusAck{Message: "Load updated"}, nil
		}
	}
	return &pb.StatusAck{Message: "Server not found"}, nil
}

func (s *lbServer) GetBestServer(ctx context.Context, req *pb.ClientRequest) (*pb.ServerInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	var chosen serverRecord
	switch s.lbPolicy {
	case "pickfirst":
		chosen = s.servers[0]
	case "roundrobin":
		chosen = s.servers[s.rrIndex]
		s.rrIndex = (s.rrIndex + 1) % len(s.servers)
	case "leastload":
		chosen = s.servers[0]
		for _, rec := range s.servers {
			if rec.load < chosen.load {
				chosen = rec
			}
		}
	default:
		chosen = s.servers[0]
	}

	log.Printf("Client %s assigned to server %s (%s:%d) using policy %s", req.ClientId, chosen.id, chosen.address, chosen.port, s.lbPolicy)

	return &pb.ServerInfo{
		Id:      chosen.id,
		Address: chosen.address,
		Port:    chosen.port,
	}, nil
}

func main() {
	policy := flag.String("policy", "pickfirst", "Load balancing policy: pickfirst, roundrobin, leastload")
	port := flag.Int("port", 50051, "Port on which the LB server listens")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	lbSrv := &lbServer{
		servers:  make([]serverRecord, 0),
		lbPolicy: *policy,
	}
	pb.RegisterLoadBalancerServer(grpcServer, lbSrv)
	log.Printf("Load Balancer server listening on port %d with policy %s", *port, *policy)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
