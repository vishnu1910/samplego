package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vishnu1910/samplego/P3/client"
)

func main() {
	// Command-line flags for Payment Gateway address and TLS files.
	pgAddress := flag.String("pg", "localhost:50051", "Payment Gateway address")
	certFile := flag.String("cert", "../../certs/client.crt", "TLS cert file")
	keyFile := flag.String("key", "../../certs/client.key", "TLS key file")
	caFile := flag.String("ca", "../../certs/ca.crt", "CA cert file")
	flag.Parse()

	// Create a new client instance.
	cl, err := client.NewClient(*pgAddress, *certFile, *keyFile, *caFile)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cl.Close()

	// Interactive login prompt.
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Set client's credentials.
	cl.SetCredentials(username, password)

	// Determine bank details for the user.
	var bank, account string
	switch username {
	case "alice":
		bank = "BankA"
		account = "ACC123"
	case "bob":
		bank = "BankB"
		account = "ACC456"
	default:
		log.Fatalf("Unknown user: %s", username)
	}

	// Register (login) with the Payment Gateway.
	err = cl.Register(bank, account)
	if err != nil {
		log.Fatalf("Registration failed: %v", err)
	}

	fmt.Println("Login successful!")
	// Command loop.
	for {
		fmt.Print("Enter command (pay, balance, exit): ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		switch cmd {
		case "exit":
			fmt.Println("Exiting...")
			return
		case "balance":
			// Query balance.
			bal, err := cl.GetBalance()
			if err != nil {
				fmt.Printf("Error getting balance: %v\n", err)
			} else {
				fmt.Printf("Your balance: %.2f\n", bal)
			}
		case "pay":
			// Payment command.
			fmt.Print("Enter recipient bank (BankA or BankB): ")
			toBank, _ := reader.ReadString('\n')
			toBank = strings.TrimSpace(toBank)
			fmt.Print("Enter recipient account (e.g., ACC123 or ACC456): ")
			toAccount, _ := reader.ReadString('\n')
			toAccount = strings.TrimSpace(toAccount)
			fmt.Print("Enter amount: ")
			amtStr, _ := reader.ReadString('\n')
			amtStr = strings.TrimSpace(amtStr)
			amount, err := strconv.ParseFloat(amtStr, 64)
			if err != nil {
				fmt.Println("Invalid amount")
				continue
			}
			// For idempotency, use a timestamp-based key.
			idemKey := fmt.Sprintf("%d", time.Now().UnixNano())
			err = cl.ProcessPayment(bank, account, toBank, toAccount, amount, idemKey)
			if err != nil {
				fmt.Printf("Payment error: %v\n", err)
			} else {
				fmt.Println("Payment processed successfully.")
			}
		default:
			fmt.Println("Unknown command. Valid commands: pay, balance, exit")
		}
	}
}

