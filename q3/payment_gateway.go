// payment_gateway.go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/vishnu1910/samplego/q3/protofiles/paymentpb"
)

// idempotency record: maps transaction ID to PaymentResponse.
var processedTx sync.Map

// authToken that clients must supply.
const validAuthToken = "valid-token"

// --- Interceptors ---

// authInterceptor checks for a valid auth token in the incoming PaymentRequest.
func authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// We only check auth for PaymentGatewayService.ProcessPayment
	if info.FullMethod == "/payment.PaymentGatewayService/ProcessPayment" {
		// Extract metadata from context.
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}
		// Expect an "auth_token" value.
		tokens := md.Get("auth_token")
		if len(tokens) == 0 || tokens[0] != validAuthToken {
			return nil, errors.New("invalid auth token")
		}
	}
	// Continue handling the RPC.
	return handler(ctx, req)
}

// loggingInterceptor logs the request and response.
func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("Received RPC: %s, Request: %+v", info.FullMethod, req)
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("RPC %s error: %v", info.FullMethod, err)
	} else {
		log.Printf("RPC %s response: %+v", info.FullMethod, resp)
	}
	return resp, err
}

// --- Payment Gateway Implementation ---

type paymentGatewayServer struct {
	paymentpb.UnimplementedPaymentGatewayServiceServer
	// timeout for bank RPCs
	bankRPCTimeout time.Duration
}

func (pg *paymentGatewayServer) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
	// Idempotency check: if we've seen this transaction ID, return previous response.
	if resp, ok := processedTx.Load(req.TransactionId); ok {
		log.Printf("Idempotency: returning cached response for transaction %s", req.TransactionId)
		return resp.(*paymentpb.PaymentResponse), nil
	}

	// (In a real system, more robust auth would be done here.)
	// Begin two-phase commit (2PC)
	// Prepare phase: contact both source and destination banks.
	prepareReq := &paymentpb.PrepareRequest{
		TransactionId: req.TransactionId,
		Amount:        req.Amount,
		ClientId:      req.ClientId,
	}

	// Prepare calls concurrently.
	var wg sync.WaitGroup
	wg.Add(2)
	var srcVote, destVote bool
	var srcMsg, destMsg string
	var srcErr, destErr error

	callPrepare := func(bankAddr string, vote *bool, msg *string, errOut *error) {
		defer wg.Done()
		conn, err := grpc.Dial(bankAddr, grpc.WithInsecure())
		if err != nil {
			*errOut = fmt.Errorf("failed to dial bank %s: %v", bankAddr, err)
			return
		}
		defer conn.Close()
		client := paymentpb.NewBankServiceClient(conn)
		ctx2, cancel := context.WithTimeout(context.Background(), pg.bankRPCTimeout)
		defer cancel()
		resp, err := client.PreparePayment(ctx2, prepareReq)
		if err != nil {
			*errOut = err
			return
		}
		*vote = resp.Vote
		*msg = resp.Message
	}

	go callPrepare(req.SourceBankAddr, &srcVote, &srcMsg, &srcErr)
	go callPrepare(req.DestBankAddr, &destVote, &destMsg, &destErr)
	wg.Wait()

	// Check for errors or negative votes.
	if srcErr != nil || destErr != nil || !srcVote || !destVote {
		// Abort transaction: call AbortPayment on both banks.
		abortReq := &paymentpb.AbortRequest{TransactionId: req.TransactionId}
		go func() {
			callAbort(req.SourceBankAddr, abortReq)
		}()
		go func() {
			callAbort(req.DestBankAddr, abortReq)
		}()
		response := &paymentpb.PaymentResponse{
			Status:  "FAILURE",
			Message: fmt.Sprintf("Transaction aborted: srcErr=%v, destErr=%v, srcVote=%v, destVote=%v", srcErr, destErr, srcVote, destVote),
		}
		processedTx.Store(req.TransactionId, response)
		return response, nil
	}

	// Commit phase: call CommitPayment on both banks.
	commitReq := &paymentpb.CommitRequest{TransactionId: req.TransactionId}
	var commitErrSrc, commitErrDest error
	var wg2 sync.WaitGroup
	wg2.Add(2)
	go func() {
		defer wg2.Done()
		commitErrSrc = callCommit(req.SourceBankAddr, commitReq)
	}()
	go func() {
		defer wg2.Done()
		commitErrDest = callCommit(req.DestBankAddr, commitReq)
	}()
	wg2.Wait()

	if commitErrSrc != nil || commitErrDest != nil {
		response := &paymentpb.PaymentResponse{
			Status:  "FAILURE",
			Message: fmt.Sprintf("Commit failed: srcErr=%v, destErr=%v", commitErrSrc, commitErrDest),
		}
		processedTx.Store(req.TransactionId, response)
		return response, nil
	}

	response := &paymentpb.PaymentResponse{
		Status:  "SUCCESS",
		Message: "Payment processed successfully",
	}
	processedTx.Store(req.TransactionId, response)
	return response, nil
}

func callCommit(bankAddr string, req *paymentpb.CommitRequest) error {
	conn, err := grpc.Dial(bankAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to dial bank %s: %v", bankAddr, err)
	}
	defer conn.Close()
	client := paymentpb.NewBankServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := client.CommitPayment(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Committed {
		return errors.New(resp.Message)
	}
	return nil
}

func callAbort(bankAddr string, req *paymentpb.AbortRequest) {
	conn, err := grpc.Dial(bankAddr, grpc.WithInsecure())
	if err != nil {
		log.Printf("Abort: failed to dial bank %s: %v", bankAddr, err)
		return
	}
	defer conn.Close()
	client := paymentpb.NewBankServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = client.AbortPayment(ctx, req)
	if err != nil {
		log.Printf("Abort error on bank %s: %v", bankAddr, err)
	}
}

func main() {
	port := flag.Int("port", 50060, "Port for Payment Gateway")
	bankTimeout := flag.Int("bank_timeout", 5, "Timeout (in seconds) for bank RPC calls")
	flag.Parse()

	// Set up interceptors.
	opts := []grpc.ServerOption{
    		grpc.ChainUnaryInterceptor(authInterceptor, loggingInterceptor),
	}


	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(opts...)
	paymentpb.RegisterPaymentGatewayServiceServer(grpcServer, &paymentGatewayServer{
		bankRPCTimeout: time.Duration(*bankTimeout) * time.Second,
	})
	log.Printf("Payment Gateway server listening on port %d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

