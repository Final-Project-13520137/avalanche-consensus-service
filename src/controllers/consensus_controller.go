package controllers

import (
	"net/http"
	"time"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/dag"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/views"
)

// ConsensusServiceInterface defines the interface for consensus operations
type ConsensusServiceInterface interface {
	ProposeVertex(id string, data interface{}, parentIDs []string) (*dag.Vertex, error)
	GetVertex(id string) (*dag.Vertex, error)
	GetVertices() []*dag.Vertex
	GetFinalizedVertices() []*dag.Vertex
	IsVertexFinalized(id string) bool
	IsVertexPending(id string) bool
	StartConsensus() error
	StopConsensus() error
}

// ConsensusController handles consensus-related requests
type ConsensusController struct {
	consensusService ConsensusServiceInterface
	responseBuilder  *views.ResponseBuilder
}

// NewConsensusController creates a new consensus controller
func NewConsensusController(consensusService ConsensusServiceInterface) *ConsensusController {
	return &ConsensusController{
		consensusService: consensusService,
		responseBuilder:  views.NewResponseBuilder(),
	}
}

// HandleStartConsensus handles starting the consensus algorithm
func (c *ConsensusController) HandleStartConsensus(w http.ResponseWriter, r *http.Request) {
	// Only POST is allowed
	if r.Method != http.MethodPost {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Start consensus
	if err := c.consensusService.StartConsensus(); err != nil {
		c.responseBuilder.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	c.responseBuilder.JSONResponse(w, map[string]string{
		"status":  "success",
		"message": "Consensus algorithm started",
	}, http.StatusOK)
}

// HandleStopConsensus handles stopping the consensus algorithm
func (c *ConsensusController) HandleStopConsensus(w http.ResponseWriter, r *http.Request) {
	// Only POST is allowed
	if r.Method != http.MethodPost {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Stop consensus
	if err := c.consensusService.StopConsensus(); err != nil {
		c.responseBuilder.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	c.responseBuilder.JSONResponse(w, map[string]string{
		"status":  "success",
		"message": "Consensus algorithm stopped",
	}, http.StatusOK)
}

// HandleConsensusStatus handles getting the status of the consensus algorithm
func (c *ConsensusController) HandleConsensusStatus(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get stats
	finalized := c.consensusService.GetFinalizedVertices()
	vertices := c.consensusService.GetVertices()

	// Build response
	response := struct {
		TotalVertices    int   `json:"total_vertices"`
		FinalizedCount   int   `json:"finalized_count"`
		PendingCount     int   `json:"pending_count"`
		TimestampSeconds int64 `json:"timestamp_seconds"`
	}{
		TotalVertices:    len(vertices),
		FinalizedCount:   len(finalized),
		PendingCount:     len(vertices) - len(finalized),
		TimestampSeconds: time.Now().Unix(),
	}

	// Return response
	c.responseBuilder.JSONResponse(w, response, http.StatusOK)
} 