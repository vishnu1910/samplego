package main

import (
	"flag"
	"log"
	"time"

	"github.com/vishnu1910/samplego/q3/client"
	"github.com/vishnu1910/samplego/q3/pb"
)

func main() {
	// Command-line flags.
	pgAddress := flag.String("pg", "localhost:50051", "Payment Gateway address")
	certFile := flag.String("cert", "certs/client.crt", "TLS cert file")
	keyFile := flag.String("key", "certs/client.key", "TLS key file")
	caFile := flag.String("ca", "certs/ca.crt", "CA cert file")
	clientID := flag.String("id", "client1", "Client identifier")
	username := flag.String("username", "alice", "Username")
	password := flag.String("password", "alicepassword", "Password")
	flag.Parse()

	// Create a new client.
	cl, err := client.NewClient(*clientID, *username, *password, *pgAddress, *certFile, *keyFile, *caFile)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cl.Close()

	// Register with the Payment Gateway.
	err = cl.Register("BankA", "ACC123")
	if err != nil {
		log.Fatalf("Registration failed: %v", err)
	}

	// Simulate a payment.
	paymentReq := &pb.PaymentRequest{
		Client_id:      *username,
		Idempotency_key: "unique-payment-key-123", // In production, use a UUID.
		From_bank:      "BankA",
		From_account:   "ACC123",
		To_bank:        "BankB",
		To_account:     "ACC456",
		Amount:         100.0,
	}

	// Try to process the payment. If the Payment Gateway is offline, it will be queued.
	err = cl.ProcessPayment(paymentReq)
	if err != nil {
		log.Printf("Payment processing error: %v", err)
	}

	// Simulate the client running for a while to process any offline queued payments.
	time.Sleep(30 * time.Second)
}

