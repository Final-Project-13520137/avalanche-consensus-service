package consensus

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/dag"
)

// Parameters for the Avalanche consensus
type AvalancheParams struct {
	K              int           // Sample size (number of vertices to query)
	Alpha          int           // Threshold for decision making
	BetaVirtuous   int           // Confidence threshold for virtuous vertices
	BetaRogue      int           // Confidence threshold for rogue vertices
	ConcurrencyNum int           // Number of concurrent requests
	BatchSize      int           // Number of vertices to process in a batch
	MaxOutstanding int           // Maximum number of outstanding operations
	MaxSampleSize  int           // Maximum sample size per operation
	SampleTimeout  time.Duration // Timeout for a single sample query
}

// Default params
func DefaultParams() AvalancheParams {
	return AvalancheParams{
		K:              10,         // Query 10 validators
		Alpha:          8,          // Require 80% supermajority (8/10) for decisions
		BetaVirtuous:   20,         // Require 20 consecutive successful queries for finality (virtuous vertices)
		BetaRogue:      30,         // Require 30 consecutive successful queries for finality (conflicting vertices)
		ConcurrencyNum: 4,          // Allow 4 concurrent ops
		BatchSize:      10,         // Process 10 vertices in a batch
		MaxOutstanding: 1024,       // Max 1024 outstanding vertices
		MaxSampleSize:  20,         // Sample at most 20 validators
		SampleTimeout:  time.Second, // 1s timeout for sample queries
	}
}

// Avalanche implements the Avalanche consensus protocol
type Avalanche struct {
	mu       sync.RWMutex
	dag      *dag.DAG       // The underlying DAG data structure
	params   AvalancheParams // Protocol parameters
	pending  map[string]int  // Map from vertex ID to confidence count
	finalized map[string]bool // Vertices that have been finalized
}

// NewAvalanche creates a new Avalanche instance with the given parameters
func NewAvalanche(d *dag.DAG, params AvalancheParams) *Avalanche {
	return &Avalanche{
		dag:      d,
		params:   params,
		pending:  make(map[string]int),
		finalized: make(map[string]bool),
	}
}

// AddVertex adds a new vertex to the consensus mechanism
func (a *Avalanche) AddVertex(id string, data interface{}, parentIDs []string) (*dag.Vertex, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Add vertex to DAG
	vertex, err := a.dag.AddVertex(id, data)
	if err != nil {
		return nil, err
	}

	// Connect to parents
	for _, pid := range parentIDs {
		if err := a.dag.AddEdge(pid, id); err != nil {
			// Rollback on error
			a.dag.RemoveVertex(id)
			return nil, err
		}
	}

	// Add to pending set for consensus
	a.pending[id] = 0

	return vertex, nil
}

// RunConsensus starts the consensus algorithm
func (a *Avalanche) RunConsensus(stop <-chan struct{}) {
	// Run consensus in a loop until stopped
	for {
		select {
		case <-stop:
			return
		default:
			a.consensusRound()
			time.Sleep(10 * time.Millisecond) // Prevent CPU overuse
		}
	}
}

// consensusRound performs one round of the consensus algorithm
func (a *Avalanche) consensusRound() {
	a.mu.Lock()
	// Make a copy of pending to avoid long lock times
	pending := make([]string, 0, len(a.pending))
	for id := range a.pending {
		pending = append(pending, id)
	}
	a.mu.Unlock()

	// Process each pending vertex
	for _, id := range pending {
		a.processVertex(id)
	}
}

// processVertex processes a single vertex
func (a *Avalanche) processVertex(id string) {
	a.mu.RLock()
	// Skip if already finalized
	if a.finalized[id] {
		a.mu.RUnlock()
		return
	}
	currentCount := a.pending[id]
	a.mu.RUnlock()

	// Get k random vertices to query (preferably from parents)
	samples := a.getSamples(id, a.params.K)
	if len(samples) == 0 {
		return // Not enough samples available
	}

	// Query the samples for their preference
	// In a real implementation, this would involve network calls
	// Here we simulate it by checking local preferences
	preferCount := 0
	for _, sampleID := range samples {
		if a.checkPreference(sampleID, id) {
			preferCount++
		}
	}

	// Update confidence if we reached Alpha majority
	if preferCount >= a.params.Alpha {
		a.mu.Lock()
		a.pending[id] = currentCount + 1

		// Check if we've reached confidence threshold
		threshold := a.getConfidenceThreshold(id)
		if a.pending[id] >= threshold {
			// Finalize vertex
			a.finalized[id] = true
			delete(a.pending, id)

			// Mark vertex as finalized in DAG
			if v, err := a.dag.GetVertex(id); err == nil {
				v.Finalized = true
			}
		}
		a.mu.Unlock()
	} else {
		// Reset confidence counter on failure
		a.mu.Lock()
		a.pending[id] = 0
		a.mu.Unlock()
	}
}

// getSamples returns k random vertices to query
func (a *Avalanche) getSamples(id string, k int) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Get all vertices
	allVertices := a.dag.GetVertices()
	if len(allVertices) < k {
		return nil // Not enough vertices for sampling
	}

	// Prioritize parents (in a real implementation, this would prioritize validators)
	vertex, err := a.dag.GetVertex(id)
	if err != nil {
		return nil
	}

	// Build candidate list - parents first, then others
	candidates := make([]string, 0, len(allVertices))
	for pid := range vertex.Parents {
		candidates = append(candidates, pid)
	}

	// Add other vertices that aren't parents or the vertex itself
	for _, v := range allVertices {
		if v.ID != id && vertex.Parents[v.ID] == nil {
			candidates = append(candidates, v.ID)
		}
	}

	// Randomly select k samples
	if len(candidates) <= k {
		return candidates
	}

	// Fisher-Yates shuffle to randomly select k elements
	samples := make([]string, len(candidates))
	copy(samples, candidates)
	for i := len(samples) - 1; i > 0; i-- {
		// Generate a random index between 0 and i
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		// Swap elements at i and j
		samples[i], samples[int(j.Int64())] = samples[int(j.Int64())], samples[i]
	}

	return samples[:k]
}

// checkPreference checks if a vertex prefers another vertex
// In a real implementation, this would involve querying other nodes
func (a *Avalanche) checkPreference(sampleID, targetID string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// In a simple implementation, we'll say a vertex prefers another if:
	// 1. It's already finalized
	// 2. It's a direct or indirect parent
	// 3. Or by a random choice with bias towards consensus

	// Check if the sample vertex exists
	sampleVertex, err := a.dag.GetVertex(sampleID)
	if err != nil {
		return false
	}

	targetVertex, err := a.dag.GetVertex(targetID)
	if err != nil {
		return false
	}

	// If the target is already finalized, prefer it
	if targetVertex.Finalized {
		return true
	}

	// Check if target is a parent (direct or indirect) of the sample
	isParent := false
	visited := make(map[string]bool)
	var checkParent func(v *dag.Vertex) bool
	checkParent = func(v *dag.Vertex) bool {
		if v.ID == targetID {
			return true
		}
		visited[v.ID] = true
		for pid, parent := range v.Parents {
			if !visited[pid] {
				if checkParent(parent) {
					return true
				}
			}
		}
		return false
	}
	isParent = checkParent(sampleVertex)
	if isParent {
		return true
	}

	// Use the vertex's preferred flag if set
	if sampleVertex.Preferred {
		return true
	}

	// For conflicting vertices, make a biased random choice
	// In practice, nodes would make this decision based on their local state
	r, _ := rand.Int(rand.Reader, big.NewInt(100))
	return r.Int64() < 70 // 70% chance to prefer, biasing towards consensus
}

// getConfidenceThreshold returns the confidence threshold for a vertex
func (a *Avalanche) getConfidenceThreshold(id string) int {
	// Check if this vertex is "virtuous" (no conflicts)
	// In a real implementation, this would check for conflicting transactions
	isVirtuous := true
	v, err := a.dag.GetVertex(id)
	if err != nil {
		return a.params.BetaRogue // Default to higher threshold on error
	}

	// Check for conflicts (simplified version)
	// In a real implementation, this would be based on transaction conflicts
	for _, other := range a.dag.GetVertices() {
		if v.ID != other.ID && !a.areCompatible(v, other) {
			isVirtuous = false
			break
		}
	}

	if isVirtuous {
		return a.params.BetaVirtuous
	}
	return a.params.BetaRogue
}

// areCompatible determines if two vertices are compatible
// In a real implementation, this would check transaction conflicts
func (a *Avalanche) areCompatible(v1, v2 *dag.Vertex) bool {
	// For this example, we'll consider all vertices compatible unless they have the same data
	// In a real implementation, this would involve transaction conflict rules
	return v1.Data != v2.Data
}

// GetFinalized returns all finalized vertices
func (a *Avalanche) GetFinalized() []*dag.Vertex {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*dag.Vertex, 0, len(a.finalized))
	for id := range a.finalized {
		if v, err := a.dag.GetVertex(id); err == nil {
			result = append(result, v)
		}
	}
	return result
}

// IsPending checks if a vertex is still pending consensus
func (a *Avalanche) IsPending(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, isPending := a.pending[id]
	return isPending
}

// IsFinalized checks if a vertex has been finalized
func (a *Avalanche) IsFinalized(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.finalized[id]
}

// GetVertex retrieves a vertex by ID
func (a *Avalanche) GetVertex(id string) (*dag.Vertex, error) {
	return a.dag.GetVertex(id)
}

// GetAllVertices returns all vertices in the DAG
func (a *Avalanche) GetAllVertices() []*dag.Vertex {
	return a.dag.GetVertices()
} 