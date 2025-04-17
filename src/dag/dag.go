package dag

import (
	"sync"
)

// Vertex represents a vertex in the DAG
type Vertex struct {
	ID        string
	Data      interface{}
	Parents   map[string]*Vertex
	Children  map[string]*Vertex
	Preferred bool // Used in the avalanche consensus decision
	Color     int  // For coloring algorithm
	Finalized bool // Whether this vertex has been finalized
}

// DAG represents a Directed Acyclic Graph
type DAG struct {
	mu       sync.RWMutex
	vertices map[string]*Vertex
	roots    map[string]*Vertex // Vertices with no parents
}

// NewDAG creates a new DAG
func NewDAG() *DAG {
	return &DAG{
		vertices: make(map[string]*Vertex),
		roots:    make(map[string]*Vertex),
	}
}

// AddVertex adds a vertex to the DAG
func (d *DAG) AddVertex(id string, data interface{}) (*Vertex, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if vertex already exists
	if _, exists := d.vertices[id]; exists {
		return nil, ErrVertexAlreadyExists
	}

	v := &Vertex{
		ID:       id,
		Data:     data,
		Parents:  make(map[string]*Vertex),
		Children: make(map[string]*Vertex),
	}

	d.vertices[id] = v
	d.roots[id] = v // Initially, a new vertex is a root

	return v, nil
}

// AddEdge adds a directed edge from parent to child
func (d *DAG) AddEdge(parentID, childID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	parent, exists := d.vertices[parentID]
	if !exists {
		return ErrVertexNotFound
	}

	child, exists := d.vertices[childID]
	if !exists {
		return ErrVertexNotFound
	}

	// Check for cycles
	if d.wouldCreateCycle(parent, child) {
		return ErrWouldCreateCycle
	}

	// Add edge
	parent.Children[childID] = child
	child.Parents[parentID] = parent

	// Child is no longer a root
	delete(d.roots, childID)

	return nil
}

// wouldCreateCycle checks if adding an edge would create a cycle
func (d *DAG) wouldCreateCycle(parent, child *Vertex) bool {
	// Simple DFS to check if child is an ancestor of parent
	visited := make(map[string]bool)
	var dfs func(v *Vertex) bool
	dfs = func(v *Vertex) bool {
		if v.ID == parent.ID {
			return true
		}
		visited[v.ID] = true
		for _, p := range v.Parents {
			if !visited[p.ID] {
				if dfs(p) {
					return true
				}
			}
		}
		return false
	}
	return dfs(child)
}

// GetVertex retrieves a vertex by ID
func (d *DAG) GetVertex(id string) (*Vertex, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	v, exists := d.vertices[id]
	if !exists {
		return nil, ErrVertexNotFound
	}
	return v, nil
}

// RemoveVertex removes a vertex and all its edges
func (d *DAG) RemoveVertex(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	v, exists := d.vertices[id]
	if !exists {
		return ErrVertexNotFound
	}

	// Remove from children of its parents
	for pid, parent := range v.Parents {
		delete(parent.Children, id)
	}

	// Remove from parents of its children
	for cid, child := range v.Children {
		delete(child.Parents, id)
		// If the child has no other parents, it becomes a root
		if len(child.Parents) == 0 {
			d.roots[cid] = child
		}
	}

	// Remove from roots if it's a root
	delete(d.roots, id)

	// Remove the vertex
	delete(d.vertices, id)

	return nil
}

// GetRoots returns all root vertices
func (d *DAG) GetRoots() []*Vertex {
	d.mu.RLock()
	defer d.mu.RUnlock()

	roots := make([]*Vertex, 0, len(d.roots))
	for _, v := range d.roots {
		roots = append(roots, v)
	}
	return roots
}

// GetVertices returns all vertices
func (d *DAG) GetVertices() []*Vertex {
	d.mu.RLock()
	defer d.mu.RUnlock()

	vertices := make([]*Vertex, 0, len(d.vertices))
	for _, v := range d.vertices {
		vertices = append(vertices, v)
	}
	return vertices
}

// Errors
var (
	ErrVertexAlreadyExists = func() error { return &DAGError{message: "vertex already exists"} }()
	ErrVertexNotFound      = func() error { return &DAGError{message: "vertex not found"} }()
	ErrWouldCreateCycle    = func() error { return &DAGError{message: "operation would create a cycle"} }()
)

// DAGError represents an error in DAG operations
type DAGError struct {
	message string
}

func (e *DAGError) Error() string {
	return e.message
} 