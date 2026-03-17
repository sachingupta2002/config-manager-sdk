package http

import (
	"encoding/json"
	"net/http"

	"github.com/sachin/config-manager/internal/dto"
	"github.com/sachin/config-manager/internal/service"
)

// ServiceHandler handles HTTP requests for services
type ServiceHandler struct {
	serviceService *service.ServiceService
}

// NewServiceHandler creates a new ServiceHandler instance
func NewServiceHandler(serviceService *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{
		serviceService: serviceService,
	}
}

// CreateService handles POST /api/v1/services
func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	svc, err := h.serviceService.CreateService(r.Context(), req.Name)
	if err != nil {
		if err == service.ErrServiceAlreadyExists {
			writeError(w, http.StatusConflict, "already_exists", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create service")
		return
	}

	resp := dto.ServiceResponse{
		ID:        svc.ID,
		Name:      svc.Name,
		CreatedAt: svc.CreatedAt,
		UpdatedAt: svc.UpdatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetService handles GET /api/v1/services/{id}
func (h *ServiceHandler) GetService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Service ID is required")
		return
	}

	svc, err := h.serviceService.GetService(r.Context(), id)
	if err != nil {
		if err == service.ErrServiceNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get service")
		return
	}

	resp := dto.ServiceResponse{
		ID:        svc.ID,
		Name:      svc.Name,
		CreatedAt: svc.CreatedAt,
		UpdatedAt: svc.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListServices handles GET /api/v1/services
func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.serviceService.ListServices(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list services")
		return
	}

	var serviceResponses []dto.ServiceResponse
	for _, s := range services {
		serviceResponses = append(serviceResponses, dto.ServiceResponse{
			ID:        s.ID,
			Name:      s.Name,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}

	resp := dto.ServiceListResponse{
		Services: serviceResponses,
		Total:    len(serviceResponses),
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateService handles PUT /api/v1/services/{id}
func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Service ID is required")
		return
	}

	var req dto.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	svc, err := h.serviceService.UpdateService(r.Context(), id, req.Name)
	if err != nil {
		if err == service.ErrServiceNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		if err == service.ErrServiceAlreadyExists {
			writeError(w, http.StatusConflict, "already_exists", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update service")
		return
	}

	resp := dto.ServiceResponse{
		ID:        svc.ID,
		Name:      svc.Name,
		CreatedAt: svc.CreatedAt,
		UpdatedAt: svc.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteService handles DELETE /api/v1/services/{id}
func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Service ID is required")
		return
	}

	err := h.serviceService.DeleteService(r.Context(), id)
	if err != nil {
		if err == service.ErrServiceNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete service")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
