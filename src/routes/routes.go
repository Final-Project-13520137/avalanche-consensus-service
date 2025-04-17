package routes

import (
	"net/http"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/controllers"
	"github.com/Final-Project-13520137/avalanche-consensus-service/src/middleware"
)

// Router sets up all the routes for the application
type Router struct {
	vertexController    *controllers.VertexController
	consensusController *controllers.ConsensusController
	peerController      *controllers.PeerController
	healthController    *controllers.HealthController
	loggingMiddleware   *middleware.LoggingMiddleware
}

// NewRouter creates a new router with the given controllers
func NewRouter(
	vertexController *controllers.VertexController,
	consensusController *controllers.ConsensusController,
	peerController *controllers.PeerController,
	healthController *controllers.HealthController,
) *Router {
	return &Router{
		vertexController:    vertexController,
		consensusController: consensusController,
		peerController:      peerController,
		healthController:    healthController,
		loggingMiddleware:   middleware.NewLoggingMiddleware(),
	}
}

// RegisterRoutes registers all routes with the given mux
func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	// Apply middleware to all routes
	withLogging := func(handler http.HandlerFunc) http.HandlerFunc {
		return r.loggingMiddleware.LogRequest(handler)
	}

	// Vertex endpoints
	mux.HandleFunc("/api/v1/vertex", withLogging(r.vertexController.HandleCreateVertex))
	mux.HandleFunc("/api/v1/vertex/", withLogging(r.vertexController.HandleGetVertex))
	mux.HandleFunc("/api/v1/vertices", withLogging(r.vertexController.HandleListVertices))
	mux.HandleFunc("/api/v1/vertices/finalized", withLogging(r.vertexController.HandleListFinalizedVertices))

	// Peer endpoints
	mux.HandleFunc("/api/v1/connect", withLogging(r.peerController.HandleConnect))
	mux.HandleFunc("/api/v1/peers", withLogging(r.peerController.HandleListPeers))
	mux.HandleFunc("/api/v1/peers/connect", withLogging(r.peerController.HandleConnectToPeers))

	// Consensus endpoints
	mux.HandleFunc("/api/v1/consensus/start", withLogging(r.consensusController.HandleStartConsensus))
	mux.HandleFunc("/api/v1/consensus/stop", withLogging(r.consensusController.HandleStopConsensus))
	mux.HandleFunc("/api/v1/consensus/status", withLogging(r.consensusController.HandleConsensusStatus))

	// Health check
	mux.HandleFunc("/health", withLogging(r.healthController.HandleHealthCheck))
} 