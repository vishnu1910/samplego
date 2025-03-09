// client.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/vishnu1910/samplego/q3/protofiles/paymentpb"
)

// paymentQueue holds pending payment requests.
var paymentQueue = struct {
	sync.Mutex
	queue []*paymentpb.PaymentRequest
}{}

func sendPayment(gatewayAddr string, req *paymentpb.PaymentRequest) error {
	conn, err := grpc.Dial(gatewayAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	client := paymentpb.NewPaymentGatewayServiceClient(conn)

	// Set auth token in metadata.
	ctx := metadata.AppendToOutgoingContext(context.Background(), "auth_token", req.AuthToken)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := client.ProcessPayment(ctx, req)
	if err != nil {
		return err
	}
	log.Printf("Payment response: %s, message: %s", resp.Status, resp.Message)
	return nil
}

func processQueue(gatewayAddr string) {
	paymentQueue.Lock()
	defer paymentQueue.Unlock()
	var remaining []*paymentpb.PaymentRequest
	for _, req := range paymentQueue.queue {
		err := sendPayment(gatewayAddr, req)
		if err != nil {
			log.Printf("Retry failed for transaction %s: %v", req.TransactionId, err)
			remaining = append(remaining, req)
		} else {
			log.Printf("Retried transaction %s succeeded", req.TransactionId)
		}
	}
	paymentQueue.queue = remaining
}

func main() {
	gatewayAddr := flag.String("gateway", "localhost:50060", "Address of Payment Gateway")
	clientID := flag.String("client_id", "client1", "Client ID")
	authToken := flag.String("auth_token", "valid-token", "Authentication token")
	offlineMode := flag.Bool("offline", false, "Simulate offline mode")
	flag.Parse()

	// For demonstration, create a sample payment request.
	paymentReq := &paymentpb.PaymentRequest{
		ClientId:         *clientID,
		TransactionId:    fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		Amount:           250.0,
		SourceBankAddr:   "localhost:50061", // adjust as needed
		DestBankAddr:     "localhost:50062", // adjust as needed
		AuthToken:        *authToken,
	}

	// If offline flag is set, simulate failure by not contacting the gateway.
	if *offlineMode {
		log.Printf("Simulating offline: queuing transaction %s", paymentReq.TransactionId)
		paymentQueue.Lock()
		paymentQueue.queue = append(paymentQueue.queue, paymentReq)
		paymentQueue.Unlock()
	} else {
		err := sendPayment(*gatewayAddr, paymentReq)
		if err != nil {
			log.Printf("Error sending payment, queuing transaction: %v", err)
			paymentQueue.Lock()
			paymentQueue.queue = append(paymentQueue.queue, paymentReq)
			paymentQueue.Unlock()
		}
	}

	// Periodically try to process the offline queue.
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		log.Printf("Attempting to resend queued payments...")
		processQueue(*gatewayAddr)
	}
}

