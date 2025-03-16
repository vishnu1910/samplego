package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
	"time"

	pb "github.com/vishnu1910/samplego/P3/protofiles/pb"
	"github.com/vishnu1910/samplego/P3/tlsutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Client encapsulates a Payment Gateway client.
type Client struct {
	username  string
	password  string
	pgAddress string
	conn      *grpc.ClientConn
	pgClient  pb.PaymentGatewayClient

	offlineQueue []*Payment
	queueLock    sync.Mutex
}

// Payment represents a payment request that can be queued.
type Payment struct {
	Request *pb.PaymentRequest
}

// NewClient creates a new client with mutual TLS.
func NewClient(pgAddress, certFile, keyFile, caFile string) (*Client, error) {
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
		pgAddress: pgAddress,
		conn:      conn,
		pgClient:  pb.NewPaymentGatewayClient(conn),
	}
	client.loadQueue()
	go client.processOfflineQueue()
	return client, nil
}

// SetCredentials sets the client's username and password.
func (c *Client) SetCredentials(username, password string) {
	c.username = username
	c.password = password
}

// Close closes the gRPC connection.
func (c *Client) Close() {
	c.conn.Close()
}

// Register registers the client with the Payment Gateway.
func (c *Client) Register(bank, accountNumber string) error {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "username", c.username, "password", c.password)
	req := &pb.RegisterRequest{
		Username:      c.username,
		Password:      c.password,
		Bank:          bank,
		AccountNumber: accountNumber,
	}
	resp, err := c.pgClient.RegisterClient(ctx, req)
	if err != nil {
		return err
	}
	log.Printf("[REGISTER] %v", resp.Message)
	return nil
}

// ProcessPayment sends a payment request. If the Payment Gateway is unreachable, it queues the payment.
func (c *Client) ProcessPayment(fromBank, fromAccount, toBank, toAccount string, amount float64, idempotencyKey string) error {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "username", c.username, "password", c.password)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.PaymentRequest{
		ClientId:       c.username,
		IdempotencyKey: idempotencyKey,
		FromBank:       fromBank,
		FromAccount:    fromAccount,
		ToBank:         toBank,
		ToAccount:      toAccount,
		Amount:         amount,
	}
	resp, err := c.pgClient.ProcessPayment(ctx, req)
	if err != nil {
		log.Printf("[OFFLINE] Payment gateway unreachable, queueing payment: %v", err)
		c.enqueuePayment(req)
		return err
	}
	log.Printf("[PAYMENT] %v", resp.Message)
	return nil
}

// GetBalance retrieves the client's balance.
func (c *Client) GetBalance() (float64, error) {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "username", c.username, "password", c.password)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.BalanceRequest{
		ClientId: c.username,
	}
	resp, err := c.pgClient.GetBalance(ctx, req)
	if err != nil {
		return 0, err
	}
	return resp.Balance, nil
}

// Offline queue processing methods.

func (c *Client) enqueuePayment(payment *pb.PaymentRequest) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	c.offlineQueue = append(c.offlineQueue, &Payment{Request: payment})
	c.persistQueue()
}

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
			err := c.ProcessPayment(p.Request.FromBank, p.Request.FromAccount, p.Request.ToBank, p.Request.ToAccount, p.Request.Amount, p.Request.IdempotencyKey)
			if err != nil {
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

func (c *Client) persistQueue() {
	data, err := json.Marshal(c.offlineQueue)
	if err != nil {
		log.Printf("[QUEUE] Error persisting queue: %v", err)
		return
	}
	err = ioutil.WriteFile("offline_queue_"+c.username+".json", data, 0644)
	if err != nil {
		log.Printf("[QUEUE] Error writing queue file: %v", err)
	}
}

func (c *Client) loadQueue() {
	data, err := ioutil.ReadFile("offline_queue_" + c.username + ".json")
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

