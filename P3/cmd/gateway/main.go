package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/vishnu1910/samplego/P3/gateway"
	"github.com/vishnu1910/samplego/P3/middleware"
	"github.com/vishnu1910/samplego/P3/protofiles/pb"
	"github.com/vishnu1910/samplego/P3/tlsutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// Command-line flags for TLS certificate files.
	certFile := flag.String("cert", "../../certs/gateway.crt", "TLS cert file for serving")
	keyFile := flag.String("key", "../../certs/gateway.key", "TLS key file for serving")
	caFile := flag.String("ca", "../../certs/ca.crt", "CA cert file")
	port := flag.Int("port", 50051, "Gateway server port")
	flag.Parse()

	// Load server TLS configuration for the Payment Gateway server.
	serverTLSConfig, err := tlsutil.LoadServerTLSConfig(*certFile, *keyFile, *caFile)
	if err != nil {
		log.Fatalf("failed to load server TLS config: %v", err)
	}
	serverCreds := credentials.NewTLS(serverTLSConfig)

	// Load a separate client TLS configuration for dialing bank servers.
	// Make sure you have a client certificate for the gateway (e.g., gateway-client.crt and gateway-client.key)
	clientTLSConfig, err := tlsutil.LoadClientTLSConfig("../../certs/gateway-client.crt", "../../certs/gateway-client.key", *caFile)
	if err != nil {
		log.Fatalf("failed to load client TLS config for bank connection: %v", err)
	}
	clientCreds := credentials.NewTLS(clientTLSConfig)

	// Dial connections to bank servers using the client TLS config.
	bankAddresses := map[string]string{
		"BankA": "localhost:60051",
		"BankB": "localhost:60052",
	}
	bankClients := make(map[string]pb.BankServiceClient)
	for bank, addr := range bankAddresses {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(clientCreds))
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
	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.UnaryInterceptor(middleware.ChainUnaryInterceptors(
			middleware.UnaryAuthInterceptor,
			middleware.UnaryLoggingInterceptor,
		)),
	)
	pb.RegisterPaymentGatewayServer(grpcServer, pgServer)

	log.Printf("Payment Gateway running on port %d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

