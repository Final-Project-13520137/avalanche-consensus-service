package controllers

import (
	"net/http"
	"time"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/views"
)

// HealthController handles health check requests
type HealthController struct {
	responseBuilder *views.ResponseBuilder
}

// NewHealthController creates a new health controller
func NewHealthController() *HealthController {
	return &HealthController{
		responseBuilder: views.NewResponseBuilder(),
	}
}

// HandleHealthCheck handles health check requests
func (c *HealthController) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Only GET is allowed
	if r.Method != http.MethodGet {
		c.responseBuilder.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create response
	response := struct {
		Status    string `json:"status"`
		Timestamp int64  `json:"timestamp"`
		Message   string `json:"message"`
	}{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
		Message:   "Service is running",
	}

	// Return response
	c.responseBuilder.JSONResponse(w, response, http.StatusOK)
} 