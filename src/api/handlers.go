package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/service"
)

// API handles HTTP requests for the consensus service
type API struct {
	consensusService *service.ConsensusService
	peerService      *service.PeerService
}

// NewAPI creates a new API handler
func NewAPI(consensusService *service.ConsensusService, peerService *service.PeerService) *API {
	return &API{
		consensusService: consensusService,
		peerService:      peerService,
	}
}

// RegisterHandlers registers API handlers with the router
func (a *API) RegisterHandlers(mux *http.ServeMux) {
	// Vertex endpoints
	mux.HandleFunc("/api/v1/vertex", a.handleVertex)
	mux.HandleFunc("/api/v1/vertex/", a.handleVertexByID)
	mux.HandleFunc("/api/v1/vertices", a.handleListVertices)
	mux.HandleFunc("/api/v1/vertices/finalized", a.handleListFinalized)

	// Peer endpoints
	mux.HandleFunc("/api/v1/connect", a.handleConnect)
	mux.HandleFunc("/api/v1/peers", a.handleListPeers)

	// Consensus endpoints
	mux.HandleFunc("/api/v1/consensus/start", a.handleStartConsensus)
	mux.HandleFunc("/api/v1/consensus/stop", a.handleStopConsensus)
	mux.HandleFunc("/api/v1/consensus/status", a.handleConsensusStatus)

	// Health check
	mux.HandleFunc("/health", a.handleHealth)
}

// VertexRequest represents a request to create a new vertex
type VertexRequest struct {
	ID        string      `json:"id"`
	Data      interface{} `json:"data"`
	ParentIDs []string    `json:"parent_ids"`
}

// VertexResponse represents a vertex response
type VertexResponse struct {
	ID        string      `json:"id"`
	Data      interface{} `json:"data"`
	ParentIDs []string    `json:"parent_ids"`
	Finalized bool        `json:"finalized"`
	Pending   bool        `json:"pending"`
}

// handleVertex handles POST and GET requests for vertices
func (a *API) handleVertex(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle vertex submission - same handler as peer service
		a.peerService.HandleVertexRequest(w, r)
		return
	}

	// Method not allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// List vertices (redirect to /vertices)
	a.handleListVertices(w, r)
}

// handleVertexByID handles requests for a specific vertex
func (a *API) handleVertexByID(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract vertex ID from URL
	id := r.URL.Path[len("/api/v1/vertex/"):]
	if id == "" {
		http.Error(w, "Vertex ID required", http.StatusBadRequest)
		return
	}

	// Get vertex
	vertex, err := a.consensusService.GetVertex(id)
	if err != nil {
		http.Error(w, "Vertex not found", http.StatusNotFound)
		return
	}

	// Build response
	parentIDs := make([]string, 0, len(vertex.Parents))
	for pid := range vertex.Parents {
		parentIDs = append(parentIDs, pid)
	}

	response := VertexResponse{
		ID:        vertex.ID,
		Data:      vertex.Data,
		ParentIDs: parentIDs,
		Finalized: vertex.Finalized,
		Pending:   a.consensusService.IsVertexPending(id),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListVertices handles requests to list all vertices
func (a *API) handleListVertices(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all vertices
	vertices := a.consensusService.GetVertices()

	// Build response
	responses := make([]VertexResponse, 0, len(vertices))
	for _, v := range vertices {
		parentIDs := make([]string, 0, len(v.Parents))
		for pid := range v.Parents {
			parentIDs = append(parentIDs, pid)
		}

		responses = append(responses, VertexResponse{
			ID:        v.ID,
			Data:      v.Data,
			ParentIDs: parentIDs,
			Finalized: v.Finalized,
			Pending:   a.consensusService.IsVertexPending(v.ID),
		})
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// handleListFinalized handles requests to list all finalized vertices
func (a *API) handleListFinalized(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get finalized vertices
	vertices := a.consensusService.GetFinalizedVertices()

	// Build response
	responses := make([]VertexResponse, 0, len(vertices))
	for _, v := range vertices {
		parentIDs := make([]string, 0, len(v.Parents))
		for pid := range v.Parents {
			parentIDs = append(parentIDs, pid)
		}

		responses = append(responses, VertexResponse{
			ID:        v.ID,
			Data:      v.Data,
			ParentIDs: parentIDs,
			Finalized: true,
			Pending:   false,
		})
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// handleConnect handles peer connection requests
func (a *API) handleConnect(w http.ResponseWriter, r *http.Request) {
	// This is already implemented in the peer service
	a.peerService.HandleConnectRequest(w, r)
}

// handleListPeers handles requests to list all peers
func (a *API) handleListPeers(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all peers
	peers := a.peerService.GetPeers()

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// handleStartConsensus handles requests to start the consensus algorithm
func (a *API) handleStartConsensus(w http.ResponseWriter, r *http.Request) {
	// Only POST is allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Start consensus
	if err := a.consensusService.StartConsensus(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "consensus started"})
}

// handleStopConsensus handles requests to stop the consensus algorithm
func (a *API) handleStopConsensus(w http.ResponseWriter, r *http.Request) {
	// Only POST is allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Stop consensus
	if err := a.consensusService.StopConsensus(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "consensus stopped"})
}

// handleConsensusStatus handles requests for consensus status
func (a *API) handleConsensusStatus(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get stats
	finalized := a.consensusService.GetFinalizedVertices()
	vertices := a.consensusService.GetVertices()

	// Build response
	response := struct {
		TotalVertices    int `json:"total_vertices"`
		FinalizedCount   int `json:"finalized_count"`
		PendingCount     int `json:"pending_count"`
		TimestampSeconds int64 `json:"timestamp_seconds"`
	}{
		TotalVertices:    len(vertices),
		FinalizedCount:   len(finalized),
		PendingCount:     len(vertices) - len(finalized),
		TimestampSeconds: time.Now().Unix(),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles health check requests
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Build response
	response := struct {
		Status    string `json:"status"`
		Timestamp int64  `json:"timestamp"`
	}{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
} 