package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/vishnu1910/samplego/q3/gateway"
	"github.com/vishnu1910/samplego/q3/middleware"
	"github.com/vishnu1910/samplego/q3/pb"
	"github.com/vishnu1910/samplego/q3/tlsutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// Command-line flags for TLS certificate files.
	certFile := flag.String("cert", "certs/gateway.crt", "TLS cert file")
	keyFile := flag.String("key", "certs/gateway.key", "TLS key file")
	caFile := flag.String("ca", "certs/ca.crt", "CA cert file")
	port := flag.Int("port", 50051, "Gateway server port")
	flag.Parse()

	// Load TLS configuration.
	tlsConfig, err := tlsutil.LoadServerTLSConfig(*certFile, *keyFile, *caFile)
	if err != nil {
		log.Fatalf("failed to load TLS config: %v", err)
	}
	creds := credentials.NewTLS(tlsConfig)

	// Create gRPC options with TLS and chained interceptors.
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(middleware.ChainUnaryInterceptors(
			middleware.UnaryAuthInterceptor,
			middleware.UnaryLoggingInterceptor,
		)),
	}

	// Dial connections to bank servers.
	// For simplicity, we assume two banks running on known ports.
	bankAddresses := map[string]string{
		"BankA": "localhost:60051",
		"BankB": "localhost:60052",
	}
	bankClients := make(map[string]pb.BankServiceClient)
	for bank, addr := range bankAddresses {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
		if err != nil {
			log.Fatalf("failed to connect to bank %s: %v", bank, err)
		}
		bankClients[bank] = pb.NewBankServiceClient(conn)
		log.Printf("Connected to bank %s at %s", bank, addr)
	}

	// Create PaymentGateway server with a 2PC timeout of 5 seconds.
	pgServer := gateway.NewPaymentGatewayServer(bankClients, 5*time.Second)

	// Start the gRPC server.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPaymentGatewayServer(grpcServer, pgServer)

	log.Printf("Payment Gateway running on port %d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

