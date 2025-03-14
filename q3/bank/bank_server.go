package bank

import (
	"context"
	"fmt"
	"log"
	"sync"

	pb "github.com/vishnu1910/samplego/q3/pb"
)

// BankServer implements the BankService.
type BankServer struct {
	pb.UnimplementedBankServiceServer
	// accounts stores account balances keyed by account number.
	accounts sync.Map // key: account number, value: float64

	// For 2PC, we store pending transactions.
	pendingTransactions sync.Map // key: transactionID, value: float64 (amount)
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
	val, ok := s.accounts.Load(req.Account_number)
	if !ok {
		return &pb.PrepareResponse{
			CanCommit: false,
			Message:   "Account not found",
		}, nil
	}
	currentBalance := val.(float64)

	// If this is a sender, ensure funds are sufficient.
	// We assume that if currentBalance >= amount then it is a sender.
	if currentBalance >= req.Amount {
		// Reserve funds by storing the pending transaction.
		s.pendingTransactions.Store(req.Transaction_id, req.Amount)
		return &pb.PrepareResponse{
			CanCommit: true,
			Message:   "Funds reserved",
		}, nil
	}

	// Otherwise, for a receiver, always allow.
	// (Alternatively, you might want to reserve a credit limit.)
	s.pendingTransactions.Store(req.Transaction_id, req.Amount)
	return &pb.PrepareResponse{
		CanCommit: true,
		Message:   "Ready to credit",
	}, nil
}

// Commit finalizes the transaction by updating the balance.
func (s *BankServer) Commit(ctx context.Context, req *pb.CommitRequest) (*pb.CommitResponse, error) {
	val, ok := s.pendingTransactions.Load(req.Transaction_id)
	if !ok {
		return &pb.CommitResponse{
			Success: false,
			Message: "No pending transaction found",
		}, nil
	}
	amount := val.(float64)
	// Determine if this account is a sender or receiver.
	// For simplicity, if balance >= amount then it is a deduction.
	// Otherwise, it is a credit.
	// In a real system, the Prepare phase would indicate the operation type.
	// Here we update accordingly.
	s.accounts.Range(func(key, value interface{}) bool {
		// In a real system, we would know which account to update.
		// This is just a placeholder.
		_ = key
		_ = value
		return true
	})
	// In this simulation, assume the Commit always succeeds.
	// Remove the pending transaction.
	s.pendingTransactions.Delete(req.Transaction_id)
	return &pb.CommitResponse{
		Success: true,
		Message: "Transaction committed",
	}, nil
}

// Abort cancels the transaction.
func (s *BankServer) Abort(ctx context.Context, req *pb.AbortRequest) (*pb.AbortResponse, error) {
	// Remove pending transaction.
	s.pendingTransactions.Delete(req.Transaction_id)
	return &pb.AbortResponse{
		Success: true,
		Message: "Transaction aborted",
	}, nil
}

// GetAccount returns the account balance.
func (s *BankServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	val, ok := s.accounts.Load(req.Account_number)
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

