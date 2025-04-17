package vertex

import (
	"time"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/dag"
)

// VertexData represents the data stored in a vertex
type VertexData struct {
	Content     interface{} `json:"content"`
	Creator     string      `json:"creator"`
	CreatedAt   time.Time   `json:"created_at"`
	Transaction string      `json:"transaction,omitempty"`
}

// NewVertexData creates a new vertex data object
func NewVertexData(content interface{}, creator string, transaction string) VertexData {
	return VertexData{
		Content:     content,
		Creator:     creator,
		CreatedAt:   time.Now(),
		Transaction: transaction,
	}
}

// VertexModel provides business logic for vertex operations
type VertexModel struct {
	// Could add additional fields like cache, validation rules, etc.
}

// NewVertexModel creates a new vertex model
func NewVertexModel() *VertexModel {
	return &VertexModel{}
}

// ConvertToResponse converts a DAG vertex to a response object
func (m *VertexModel) ConvertToResponse(vertex *dag.Vertex, isFinalized, isPending bool) VertexResponse {
	// Extract parent IDs
	parentIDs := make([]string, 0, len(vertex.Parents))
	for pid := range vertex.Parents {
		parentIDs = append(parentIDs, pid)
	}

	// Extract child IDs
	childIDs := make([]string, 0, len(vertex.Children))
	for cid := range vertex.Children {
		childIDs = append(childIDs, cid)
	}

	// Parse data as VertexData if possible
	var data VertexData
	if vd, ok := vertex.Data.(VertexData); ok {
		data = vd
	} else {
		// If data is not VertexData, create a minimal VertexData
		data = VertexData{
			Content:   vertex.Data,
			CreatedAt: time.Now(),
		}
	}

	return VertexResponse{
		ID:        vertex.ID,
		Data:      data,
		ParentIDs: parentIDs,
		ChildIDs:  childIDs,
		Finalized: isFinalized,
		Pending:   isPending,
	}
}

// ValidateVertex validates a vertex request
func (m *VertexModel) ValidateVertex(req VertexRequest) error {
	// Could add validation rules here
	// For example, check if ID is valid, data is not nil, etc.
	return nil
}

// VertexRequest represents a request to create a new vertex
type VertexRequest struct {
	ID        string      `json:"id"`
	Data      interface{} `json:"data"`
	ParentIDs []string    `json:"parent_ids"`
}

// VertexResponse represents a vertex response
type VertexResponse struct {
	ID        string     `json:"id"`
	Data      VertexData `json:"data"`
	ParentIDs []string   `json:"parent_ids"`
	ChildIDs  []string   `json:"child_ids"`
	Finalized bool       `json:"finalized"`
	Pending   bool       `json:"pending"`
} 