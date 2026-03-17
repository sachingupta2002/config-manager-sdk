package http

import (
	"encoding/json"
	"net/http"

	"github.com/sachin/config-manager/internal/dto"
	"github.com/sachin/config-manager/internal/service"
)

// EnvironmentHandler handles HTTP requests for environments
type EnvironmentHandler struct {
	envService *service.EnvironmentService
}

// NewEnvironmentHandler creates a new EnvironmentHandler instance
func NewEnvironmentHandler(envService *service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{
		envService: envService,
	}
}

// CreateEnvironment handles POST /api/v1/services/{serviceId}/environments
func (h *EnvironmentHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("serviceId")
	if serviceID == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Service ID is required")
		return
	}

	var req dto.CreateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	env, err := h.envService.CreateEnvironment(r.Context(), serviceID, req.Name)
	if err != nil {
		if err == service.ErrServiceNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		if err == service.ErrEnvironmentAlreadyExists {
			writeError(w, http.StatusConflict, "already_exists", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create environment")
		return
	}

	resp := dto.EnvironmentResponse{
		ID:        env.ID,
		ServiceID: env.ServiceID,
		Name:      env.Name,
		CreatedAt: env.CreatedAt,
		UpdatedAt: env.UpdatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetEnvironment handles GET /api/v1/environments/{id}
func (h *EnvironmentHandler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID is required")
		return
	}

	env, err := h.envService.GetEnvironment(r.Context(), id)
	if err != nil {
		if err == service.ErrEnvironmentNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get environment")
		return
	}

	resp := dto.EnvironmentResponse{
		ID:        env.ID,
		ServiceID: env.ServiceID,
		Name:      env.Name,
		CreatedAt: env.CreatedAt,
		UpdatedAt: env.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListEnvironments handles GET /api/v1/services/{serviceId}/environments
func (h *EnvironmentHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("serviceId")
	if serviceID == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Service ID is required")
		return
	}

	envs, err := h.envService.ListEnvironments(r.Context(), serviceID)
	if err != nil {
		if err == service.ErrServiceNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list environments")
		return
	}

	var envResponses []dto.EnvironmentResponse
	for _, e := range envs {
		envResponses = append(envResponses, dto.EnvironmentResponse{
			ID:        e.ID,
			ServiceID: e.ServiceID,
			Name:      e.Name,
			CreatedAt: e.CreatedAt,
			UpdatedAt: e.UpdatedAt,
		})
	}

	resp := dto.EnvironmentListResponse{
		Environments: envResponses,
		Total:        len(envResponses),
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateEnvironment handles PUT /api/v1/environments/{id}
func (h *EnvironmentHandler) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID is required")
		return
	}

	var req dto.UpdateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	env, err := h.envService.UpdateEnvironment(r.Context(), id, req.Name)
	if err != nil {
		if err == service.ErrEnvironmentNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		if err == service.ErrEnvironmentAlreadyExists {
			writeError(w, http.StatusConflict, "already_exists", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update environment")
		return
	}

	resp := dto.EnvironmentResponse{
		ID:        env.ID,
		ServiceID: env.ServiceID,
		Name:      env.Name,
		CreatedAt: env.CreatedAt,
		UpdatedAt: env.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteEnvironment handles DELETE /api/v1/environments/{id}
func (h *EnvironmentHandler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID is required")
		return
	}

	err := h.envService.DeleteEnvironment(r.Context(), id)
	if err != nil {
		if err == service.ErrEnvironmentNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete environment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
