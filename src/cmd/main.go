package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/config"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/controllers"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/consensus"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/dag"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/routes"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/services"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	simulationMode := flag.Bool("simulation", false, "Run in simulation mode")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	if *simulationMode {
		runSimulation(cfg)
		return
	}

	// Initialize models
	dagModel := dag.NewDAG()
	consensusModel := consensus.NewAvalanche(dagModel, cfg.ConsensusParams)

	// Initialize services
	// Create peer service with a placeholder receive function first
	peerService := services.NewPeerService(cfg.NodeID, nil)

	// Create consensus service
	consensusService := services.NewConsensusService(
		cfg.NodeID,
		consensusModel,
		peerService,
	)

	// Set the receive function for the peer service
	peerService.SetReceiveVertexFunc(func(id string, data interface{}, parentIDs []string) error {
		_, err := consensusService.ReceiveVertex(id, data, parentIDs)
		return err
	})

	// Initialize controllers
	vertexController := controllers.NewVertexController(consensusService)
	consensusController := controllers.NewConsensusController(consensusService)
	peerController := controllers.NewPeerController(peerService)
	healthController := controllers.NewHealthController()

	// Initialize router
	router := routes.NewRouter(
		vertexController,
		consensusController,
		peerController,
		healthController,
	)

	// Create HTTP server
	mux := http.NewServeMux()
	router.RegisterRoutes(mux)

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: mux,
	}

	// Connect to peers
	if len(cfg.PeerAddresses) > 0 {
		if err := peerService.ConnectToPeers(cfg.PeerAddresses); err != nil {
			log.Printf("Error connecting to peers: %v", err)
		}
	}

	// Start consensus
	if err := consensusService.StartConsensus(); err != nil {
		log.Printf("Error starting consensus: %v", err)
	}

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on port %d", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutting down...")

	// Stop consensus
	if err := consensusService.StopConsensus(); err != nil {
		log.Printf("Error stopping consensus: %v", err)
	}

	log.Println("Server stopped")
}

// runSimulation runs the consensus simulation
func runSimulation(cfg *config.Config) {
	log.Println("Running simulation mode...")
	
	// Initialize models
	dagModel := dag.NewDAG()
	consensusModel := consensus.NewAvalanche(dagModel, cfg.ConsensusParams)
	
	// Create a simple simulation
	sim := services.NewSimulationService(consensusModel)
	
	// Run simulation for 30 seconds
	duration := 30 * time.Second
	log.Printf("Running simulation for %s...", duration)
	
	stop := make(chan struct{})
	go consensusModel.RunConsensus(stop)
	
	results := sim.RunRandomVertices(100, 5)
	
	time.Sleep(duration)
	close(stop)
	
	// Print results
	log.Printf("Simulation completed with %d vertices", len(results))
	
	finalized := consensusModel.GetFinalized()
	log.Printf("Finalized %d vertices", len(finalized))
} 