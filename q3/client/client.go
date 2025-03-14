package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	pb "github.com/vishnu1910/samplego/q3/pb"
	"github.com/vishnu1910/samplego/q3/tlsutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Payment represents a payment request that can be queued.
type Payment struct {
	Request *pb.PaymentRequest
}

// Client encapsulates a Payment Gateway client.
type Client struct {
	id           string
	username     string
	password     string
	pgAddress    string
	conn         *grpc.ClientConn
	pgClient     pb.PaymentGatewayClient
	offlineQueue []*Payment
	queueLock    sync.Mutex
	tlsConfig    *credentials.TransportCredentials
}

// NewClient creates a new client with mutual TLS.
func NewClient(id, username, password, pgAddress, certFile, keyFile, caFile string) (*Client, error) {
	tlsConfig, err := tlsutil.LoadClientTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}
	creds := credentials.NewTLS(tlsConfig)
	conn, err := grpc.Dial(pgAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}
	client := &Client{
		id:        id,
		username:  username,
		password:  password,
		pgAddress: pgAddress,
		conn:      conn,
		pgClient:  pb.NewPaymentGatewayClient(conn),
	}
	// Start a background routine to retry offline payments.
	go client.processOfflineQueue()
	return client, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() {
	c.conn.Close()
}

// Register registers the client with the Payment Gateway.
func (c *Client) Register(bank, accountNumber string) error {
	// Attach metadata for authentication.
	ctx := metadata.AppendToOutgoingContext(context.Background(), "username", c.username, "password", c.password)
	req := &pb.RegisterRequest{
		Username:       c.username,
		Password:       c.password,
		Bank:           bank,
		Account_number: accountNumber,
	}
	resp, err := c.pgClient.RegisterClient(ctx, req)
	if err != nil {
		return err
	}
	log.Printf("[REGISTER] %v", resp.Message)
	return nil
}

// ProcessPayment sends a payment request.
// If the Payment Gateway is unreachable, it queues the payment.
func (c *Client) ProcessPayment(payment *pb.PaymentRequest) error {
	// Attach metadata for authentication.
	ctx := metadata.AppendToOutgoingContext(context.Background(), "username", c.username, "password", c.password)
	// Set a timeout.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.pgClient.ProcessPayment(ctx, payment)
	if err != nil {
		log.Printf("[OFFLINE] Payment gateway unreachable, queueing payment: %v", err)
		c.enqueuePayment(payment)
		return err
	}
	log.Printf("[PAYMENT] %v", resp.Message)
	return nil
}

// enqueuePayment adds a payment to the offline queue.
func (c *Client) enqueuePayment(payment *pb.PaymentRequest) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	c.offlineQueue = append(c.offlineQueue, &Payment{Request: payment})
	// Optionally, persist the queue to disk.
	c.persistQueue()
}

// processOfflineQueue periodically attempts to send queued payments.
func (c *Client) processOfflineQueue() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		c.queueLock.Lock()
		if len(c.offlineQueue) == 0 {
			c.queueLock.Unlock()
			continue
		}
		log.Printf("[QUEUE] Attempting to process %d offline payments", len(c.offlineQueue))
		var remaining []*Payment
		for _, p := range c.offlineQueue {
			err := c.ProcessPayment(p.Request)
			if err != nil {
				// Payment still failed, keep it in the queue.
				remaining = append(remaining, p)
			} else {
				log.Printf("[QUEUE] Payment processed successfully")
			}
		}
		c.offlineQueue = remaining
		c.persistQueue()
		c.queueLock.Unlock()
	}
}

// persistQueue saves the offline queue to disk (for recovery on restart).
func (c *Client) persistQueue() {
	data, err := json.Marshal(c.offlineQueue)
	if err != nil {
		log.Printf("[QUEUE] Error persisting queue: %v", err)
		return
	}
	err = ioutil.WriteFile("offline_queue_"+c.id+".json", data, 0644)
	if err != nil {
		log.Printf("[QUEUE] Error writing queue file: %v", err)
	}
}

// loadQueue loads the offline queue from disk.
func (c *Client) loadQueue() {
	data, err := ioutil.ReadFile("offline_queue_" + c.id + ".json")
	if err != nil {
		log.Printf("[QUEUE] No offline queue found, starting fresh.")
		return
	}
	var queue []*Payment
	err = json.Unmarshal(data, &queue)
	if err != nil {
		log.Printf("[QUEUE] Error parsing queue file: %v", err)
		return
	}
	c.offlineQueue = queue
	log.Printf("[QUEUE] Loaded %d offline payments", len(queue))
}

