package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// PeerService handles communication with other peers in the network
type PeerService struct {
	mu            sync.RWMutex
	nodeID        string
	peers         map[string]string // Map of peer ID to address
	client        *http.Client
	receiveVertex func(id string, data interface{}, parentIDs []string) error
}

// VertexMessage represents a vertex message for network transmission
type VertexMessage struct {
	ID        string      `json:"id"`
	Data      interface{} `json:"data"`
	ParentIDs []string    `json:"parent_ids"`
	SenderID  string      `json:"sender_id"`
}

// NewPeerService creates a new peer service
func NewPeerService(nodeID string, receiveFunc func(id string, data interface{}, parentIDs []string) error) *PeerService {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	return &PeerService{
		nodeID:        nodeID,
		peers:         make(map[string]string),
		client:        client,
		receiveVertex: receiveFunc,
	}
}

// SetReceiveVertexFunc sets the function to handle receiving vertices
func (p *PeerService) SetReceiveVertexFunc(receiveFunc func(id string, data interface{}, parentIDs []string) error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.receiveVertex = receiveFunc
}

// AddPeer adds a peer to the network
func (p *PeerService) AddPeer(peerID, address string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers[peerID] = address
}

// RemovePeer removes a peer from the network
func (p *PeerService) RemovePeer(peerID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.peers, peerID)
}

// GetPeers returns all peers in the network
func (p *PeerService) GetPeers() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	peerIDs := make([]string, 0, len(p.peers))
	for id := range p.peers {
		peerIDs = append(peerIDs, id)
	}
	
	return peerIDs
}

// ConnectToPeers connects to a list of peer addresses
func (p *PeerService) ConnectToPeers(peerAddresses []string) error {
	for _, addr := range peerAddresses {
		// Send connect request to peer
		resp, err := p.client.Get(addr + "/api/v1/connect?nodeID=" + p.nodeID)
		if err != nil {
			fmt.Printf("Error connecting to peer %s: %v\n", addr, err)
			continue
		}
		defer resp.Body.Close()
		
		// Parse response
		var peerInfo struct {
			NodeID string `json:"node_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&peerInfo); err != nil {
			fmt.Printf("Error parsing peer info: %v\n", err)
			continue
		}
		
		// Add peer
		p.AddPeer(peerInfo.NodeID, addr)
	}
	
	return nil
}

// BroadcastVertex broadcasts a vertex to all peers
func (p *PeerService) BroadcastVertex(id string, data interface{}, parentIDs []string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	// Create vertex message
	msg := VertexMessage{
		ID:        id,
		Data:      data,
		ParentIDs: parentIDs,
		SenderID:  p.nodeID,
	}
	
	// Marshal to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	// Send to all peers
	for peerID, addr := range p.peers {
		go func(id, address string) {
			resp, err := p.client.Post(address+"/api/v1/vertex", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("Error sending vertex to peer %s: %v\n", id, err)
				return
			}
			defer resp.Body.Close()
		}(peerID, addr)
	}
	
	return nil
}

// HandleVertexRequest handles incoming vertex requests
func (p *PeerService) HandleVertexRequest(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var msg VertexMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Process vertex
	if err := p.receiveVertex(msg.ID, msg.Data, msg.ParentIDs); err != nil {
		http.Error(w, fmt.Sprintf("Error processing vertex: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Add sender as peer if not already known
	if _, exists := p.peers[msg.SenderID]; !exists {
		host := r.RemoteAddr
		p.AddPeer(msg.SenderID, "http://"+host)
	}
	
	w.WriteHeader(http.StatusOK)
}

// HandleConnectRequest handles incoming connect requests
func (p *PeerService) HandleConnectRequest(w http.ResponseWriter, r *http.Request) {
	// Get peer ID from query params
	peerID := r.URL.Query().Get("nodeID")
	if peerID == "" {
		http.Error(w, "Missing nodeID parameter", http.StatusBadRequest)
		return
	}
	
	// Add peer
	host := r.RemoteAddr
	p.AddPeer(peerID, "http://"+host)
	
	// Return our node ID
	response := struct {
		NodeID string `json:"node_id"`
	}{
		NodeID: p.nodeID,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
} 