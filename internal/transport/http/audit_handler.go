package http

import (
	"net/http"
	"strconv"

	"github.com/sachin/config-manager/internal/dto"
	"github.com/sachin/config-manager/internal/service"
)

// AuditHandler handles HTTP requests for audit logs
type AuditHandler struct {
	auditService *service.AuditService
}

// NewAuditHandler creates a new AuditHandler instance
func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLog handles GET /api/v1/audit/{id}
func (h *AuditHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Audit log ID is required")
		return
	}

	log, err := h.auditService.GetAuditLog(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Audit log not found")
		return
	}

	resp := dto.AuditLogResponse{
		ID:          log.ID,
		ServiceID:   log.ServiceID,
		ConfigKeyID: log.ConfigKeyID,
		Action:      log.Action,
		OldValue:    log.OldValue,
		NewValue:    log.NewValue,
		PerformedBy: log.PerformedBy,
		PerformedAt: log.PerformedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListAuditLogsByService handles GET /api/v1/services/{serviceId}/audit
func (h *AuditHandler) ListAuditLogsByService(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("serviceId")
	if serviceID == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Service ID is required")
		return
	}

	// Parse pagination params
	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, total, err := h.auditService.ListAuditLogsByService(r.Context(), serviceID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list audit logs")
		return
	}

	var logResponses []dto.AuditLogResponse
	for _, l := range logs {
		logResponses = append(logResponses, dto.AuditLogResponse{
			ID:          l.ID,
			ServiceID:   l.ServiceID,
			ConfigKeyID: l.ConfigKeyID,
			Action:      l.Action,
			OldValue:    l.OldValue,
			NewValue:    l.NewValue,
			PerformedBy: l.PerformedBy,
			PerformedAt: l.PerformedAt,
		})
	}

	resp := dto.AuditLogListResponse{
		AuditLogs: logResponses,
		Total:     total,
	}

	writeJSON(w, http.StatusOK, resp)
}
