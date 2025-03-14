package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/vishnu1910/samplego/q3/bank"
	"github.com/vishnu1910/samplego/q3/middleware"
	"github.com/vishnu1910/samplego/q3/pb"
	"github.com/vishnu1910/samplego/q3/tlsutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// Command-line flags for TLS certificate files.
	certFile := flag.String("cert", "certs/bank.crt", "TLS cert file")
	keyFile := flag.String("key", "certs/bank.key", "TLS key file")
	caFile := flag.String("ca", "certs/ca.crt", "CA cert file")
	port := flag.Int("port", 60051, "Bank server port")
	flag.Parse()

	// Load TLS configuration.
	tlsConfig, err := tlsutil.LoadServerTLSConfig(*certFile, *keyFile, *caFile)
	if err != nil {
		log.Fatalf("failed to load TLS config: %v", err)
	}
	creds := credentials.NewTLS(tlsConfig)

	// Create gRPC server with interceptors.
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(middleware.ChainUnaryInterceptors(
			middleware.UnaryAuthInterceptor,
			middleware.UnaryLoggingInterceptor,
		)),
	}

	// Create a BankServer with some dummy accounts.
	// For demonstration, each bank server instance will be configured with its own set of accounts.
	initialAccounts := map[string]float64{
		"ACC123": 1000.0,
		"ACC456": 1000.0,
	}
	bankServer := bank.NewBankServer(initialAccounts)

	// Start gRPC server.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterBankServiceServer(grpcServer, bankServer)

	log.Printf("Bank Server running on port %d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

