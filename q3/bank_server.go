package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/vishnu1910/samplego/q3/protofiles/paymentpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// bankServer implements the Bank service.
type bankServer struct {
	pb.UnimplementedBankServer
	name string
	// In a real implementation, include account state, persistence, etc.
}

func newBankServer(name string) *bankServer {
	return &bankServer{name: name}
}

func (s *bankServer) VoteCommit(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	log.Printf("[%s] Received VoteRequest for txn %s, amount %.2f", s.name, req.TransactionId, req.Amount)
	// For demo, always vote yes. In a real bank, check for sufficient funds.
	return &pb.VoteResponse{Vote: true, Message: "OK"}, nil
}

func (s *bankServer) Commit(ctx context.Context, req *pb.CommitRequest) (*pb.CommitResponse, error) {
	log.Printf("[%s] Commit transaction %s", s.name, req.TransactionId)
	// Update balance here.
	return &pb.CommitResponse{Success: true, Message: "Committed"}, nil
}

func (s *bankServer) Rollback(ctx context.Context, req *pb.RollbackRequest) (*pb.RollbackResponse, error) {
	log.Printf("[%s] Rollback transaction %s", s.name, req.TransactionId)
	// Revert any tentative changes.
	return &pb.RollbackResponse{Success: true, Message: "Rolled back"}, nil
}

func main() {
	name := flag.String("name", "BankA", "Unique name for the bank server")
	port := flag.Int("port", 50051, "The server port for the bank server")
	flag.Parse()

	// Load TLS credentials.
	creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		log.Fatalf("[%s] Failed to load TLS keys: %v", *name, err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
	}
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterBankServer(grpcServer, newBankServer(*name))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("[%s] Failed to listen: %v", *name, err)
	}
	log.Printf("[%s] Bank server listening at %v", *name, lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[%s] Failed to serve: %v", *name, err)
	}
}
