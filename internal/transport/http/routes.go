package http

import (
	"net/http"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	Service     *ServiceHandler
	Environment *EnvironmentHandler
	Config      *ConfigHandler
	Audit       *AuditHandler
}

// SetupRoutes configures all HTTP routes
func SetupRoutes(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", HealthCheck)

	// Service endpoints
	mux.HandleFunc("POST /api/v1/services", h.Service.CreateService)
	mux.HandleFunc("GET /api/v1/services", h.Service.ListServices)
	mux.HandleFunc("GET /api/v1/services/{id}", h.Service.GetService)
	mux.HandleFunc("PUT /api/v1/services/{id}", h.Service.UpdateService)
	mux.HandleFunc("DELETE /api/v1/services/{id}", h.Service.DeleteService)

	// Environment endpoints
	mux.HandleFunc("POST /api/v1/services/{serviceId}/environments", h.Environment.CreateEnvironment)
	mux.HandleFunc("GET /api/v1/services/{serviceId}/environments", h.Environment.ListEnvironments)
	mux.HandleFunc("GET /api/v1/environments/{id}", h.Environment.GetEnvironment)
	mux.HandleFunc("PUT /api/v1/environments/{id}", h.Environment.UpdateEnvironment)
	mux.HandleFunc("DELETE /api/v1/environments/{id}", h.Environment.DeleteEnvironment)

	// Config endpoints
	mux.HandleFunc("GET /api/v1/configs/{envId}", h.Config.ListConfigs)
	mux.HandleFunc("GET /api/v1/configs/{envId}/{key}", h.Config.GetConfig)
	mux.HandleFunc("POST /api/v1/configs/{envId}/{key}", h.Config.SetConfig)
	mux.HandleFunc("DELETE /api/v1/configs/{envId}/{key}", h.Config.DeleteConfig)

	// Version and rollback endpoints
	mux.HandleFunc("GET /api/v1/configs/{envId}/{key}/versions", h.Config.GetVersionHistory)
	mux.HandleFunc("POST /api/v1/configs/{envId}/{key}/rollback", h.Config.RollbackConfig)

	// Audit endpoints
	mux.HandleFunc("GET /api/v1/audit/{id}", h.Audit.GetAuditLog)
	mux.HandleFunc("GET /api/v1/services/{serviceId}/audit", h.Audit.ListAuditLogsByService)

	// Apply middleware
	var handler http.Handler = mux
	handler = LoggingMiddleware(handler)
	handler = CORSMiddleware(handler)

	return handler
}
