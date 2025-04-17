package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/views"
)

// PeerServiceInterface defines the interface for peer operations
type PeerServiceInterface interface {
	ConnectToPeers(peers []string) error
	GetPeers() []string
	BroadcastVertex(id string, data interface{}, parentIDs []string) error
	HandleVertexRequest(w http.ResponseWriter, r *http.Request)
	HandleConnectRequest(w http.ResponseWriter, r *http.Request)
}

// PeerController handles peer-related requests
type PeerController struct {
	peerService     PeerServiceInterface
	responseBuilder *views.ResponseBuilder
}

// NewPeerController creates a new peer controller
func NewPeerController(peerService PeerServiceInterface) *PeerController {
	return &PeerController{
		peerService:     peerService,
		responseBuilder: views.NewResponseBuilder(),
	}
}

// HandleConnect handles peer connection requests
func (c *PeerController) HandleConnect(w http.ResponseWriter, r *http.Request) {
	// This is delegated to the peer service
	c.peerService.HandleConnectRequest(w, r)
}

// HandleListPeers handles listing all peers
func (c *PeerController) HandleListPeers(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all peers
	peers := c.peerService.GetPeers()

	// Create response
	response := struct {
		Peers []string `json:"peers"`
		Count int      `json:"count"`
	}{
		Peers: peers,
		Count: len(peers),
	}

	// Return response
	c.responseBuilder.JSONResponse(w, response, http.StatusOK)
}

// HandleReceiveVertex handles receiving a vertex from a peer
func (c *PeerController) HandleReceiveVertex(w http.ResponseWriter, r *http.Request) {
	// This is delegated to the peer service
	c.peerService.HandleVertexRequest(w, r)
}

// HandleConnectToPeers handles connecting to a list of peers
func (c *PeerController) HandleConnectToPeers(w http.ResponseWriter, r *http.Request) {
	// Only POST is allowed
	if r.Method != http.MethodPost {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Peers []string `json:"peers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.responseBuilder.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Connect to peers
	if err := c.peerService.ConnectToPeers(req.Peers); err != nil {
		c.responseBuilder.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	c.responseBuilder.JSONResponse(w, map[string]string{
		"status":  "success",
		"message": "Connected to peers",
	}, http.StatusOK)
} 