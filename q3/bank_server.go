// bank_server.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/vishnu1910/samplego/q3/protofiles/paymentpb"
)

type bankServer struct {
	paymentpb.UnimplementedBankServiceServer
	// For simulation, each bank has a limit.
	approvalLimit float32
	// In a real system, you’d have a map of account balances etc.
}

func (bs *bankServer) PreparePayment(ctx context.Context, req *paymentpb.PrepareRequest) (*paymentpb.PrepareResponse, error) {
	// For simulation, if amount is less than the approvalLimit, vote yes.
	if req.Amount <= bs.approvalLimit {
		log.Printf("Bank approving transaction %s for amount %.2f", req.TransactionId, req.Amount)
		return &paymentpb.PrepareResponse{Vote: true, Message: "Approved"}, nil
	}
	log.Printf("Bank rejecting transaction %s for amount %.2f", req.TransactionId, req.Amount)
	return &paymentpb.PrepareResponse{Vote: false, Message: "Amount exceeds limit"}, nil
}

func (bs *bankServer) CommitPayment(ctx context.Context, req *paymentpb.CommitRequest) (*paymentpb.CommitResponse, error) {
	log.Printf("Bank committing transaction %s", req.TransactionId)
	// Simulate processing delay.
	time.Sleep(500 * time.Millisecond)
	return &paymentpb.CommitResponse{Committed: true, Message: "Committed"}, nil
}

func (bs *bankServer) AbortPayment(ctx context.Context, req *paymentpb.AbortRequest) (*paymentpb.AbortResponse, error) {
	log.Printf("Bank aborting transaction %s", req.TransactionId)
	return &paymentpb.AbortResponse{Aborted: true, Message: "Aborted"}, nil
}

func main() {
	port := flag.Int("port", 50061, "Port for Bank Server")
	limit := flag.Float64("limit", 10000, "Approval limit for transactions")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Bank Server failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	paymentpb.RegisterBankServiceServer(grpcServer, &bankServer{
		approvalLimit: float32(*limit),
	})
	log.Printf("Bank Server listening on port %d with approval limit %.2f", *port, *limit)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Bank Server failed to serve: %v", err)
	}
}
