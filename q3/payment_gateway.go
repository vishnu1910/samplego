package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/vishnu1910/samplego/q3/protofiles/paymentpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// idempotencyStore maintains processed transaction IDs.
type idempotencyStore struct {
	mu     sync.Mutex
	store  map[string]*pb.PaymentResponse
}

func newIdempotencyStore() *idempotencyStore {
	return &idempotencyStore{
		store: make(map[string]*pb.PaymentResponse),
	}
}

func (s *idempotencyStore) Get(key string) (*pb.PaymentResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	resp, exists := s.store[key]
	return resp, exists
}

func (s *idempotencyStore) Set(key string, resp *pb.PaymentResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = resp
}

// paymentGatewayServer implements the PaymentGateway service.
type paymentGatewayServer struct {
	pb.UnimplementedPaymentGatewayServer
	idempotency *idempotencyStore
	// In a real implementation, these would be configurable addresses or service clients.
	sendingBankAddr   string
	receivingBankAddr string
	timeout           time.Duration
}

func newPaymentGateway(sendingBankAddr, receivingBankAddr string, timeout time.Duration) *paymentGatewayServer {
	return &paymentGatewayServer{
		idempotency:       newIdempotencyStore(),
		sendingBankAddr:   sendingBankAddr,
		receivingBankAddr: receivingBankAddr,
		timeout:           timeout,
	}
}

// ProcessPayment implements the two-phase commit protocol.
func (s *paymentGatewayServer) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	// Check idempotency first.
	if resp, exists := s.idempotency.Get(req.IdempotencyKey); exists {
		log.Println("Idempotent request detected, returning cached response")
		return resp, nil
	}

	transactionID := fmt.Sprintf("txn-%d", time.Now().UnixNano())
	log.Printf("Processing transaction %s for amount %.2f", transactionID, req.Amount)

	// Set up a timeout for the 2PC process.
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Phase 1: Send VoteRequest to both banks.
	voteReq := &pb.VoteRequest{
		TransactionId: transactionID,
		ClientId:      req.ClientId,
		Amount:        req.Amount,
	}

	// For demo purposes, we assume both bank servers respond positively.
	sendingVote, err := sendVoteRequest(ctx, s.sendingBankAddr, voteReq)
	if err != nil || !sendingVote.Vote {
		return s.abortTransaction(req.IdempotencyKey, transactionID, "Sending bank aborted")
	}

	receivingVote, err := sendVoteRequest(ctx, s.receivingBankAddr, voteReq)
	if err != nil || !receivingVote.Vote {
		return s.abortTransaction(req.IdempotencyKey, transactionID, "Receiving bank aborted")
	}

	// Phase 2: Commit transaction on both banks.
	commitErr1 := sendCommitRequest(ctx, s.sendingBankAddr, transactionID)
	commitErr2 := sendCommitRequest(ctx, s.receivingBankAddr, transactionID)
	if commitErr1 != nil || commitErr2 != nil {
		// If commit fails, attempt rollback.
		sendRollbackRequest(ctx, s.sendingBankAddr, transactionID)
		sendRollbackRequest(ctx, s.receivingBankAddr, transactionID)
		return s.abortTransaction(req.IdempotencyKey, transactionID, "Commit failed; transaction rolled back")
	}

	resp := &pb.PaymentResponse{
		Success: true,
		Message: fmt.Sprintf("Transaction %s committed successfully", transactionID),
	}
	s.idempotency.Set(req.IdempotencyKey, resp)
	return resp, nil
}

func (s *paymentGatewayServer) abortTransaction(idempotencyKey, txnID, reason string) (*pb.PaymentResponse, error) {
	// Send rollback to both banks.
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	sendRollbackRequest(ctx, s.sendingBankAddr, txnID)
	sendRollbackRequest(ctx, s.receivingBankAddr, txnID)
	resp := &pb.PaymentResponse{
		Success: false,
		Message: fmt.Sprintf("Transaction %s aborted: %s", txnID, reason),
	}
	s.idempotency.Set(idempotencyKey, resp)
	return resp, nil
}

// Dummy functions to simulate remote calls to bank servers.
// In a complete solution these would be full gRPC client calls.
func sendVoteRequest(ctx context.Context, bankAddr string, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	// Simulate a bank vote by always voting yes.
	log.Printf("Sending VoteRequest to bank at %s for txn %s", bankAddr, req.TransactionId)
	return &pb.VoteResponse{Vote: true, Message: "Vote yes"}, nil
}

func sendCommitRequest(ctx context.Context, bankAddr, txnID string) error {
	log.Printf("Sending CommitRequest to bank at %s for txn %s", bankAddr, txnID)
	// Simulate successful commit.
	return nil
}

func sendRollbackRequest(ctx context.Context, bankAddr, txnID string) error {
	log.Printf("Sending RollbackRequest to bank at %s for txn %s", bankAddr, txnID)
	// Simulate successful rollback.
	return nil
}

// --- Interceptors for logging and authentication ---

// unaryLogger logs each incoming request.
func unaryLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	log.Printf("Started method: %s with req: %+v", info.FullMethod, req)
	resp, err := handler(ctx, req)
	log.Printf("Completed method: %s; Duration: %s; Error: %v; Response: %+v", info.FullMethod, time.Since(start), err, resp)
	return resp, err
}

// unaryAuth verifies a simple authentication token in metadata.
func unaryAuth(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// For demo purposes, check for a token in metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["authorization"]) == 0 || md["authorization"][0] != "Bearer secret-token" {
		return nil, fmt.Errorf("unauthenticated request")
	}
	return handler(ctx, req)
}

func main() {
	// For demo purposes, the bank addresses are fixed.
	sendingBankAddr := "localhost:50051"
	receivingBankAddr := "localhost:50052"
	port := flag.Int("port", 50050, "The server port for Payment Gateway")
	timeout := flag.Duration("timeout", 5*time.Second, "2PC timeout duration")
	flag.Parse()

	// Load TLS credentials.
	creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		log.Fatalf("Failed to load TLS keys: %v", err)
	}

	// Chain interceptors: first auth, then logging.
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(
			grpc.ChainUnaryInterceptor(unaryAuth, unaryLogger),
		),
	}
	grpcServer := grpc.NewServer(opts...)

	gateway := newPaymentGateway(sendingBankAddr, receivingBankAddr, *timeout)
	pb.RegisterPaymentGatewayServer(grpcServer, gateway)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Payment Gateway server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

