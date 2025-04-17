package views

import (
	"encoding/json"
	"net/http"
)

// ResponseBuilder builds HTTP responses
type ResponseBuilder struct{}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{}
}

// JSONResponse sends a JSON response with the given status code
func (b *ResponseBuilder) JSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")
	
	// Set status code
	w.WriteHeader(statusCode)
	
	// Encode data as JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// ErrorResponse sends an error response with the given message and status code
func (b *ResponseBuilder) ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	// Create error response
	errorResponse := struct {
		Error   string `json:"error"`
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		Error:   http.StatusText(statusCode),
		Status:  statusCode,
		Message: message,
	}
	
	// Send JSON response
	b.JSONResponse(w, errorResponse, statusCode)
} 