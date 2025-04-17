package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/consensus"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/dag"
)

// Node represents a node in the simulated network
type Node struct {
	ID        string
	Avalanche *consensus.Avalanche
	Peers     map[string]*Node
	mu        sync.RWMutex
}

// NewNode creates a new node
func NewNode(id string, params consensus.AvalancheParams) *Node {
	d := dag.NewDAG()
	a := consensus.NewAvalanche(d, params)
	return &Node{
		ID:        id,
		Avalanche: a,
		Peers:     make(map[string]*Node),
	}
}

// AddPeer adds a peer to the node
func (n *Node) AddPeer(peer *Node) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.ID != peer.ID { // Don't add self as peer
		n.Peers[peer.ID] = peer
	}
}

// RemovePeer removes a peer from the node
func (n *Node) RemovePeer(peerID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.Peers, peerID)
}

// ProposeVertex proposes a new vertex to the network
func (n *Node) ProposeVertex(id string, data interface{}, parentIDs []string) (*dag.Vertex, error) {
	vertex, err := n.Avalanche.AddVertex(id, data, parentIDs)
	if err != nil {
		return nil, err
	}

	// In a real network, this would involve broadcasting to peers
	// For simulation, we'll directly notify peers
	for _, peer := range n.Peers {
		// In a real network, this would be an async network call
		go func(p *Node) {
			p.ReceiveVertex(id, data, parentIDs)
		}(peer)
	}

	return vertex, nil
}

// ReceiveVertex handles the receipt of a vertex from a peer
func (n *Node) ReceiveVertex(id string, data interface{}, parentIDs []string) {
	n.Avalanche.AddVertex(id, data, parentIDs)
}

// Start starts the consensus algorithm
func (n *Node) Start() chan struct{} {
	stop := make(chan struct{})
	go n.Avalanche.RunConsensus(stop)
	return stop
}

// Simulator represents a network simulator
type Simulator struct {
	Nodes map[string]*Node
	mu    sync.RWMutex
}

// NewSimulator creates a new simulator
func NewSimulator() *Simulator {
	return &Simulator{
		Nodes: make(map[string]*Node),
	}
}

// AddNode adds a node to the simulator
func (s *Simulator) AddNode(id string, params consensus.AvalancheParams) *Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	node := NewNode(id, params)
	s.Nodes[id] = node
	return node
}

// ConnectNodes connects all nodes in a full mesh topology
func (s *Simulator) ConnectNodes() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, node := range s.Nodes {
		for _, peer := range s.Nodes {
			if node.ID != peer.ID {
				node.AddPeer(peer)
			}
		}
	}
}

// DisconnectNode disconnects a node from the network
func (s *Simulator) DisconnectNode(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Remove node from all peers
	for _, node := range s.Nodes {
		node.RemovePeer(id)
	}
	// Remove node from simulator
	delete(s.Nodes, id)
}

// StartAll starts all nodes
func (s *Simulator) StartAll() map[string]chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	stops := make(map[string]chan struct{})
	for id, node := range s.Nodes {
		stops[id] = node.Start()
	}
	return stops
}

// StopAll stops all nodes
func (s *Simulator) StopAll(stops map[string]chan struct{}) {
	for _, stop := range stops {
		close(stop)
	}
}

// RunSimulation runs a simulation with the given parameters
func (s *Simulator) RunSimulation(numNodes int, duration time.Duration, vertexGenerator func(nodeID string, i int) (string, interface{}, []string)) {
	// Create nodes
	params := consensus.DefaultParams()
	for i := 0; i < numNodes; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		s.AddNode(nodeID, params)
	}

	// Connect nodes
	s.ConnectNodes()

	// Start consensus on all nodes
	stops := s.StartAll()

	// Create some vertices
	go func() {
		for i := 0; i < 100; i++ { // Generate 100 vertices
			for nodeID, node := range s.Nodes {
				vid, data, parents := vertexGenerator(nodeID, i)
				_, err := node.ProposeVertex(vid, data, parents)
				if err != nil {
					fmt.Printf("Error proposing vertex: %v\n", err)
				}
			}
			time.Sleep(100 * time.Millisecond) // Space out vertex creation
		}
	}()

	// Wait for the simulation to run
	time.Sleep(duration)

	// Stop all nodes
	s.StopAll(stops)

	// Print final stats
	for id, node := range s.Nodes {
		finalized := node.Avalanche.GetFinalized()
		fmt.Printf("Node %s finalized %d vertices\n", id, len(finalized))
	}
} 