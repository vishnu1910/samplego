package gateway

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/vishnu1910/samplego/q3/pb"

	"google.golang.org/grpc"
)

// PaymentGatewayServer implements the PaymentGateway service.
type PaymentGatewayServer struct {
	pb.UnimplementedPaymentGatewayServer
	// idempotencyMap stores processed idempotency keys and their response.
	idempotencyMap sync.Map
	// Dummy user accounts (client balances) stored in-memory.
	// In a production system, use persistent storage.
	accounts sync.Map // key: client_id, value: float64 (balance)

	// Mapping of bank names to their gRPC client connections.
	bankClients map[string]pb.BankServiceClient

	// Timeout for 2PC transactions.
	tpcTimeout time.Duration
}

// NewPaymentGatewayServer creates a new PaymentGatewayServer.
func NewPaymentGatewayServer(bankClients map[string]pb.BankServiceClient, tpcTimeout time.Duration) *PaymentGatewayServer {
	// Preload some dummy account balances.
	accounts := sync.Map{}
	accounts.Store("alice", 1000.0)
	accounts.Store("bob", 1000.0)
	return &PaymentGatewayServer{
		accounts:    accounts,
		bankClients: bankClients,
		tpcTimeout:  tpcTimeout,
	}
}

// RegisterClient registers a new client with the payment gateway.
func (s *PaymentGatewayServer) RegisterClient(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// For simplicity, we assume that the user is already authenticated by interceptors.
	clientID := req.Username // note: proto field "username" is used

	// In a real system, you might add the client to a persistent datastore.
	// Here, we add to our in-memory accounts if not already present.
	if _, ok := s.accounts.Load(clientID); !ok {
		s.accounts.Store(clientID, 1000.0) // default balance
	}
	log.Printf("[REGISTER] Client %s registered with bank %s, account %s", clientID, req.Bank, req.Account_number)
	return &pb.RegisterResponse{
		Success: true,
		Message: "Registration successful",
	}, nil
}

// ProcessPayment implements the 2PC payment process.
func (s *PaymentGatewayServer) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	// Check idempotency
	if res, exists := s.idempotencyMap.Load(req.Idempotency_key); exists {
		log.Printf("[IDEMPOTENCY] Returning stored result for key: %s", req.Idempotency_key)
		return res.(*pb.PaymentResponse), nil
	}

	// Start 2PC: The gateway is the coordinator.
	transactionID := fmt.Sprintf("txn-%d", time.Now().UnixNano())
	log.Printf("[2PC] Starting transaction %s", transactionID)

	// Create a context with timeout for the 2PC
	ctx2, cancel := context.WithTimeout(ctx, s.tpcTimeout)
	defer cancel()

	// Call Prepare on both bank servers (sender and receiver)
	senderClient, ok := s.bankClients[req.From_bank]
	if !ok {
		return nil, errors.New("sender bank not found")
	}
	receiverClient, ok := s.bankClients[req.To_bank]
	if !ok {
		return nil, errors.New("receiver bank not found")
	}

	// Prepare sender (deduct funds) and receiver (credit funds)
	senderPrepReq := &pb.PrepareRequest{
		Transaction_id: transactionID,
		Account_number: req.From_account,
		Amount:         req.Amount,
	}
	receiverPrepReq := &pb.PrepareRequest{
		Transaction_id: transactionID,
		Account_number: req.To_account,
		Amount:         req.Amount,
	}

	senderPrepChan := make(chan *pb.PrepareResponse, 1)
	receiverPrepChan := make(chan *pb.PrepareResponse, 1)
	errChan := make(chan error, 2)

	go func() {
		resp, err := senderClient.Prepare(ctx2, senderPrepReq)
		if err != nil {
			errChan <- fmt.Errorf("sender prepare error: %v", err)
			return
		}
		senderPrepChan <- resp
	}()
	go func() {
		resp, err := receiverClient.Prepare(ctx2, receiverPrepReq)
		if err != nil {
			errChan <- fmt.Errorf("receiver prepare error: %v", err)
			return
		}
		receiverPrepChan <- resp
	}()

	var senderPrepResp, receiverPrepResp *pb.PrepareResponse
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			// Abort transaction on error.
			s.abortTransaction(ctx2, transactionID, senderClient, receiverClient)
			s.storeIdempotency(req.Idempotency_key, false, err.Error())
			return &pb.PaymentResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		case resp := <-senderPrepChan:
			senderPrepResp = resp
		case resp := <-receiverPrepChan:
			receiverPrepResp = resp
		case <-ctx2.Done():
			s.abortTransaction(ctx2, transactionID, senderClient, receiverClient)
			msg := "transaction timeout during prepare phase"
			s.storeIdempotency(req.Idempotency_key, false, msg)
			return &pb.PaymentResponse{
				Success: false,
				Message: msg,
			}, nil
		}
	}

	// Check if both banks agreed to commit.
	if !senderPrepResp.CanCommit || !receiverPrepResp.CanCommit {
		s.abortTransaction(ctx2, transactionID, senderClient, receiverClient)
		msg := fmt.Sprintf("prepare phase failed: sender: %s, receiver: %s", senderPrepResp.Message, receiverPrepResp.Message)
		s.storeIdempotency(req.Idempotency_key, false, msg)
		return &pb.PaymentResponse{
			Success: false,
			Message: msg,
		}, nil
	}

	// Both banks can commit. Now call Commit on both.
	commitReq := &pb.CommitRequest{Transaction_id: transactionID}
	senderCommitChan := make(chan *pb.CommitResponse, 1)
	receiverCommitChan := make(chan *pb.CommitResponse, 1)
	go func() {
		resp, err := senderClient.Commit(ctx2, commitReq)
		if err != nil {
			errChan <- fmt.Errorf("sender commit error: %v", err)
			return
		}
		senderCommitChan <- resp
	}()
	go func() {
		resp, err := receiverClient.Commit(ctx2, commitReq)
		if err != nil {
			errChan <- fmt.Errorf("receiver commit error: %v", err)
			return
		}
		receiverCommitChan <- resp
	}()

	var senderCommitResp, receiverCommitResp *pb.CommitResponse
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			msg := fmt.Sprintf("commit phase error: %v", err)
			s.storeIdempotency(req.Idempotency_key, false, msg)
			return &pb.PaymentResponse{
				Success: false,
				Message: msg,
			}, nil
		case resp := <-senderCommitChan:
			senderCommitResp = resp
		case resp := <-receiverCommitChan:
			receiverCommitResp = resp
		case <-ctx2.Done():
			msg := "transaction timeout during commit phase"
			s.storeIdempotency(req.Idempotency_key, false, msg)
			return &pb.PaymentResponse{
				Success: false,
				Message: msg,
			}, nil
		}
	}

	// Check commit responses.
	if !senderCommitResp.Success || !receiverCommitResp.Success {
		msg := fmt.Sprintf("commit failed: sender: %s, receiver: %s", senderCommitResp.Message, receiverCommitResp.Message)
		s.storeIdempotency(req.Idempotency_key, false, msg)
		return &pb.PaymentResponse{
			Success: false,
			Message: msg,
		}, nil
	}

	// Update the sender's balance locally (for query purposes).
	s.updateLocalBalance(req.Client_id, req.Amount)

	msg := fmt.Sprintf("Payment of %f successful. Transaction ID: %s", req.Amount, transactionID)
	s.storeIdempotency(req.Idempotency_key, true, msg)
	return &pb.PaymentResponse{
		Success: true,
		Message: msg,
	}, nil
}

// abortTransaction sends abort requests to both banks.
func (s *PaymentGatewayServer) abortTransaction(ctx context.Context, transactionID string, senderClient, receiverClient pb.BankServiceClient) {
	abortReq := &pb.AbortRequest{Transaction_id: transactionID}
	_, err := senderClient.Abort(ctx, abortReq)
	if err != nil {
		log.Printf("[ABORT] Error aborting transaction at sender: %v", err)
	}
	_, err = receiverClient.Abort(ctx, abortReq)
	if err != nil {
		log.Printf("[ABORT] Error aborting transaction at receiver: %v", err)
	}
}

// updateLocalBalance updates the client's balance locally.
func (s *PaymentGatewayServer) updateLocalBalance(clientID string, amount float64) {
	// For simplicity, deduct the amount from the sender's balance.
	if val, ok := s.accounts.Load(clientID); ok {
		balance := val.(float64)
		s.accounts.Store(clientID, balance-amount)
	}
}

// storeIdempotency stores the result of a processed idempotency key.
func (s *PaymentGatewayServer) storeIdempotency(key string, success bool, message string) {
	s.idempotencyMap.Store(key, &pb.PaymentResponse{
		Success: success,
		Message: message,
	})
}

// GetBalance returns the balance for a given client.
func (s *PaymentGatewayServer) GetBalance(ctx context.Context, req *pb.BalanceRequest) (*pb.BalanceResponse, error) {
	val, ok := s.accounts.Load(req.Client_id)
	if !ok {
		return &pb.BalanceResponse{
			Balance: 0,
			Message: "Client not found",
		}, nil
	}
	return &pb.BalanceResponse{
		Balance: val.(float64),
		Message: "Balance retrieved successfully",
	}, nil
}

