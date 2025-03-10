package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/vishnu1910/samplego/q3/protofiles/paymentpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// paymentQueue maintains pending payments.
type paymentQueue struct {
	mu       sync.Mutex
	payments []*pb.PaymentRequest
}

func (q *paymentQueue) Enqueue(req *pb.PaymentRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.payments = append(q.payments, req)
}

func (q *paymentQueue) DequeueAll() []*pb.PaymentRequest {
	q.mu.Lock()
	defer q.mu.Unlock()
	pending := q.payments
	q.payments = []*pb.PaymentRequest{}
	return pending
}

func main() {
	serverAddr := flag.String("server", "localhost:50050", "Payment Gateway address")
	flag.Parse()

	// Load TLS credentials for client.
	creds, err := credentials.NewClientTLSFromFile("ca.crt", "")
	if err != nil {
		log.Fatalf("Failed to load TLS credentials: %v", err)
	}

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to Payment Gateway: %v", err)
	}
	defer conn.Close()

	client := pb.NewPaymentGatewayClient(conn)

	// Create a payment queue.
	queue := &paymentQueue{}

	// For demonstration, we enqueue a payment.
	req := &pb.PaymentRequest{
		ClientId:      "client123",
		FromBank:      "BankA",
		ToBank:        "BankB",
		Amount:        100.0,
		IdempotencyKey: "unique-key-123",
	}
	queue.Enqueue(req)

	// Function to process queued payments.
	processPayments := func() {
		pending := queue.DequeueAll()
		for _, payment := range pending {
			// Add authentication metadata.
			ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer secret-token")
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resp, err := client.ProcessPayment(ctx, payment)
			if err != nil {
				log.Printf("Payment failed, requeuing: %v", err)
				queue.Enqueue(payment)
			} else {
				log.Printf("Payment result: %v", resp.Message)
			}
		}
	}

	// Retry loop: check every 5 seconds.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		processPayments()
	}
}

