package services

import (
	"fmt"
	"sync"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/consensus"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/dag"
)

// ConsensusService encapsulates consensus operations
type ConsensusService struct {
	mu          sync.RWMutex
	nodeID      string
	avalanche   *consensus.Avalanche
	stopChan    chan struct{}
	isRunning   bool
	peerService PeerServiceInterface
}

// PeerServiceInterface defines the interface for peer communications
type PeerServiceInterface interface {
	BroadcastVertex(id string, data interface{}, parentIDs []string) error
	GetPeers() []string
	ConnectToPeers(peers []string) error
}

// NewConsensusService creates a new consensus service
func NewConsensusService(nodeID string, avalanche *consensus.Avalanche, peerService PeerServiceInterface) *ConsensusService {
	return &ConsensusService{
		nodeID:      nodeID,
		avalanche:   avalanche,
		stopChan:    make(chan struct{}),
		isRunning:   false,
		peerService: peerService,
	}
}

// StartConsensus starts the consensus algorithm
func (s *ConsensusService) StartConsensus() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.isRunning {
		return fmt.Errorf("consensus is already running")
	}
	
	s.stopChan = make(chan struct{})
	go s.avalanche.RunConsensus(s.stopChan)
	s.isRunning = true
	
	return nil
}

// StopConsensus stops the consensus algorithm
func (s *ConsensusService) StopConsensus() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.isRunning {
		return fmt.Errorf("consensus is not running")
	}
	
	close(s.stopChan)
	s.isRunning = false
	
	return nil
}

// ProposeVertex proposes a new vertex to the network
func (s *ConsensusService) ProposeVertex(id string, data interface{}, parentIDs []string) (*dag.Vertex, error) {
	// Add vertex to local DAG
	vertex, err := s.avalanche.AddVertex(id, data, parentIDs)
	if err != nil {
		return nil, err
	}
	
	// Broadcast to peers if peer service is available
	if s.peerService != nil {
		if err := s.peerService.BroadcastVertex(id, data, parentIDs); err != nil {
			// Log the error but don't fail the operation
			fmt.Printf("Error broadcasting vertex: %v\n", err)
		}
	}
	
	return vertex, nil
}

// ReceiveVertex handles receiving a vertex from a peer
func (s *ConsensusService) ReceiveVertex(id string, data interface{}, parentIDs []string) (*dag.Vertex, error) {
	return s.avalanche.AddVertex(id, data, parentIDs)
}

// GetVertices returns all vertices in the DAG
func (s *ConsensusService) GetVertices() []*dag.Vertex {
	return s.avalanche.GetAllVertices()
}

// GetFinalizedVertices returns all finalized vertices
func (s *ConsensusService) GetFinalizedVertices() []*dag.Vertex {
	return s.avalanche.GetFinalized()
}

// IsVertexFinalized checks if a vertex is finalized
func (s *ConsensusService) IsVertexFinalized(id string) bool {
	return s.avalanche.IsFinalized(id)
}

// IsVertexPending checks if a vertex is pending
func (s *ConsensusService) IsVertexPending(id string) bool {
	return s.avalanche.IsPending(id)
}

// GetVertex retrieves a vertex by ID
func (s *ConsensusService) GetVertex(id string) (*dag.Vertex, error) {
	return s.avalanche.GetVertex(id)
} 