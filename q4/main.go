package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	pb "github.com/vishnu1910/samplego/q4/byzantine" // import the generated protobuf package

	"google.golang.org/grpc"
)

// General represents a node (general) in the simulation.
type General struct {
	id            int
	address       string
	isCommander   bool
	isTraitor     bool
	currentOrder  string
	roundChannels map[int]chan string // channels to collect messages per round
	mu            sync.Mutex
}

// NewGeneral creates a new General.
func NewGeneral(id int, basePort int, isCommander bool) *General {
	return &General{
		id:            id,
		address:       fmt.Sprintf("localhost:%d", basePort+id),
		isCommander:   isCommander,
		roundChannels: make(map[int]chan string),
	}
}

// startServer starts the gRPC server for the General.
func (g *General) startServer(wg *sync.WaitGroup) {
	defer wg.Done()

	lis, err := net.Listen("tcp", g.address)
	if err != nil {
		log.Fatalf("General %d failed to listen: %v", g.id, err)
	}
	s := grpc.NewServer()
	pb.RegisterByzantineServer(s, g)
	log.Printf("General %d listening on %s\n", g.id, g.address)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("General %d failed to serve: %v", g.id, err)
	}
}

// SendMessage is the gRPC method that receives messages.
func (g *General) SendMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	// Look up the channel for the round.
	g.mu.Lock()
	ch, ok := g.roundChannels[int(req.Round)]
	g.mu.Unlock()
	if ok {
		// Deliver the order into the channel.
		ch <- req.Order
	} else {
		log.Printf("General %d: no channel for round %d (message from %d)", g.id, req.Round, req.SenderId)
	}
	return &pb.MessageResponse{Ack: true}, nil
}

// runRound executes a single round of message exchange.
// For round 1: lieutenants wait for the commander's order.
// For rounds 2 to t+1: every general broadcasts its current decision.
func (g *General) runRound(round int, peers []*General, wg *sync.WaitGroup) {
	defer wg.Done()

	if round == 1 {
		// Round 1: Only lieutenants need to wait for the commander’s message.
		if !g.isCommander {
			// Expect exactly 1 message from the commander.
			ch := make(chan string, 1)
			g.mu.Lock()
			g.roundChannels[round] = ch
			g.mu.Unlock()
			// Wait for message.
			msg := <-ch
			log.Printf("General %d [Round %d] received order from commander: %s", g.id, round, msg)
			// Honest lieutenants adopt the commander's order.
			if !g.isTraitor {
				g.currentOrder = msg
			} else {
				// A traitor lieutenant might ignore the order (or keep its own arbitrary decision).
				log.Printf("General %d [Round %d] is traitorous and ignores the commander's order", g.id, round)
			}
		}
	} else {
		// Rounds 2 to t+1: every general broadcasts its current order.
		expected := len(peers) // messages expected from all other generals
		ch := make(chan string, expected)
		g.mu.Lock()
		g.roundChannels[round] = ch
		g.mu.Unlock()

		// Broadcast our (possibly modified) order to all peers concurrently.
		for _, peer := range peers {
			go sendMessage(g, peer, round, g.getMessageForPeer(peer))
		}

		// Wait for all expected messages.
		received := []string{}
		for i := 0; i < expected; i++ {
			msg := <-ch
			received = append(received, msg)
		}

		// Include our own order.
		received = append(received, g.currentOrder)
		newDecision := majority(received)
		log.Printf("General %d [Round %d] received: %v; majority vote yields: %s", g.id, round, received, newDecision)

		// Honest generals adopt the majority vote; traitors may keep their own.
		if !g.isTraitor {
			g.currentOrder = newDecision
		} else {
			// For simulation, we leave traitors with their original (possibly conflicting) decision.
			log.Printf("General %d [Round %d] is traitorous and does not change its decision (currently: %s)", g.id, round, g.currentOrder)
		}
	}
}

// getMessageForPeer returns the message that this general will send to a particular peer in the current round.
// Honest generals always send their current decision.
// Traitors can send conflicting orders by randomly choosing "Attack" or "Retreat".
func (g *General) getMessageForPeer(peer *General) string {
	if !g.isTraitor {
		return g.currentOrder
	}
	// Traitor behavior: send a random order (could be different for each peer).
	orders := []string{"Attack", "Retreat"}
	return orders[rand.Intn(len(orders))]
}

// sendMessage dials the peer’s gRPC server and sends the message.
func sendMessage(sender *General, recipient *General, round int, order string) {
	conn, err := grpc.Dial(recipient.address, grpc.WithInsecure())
	if err != nil {
		log.Printf("General %d: error dialing General %d: %v", sender.id, recipient.id, err)
		return
	}
	defer conn.Close()
	client := pb.NewByzantineClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.MessageRequest{
		Round:    int32(round),
		SenderId: int32(sender.id),
		Order:    order,
	}
	_, err = client.SendMessage(ctx, req)
	if err != nil {
		log.Printf("General %d: error sending message to General %d: %v", sender.id, recipient.id, err)
	}
}

// majority returns "Attack" if the count of "Attack" exceeds "Retreat"; otherwise "Retreat".
// (In case of tie, "Retreat" is chosen.)
func majority(messages []string) string {
	countAttack := 0
	countRetreat := 0
	for _, m := range messages {
		if m == "Attack" {
			countAttack++
		} else if m == "Retreat" {
			countRetreat++
		}
	}
	if countAttack > countRetreat {
		return "Attack"
	}
	return "Retreat"
}

func main() {
	// Seed the random number generator.
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <n> <t>\n", os.Args[0])
		os.Exit(1)
	}

	// n: total number of generals; t: number of traitors.
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid n: %v", err)
	}
	t, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Invalid t: %v", err)
	}

	// Ensure condition n > 3t
	if n <= 3*t {
		log.Fatalf("Condition n > 3t is not satisfied (n=%d, t=%d)", n, t)
	}

	// Base port for gRPC servers.
	basePort := 50051

	// Create generals. We designate general 0 as commander.
	generals := make([]*General, n)
	for i := 0; i < n; i++ {
		isCommander := (i == 0)
		generals[i] = NewGeneral(i, basePort, isCommander)
	}

	// For this simulation, mark the first t lieutenants (starting from 1) as traitors.
	// (You could choose them randomly if desired.)
	for i := 1; i <= t && i < n; i++ {
		generals[i].isTraitor = true
	}

	// Set the initial orders.
	// Commander: if honest, choose a fixed order (e.g., "Attack").
	// If traitor (not in this simulation but can be enabled), the commander may decide arbitrarily.
	if generals[0].isTraitor {
		// For a traitorous commander, we randomly choose an order.
		orders := []string{"Attack", "Retreat"}
		generals[0].currentOrder = orders[rand.Intn(len(orders))]
	} else {
		generals[0].currentOrder = "Attack"
	}
	// Lieutenants start with an empty order.
	for i := 1; i < n; i++ {
		// For traitors, you might preassign an arbitrary decision.
		if generals[i].isTraitor {
			orders := []string{"Attack", "Retreat"}
			generals[i].currentOrder = orders[rand.Intn(len(orders))]
		} else {
			generals[i].currentOrder = "" // will be set in round 1
		}
	}

	// Start gRPC servers for all generals.
	var serverWG sync.WaitGroup
	serverWG.Add(n)
	for i := 0; i < n; i++ {
		go generals[i].startServer(&serverWG)
	}
	// Give the servers a moment to start.
	time.Sleep(2 * time.Second)

	// --- Simulation Rounds ---

	// Round 1: Commander sends its order to all lieutenants.
	var roundWG sync.WaitGroup
	round := 1
	for i := 0; i < n; i++ {
		// Run round 1 concurrently for all generals.
		roundWG.Add(1)
		go generals[i].runRound(round, generalsExcept(generals, generals[i].id), &roundWG)
	}
	roundWG.Wait()

	// Rounds 2 to t+1.
	totalRounds := t + 1
	for round = 2; round <= totalRounds; round++ {
		log.Printf("=== Starting Round %d ===", round)
		roundWG = sync.WaitGroup{}
		for i := 0; i < n; i++ {
			roundWG.Add(1)
			go generals[i].runRound(round, generalsExcept(generals, generals[i].id), &roundWG)
		}
		roundWG.Wait()
	}

	// Print final decisions.
	fmt.Println("\n--- Final Decisions ---")
	for _, g := range generals {
		role := "Lieutenant"
		if g.isCommander {
			role = "Commander"
		}
		traitor := ""
		if g.isTraitor {
			traitor = " (Traitor)"
		}
		fmt.Printf("General %d [%s%s]: Final decision = %s\n", g.id, role, traitor, g.currentOrder)
	}

	// Check if all honest generals reached consensus.
	consensus := true
	var final string
	for _, g := range generals {
		if !g.isTraitor {
			if final == "" {
				final = g.currentOrder
			} else if g.currentOrder != final {
				consensus = false
			}
		}
	}
	if consensus {
		fmt.Printf("\nConsensus reached among honest generals: %s\n", final)
	} else {
		fmt.Println("\nConsensus NOT reached among honest generals!")
	}

	// Since the gRPC servers are running in goroutines, exit the process.
	os.Exit(0)
}

// generalsExcept returns a slice of pointers to all generals except the one with the given id.
func generalsExcept(all []*General, selfID int) []*General {
	others := []*General{}
	for _, g := range all {
		if g.id != selfID {
			others = append(others, g)
		}
	}
	return others
}

