package bank

import (
	"context"
	"sync"

	pb "github.com/vishnu1910/samplego/P3/protofiles/pb"
)

// BankServer implements the BankService.
type BankServer struct {
	pb.UnimplementedBankServiceServer
	// accounts stores account balances keyed by account number.
	accounts sync.Map // key: account number, value: float64

	// For 2PC, we store pending transactions.
	pendingTransactions sync.Map // key: transactionId, value: float64 (amount)
}

// NewBankServer creates a new BankServer with dummy data.
func NewBankServer(initialAccounts map[string]float64) *BankServer {
	bs := &BankServer{}
	for acc, balance := range initialAccounts {
		bs.accounts.Store(acc, balance)
	}
	return bs
}

// Prepare checks if the transaction can be committed.
// For sender accounts, it checks if sufficient funds exist.
// For receiver accounts, it always returns true (assuming no upper bound).
func (s *BankServer) Prepare(ctx context.Context, req *pb.PrepareRequest) (*pb.PrepareResponse, error) {
	// Check if account exists.
	val, ok := s.accounts.Load(req.AccountNumber)
	if !ok {
		return &pb.PrepareResponse{
			CanCommit: false,
			Message:   "Account not found",
		}, nil
	}
	currentBalance := val.(float64)

	// If this is a sender, ensure funds are sufficient.
	if currentBalance >= req.Amount {
		// Reserve funds by storing the pending transaction.
		s.pendingTransactions.Store(req.TransactionId, req.Amount)
		return &pb.PrepareResponse{
			CanCommit: true,
			Message:   "Funds reserved",
		}, nil
	}

	// Otherwise, for a receiver, always allow.
	s.pendingTransactions.Store(req.TransactionId, req.Amount)
	return &pb.PrepareResponse{
		CanCommit: true,
		Message:   "Ready to credit",
	}, nil
}

// Commit finalizes the transaction by updating the balance.
func (s *BankServer) Commit(ctx context.Context, req *pb.CommitRequest) (*pb.CommitResponse, error) {
	val, ok := s.pendingTransactions.Load(req.TransactionId)
	if !ok {
		return &pb.CommitResponse{
			Success: false,
			Message: "No pending transaction found",
		}, nil
	}
	_ = val // For now, we are not using the amount to update the balance.
	// Remove the pending transaction.
	s.pendingTransactions.Delete(req.TransactionId)
	return &pb.CommitResponse{
		Success: true,
		Message: "Transaction committed",
	}, nil
}

// Abort cancels the transaction.
func (s *BankServer) Abort(ctx context.Context, req *pb.AbortRequest) (*pb.AbortResponse, error) {
	// Remove pending transaction.
	s.pendingTransactions.Delete(req.TransactionId)
	return &pb.AbortResponse{
		Success: true,
		Message: "Transaction aborted",
	}, nil
}

// GetAccount returns the account balance.
func (s *BankServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	val, ok := s.accounts.Load(req.AccountNumber)
	if !ok {
		return &pb.GetAccountResponse{
			Balance: 0,
			Message: "Account not found",
		}, nil
	}
	return &pb.GetAccountResponse{
		Balance: val.(float64),
		Message: "Account found",
	}, nil
}

