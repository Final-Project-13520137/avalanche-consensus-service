package services

import (
	"fmt"
	"time"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/consensus"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/dag"
)

// SimulationService provides simulation functionality for the consensus algorithm
type SimulationService struct {
	consensus *consensus.Avalanche
}

// NewSimulationService creates a new simulation service
func NewSimulationService(consensus *consensus.Avalanche) *SimulationService {
	return &SimulationService{
		consensus: consensus,
	}
}

// RunRandomVertices generates random vertices and adds them to the consensus
func (s *SimulationService) RunRandomVertices(count int, maxParents int) []*dag.Vertex {
	result := make([]*dag.Vertex, 0, count)
	
	// Generate vertices
	for i := 0; i < count; i++ {
		// Generate a unique vertex ID
		id := fmt.Sprintf("vertex-%d", i)
		
		// Data can be any transaction data in a real implementation
		data := fmt.Sprintf("Transaction-%d", i)
		
		// For the first vertex, don't have parents
		parents := []string{}
		
		// For subsequent vertices, reference random previous vertices
		if i > 0 {
			// Add up to maxParents previous vertices as parents
			for j := 1; j <= maxParents && i-j >= 0; j++ {
				parentID := fmt.Sprintf("vertex-%d", i-j)
				parents = append(parents, parentID)
			}
		}
		
		// Add vertex to consensus
		vertex, err := s.consensus.AddVertex(id, data, parents)
		if err != nil {
			fmt.Printf("Error adding vertex: %v\n", err)
			continue
		}
		
		result = append(result, vertex)
		
		// Space out vertex creation to simulate real-world scenarios
		time.Sleep(50 * time.Millisecond)
	}
	
	return result
}

// SimulateNetworkDelay simulates network delay by sleeping
func (s *SimulationService) SimulateNetworkDelay(minMS, maxMS int) {
	// In a real implementation, this would use a random delay between minMS and maxMS
	time.Sleep(time.Duration(minMS) * time.Millisecond)
}