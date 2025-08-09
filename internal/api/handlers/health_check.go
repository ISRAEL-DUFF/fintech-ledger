package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// HealthCheckResponse represents the health check response structure
// swagger:response healthCheckResponse
type HealthCheckResponse struct {
	// The status of the service
	// Example: healthy
	Status string `json:"status"`
	
	// The version of the service
	// Example: 1.0.0
	Version string `json:"version"`
}

// HealthCheckHandler handles health check requests
// @Summary Health check endpoint
// @Description Returns the health status of the API
// @Tags health
// @Accept  json
// @Produce  json
// @Success 200 {object} HealthCheckResponse "OK"
// @Router /health [get]
// @Router /api/v1/health [get]
func HealthCheckHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, HealthCheckResponse{
			Status:  "healthy",
			Version: version,
		})
	}
}

// RegisterHealthRoutes registers health check routes
func RegisterHealthRoutes(router chi.Router, version string) {
	router.Get("/health", HealthCheckHandler(version))
	router.Get("/api/v1/health", HealthCheckHandler(version))
}
