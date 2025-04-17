package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/vertex"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/views"
)

// VertexController handles vertex-related requests
type VertexController struct {
	consensusService ConsensusServiceInterface
	vertexModel      *vertex.VertexModel
	responseBuilder  *views.ResponseBuilder
}

// NewVertexController creates a new vertex controller
func NewVertexController(consensusService ConsensusServiceInterface) *VertexController {
	return &VertexController{
		consensusService: consensusService,
		vertexModel:      vertex.NewVertexModel(),
		responseBuilder:  views.NewResponseBuilder(),
	}
}

// HandleCreateVertex handles creation of a new vertex
func (c *VertexController) HandleCreateVertex(w http.ResponseWriter, r *http.Request) {
	// Only POST is allowed
	if r.Method != http.MethodPost {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req vertex.VertexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.responseBuilder.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := c.vertexModel.ValidateVertex(req); err != nil {
		c.responseBuilder.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create vertex
	v, err := c.consensusService.ProposeVertex(req.ID, req.Data, req.ParentIDs)
	if err != nil {
		c.responseBuilder.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create response
	response := c.vertexModel.ConvertToResponse(
		v,
		c.consensusService.IsVertexFinalized(v.ID),
		c.consensusService.IsVertexPending(v.ID),
	)

	// Return response
	c.responseBuilder.JSONResponse(w, response, http.StatusCreated)
}

// HandleGetVertex handles fetching a vertex by ID
func (c *VertexController) HandleGetVertex(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract vertex ID from URL
	path := r.URL.Path
	parts := strings.Split(path, "/")
	id := parts[len(parts)-1]

	if id == "" || id == "vertex" {
		c.responseBuilder.ErrorResponse(w, "Vertex ID required", http.StatusBadRequest)
		return
	}

	// Get vertex from service
	v, err := c.consensusService.GetVertex(id)
	if err != nil {
		c.responseBuilder.ErrorResponse(w, "Vertex not found", http.StatusNotFound)
		return
	}

	// Create response
	response := c.vertexModel.ConvertToResponse(
		v,
		c.consensusService.IsVertexFinalized(v.ID),
		c.consensusService.IsVertexPending(v.ID),
	)

	// Return response
	c.responseBuilder.JSONResponse(w, response, http.StatusOK)
}

// HandleListVertices handles listing all vertices
func (c *VertexController) HandleListVertices(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all vertices
	vertices := c.consensusService.GetVertices()

	// Convert to response objects
	responses := make([]vertex.VertexResponse, 0, len(vertices))
	for _, v := range vertices {
		response := c.vertexModel.ConvertToResponse(
			v,
			c.consensusService.IsVertexFinalized(v.ID),
			c.consensusService.IsVertexPending(v.ID),
		)
		responses = append(responses, response)
	}

	// Return response
	c.responseBuilder.JSONResponse(w, responses, http.StatusOK)
}

// HandleListFinalizedVertices handles listing all finalized vertices
func (c *VertexController) HandleListFinalizedVertices(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get finalized vertices
	vertices := c.consensusService.GetFinalizedVertices()

	// Convert to response objects
	responses := make([]vertex.VertexResponse, 0, len(vertices))
	for _, v := range vertices {
		response := c.vertexModel.ConvertToResponse(
			v,
			true,  // isFinalized
			false, // isPending
		)
		responses = append(responses, response)
	}

	// Return response
	c.responseBuilder.JSONResponse(w, responses, http.StatusOK)
} 