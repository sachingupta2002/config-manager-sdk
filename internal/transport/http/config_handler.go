package http

import (
	"encoding/json"
	"net/http"

	"github.com/sachin/config-manager/internal/dto"
	"github.com/sachin/config-manager/internal/service"
)

// ConfigHandler handles HTTP requests for configs
type ConfigHandler struct {
	configService   *service.ConfigService
	rollbackService *service.RollbackService
}

// NewConfigHandler creates a new ConfigHandler instance
func NewConfigHandler(configService *service.ConfigService, rollbackService *service.RollbackService) *ConfigHandler {
	return &ConfigHandler{
		configService:   configService,
		rollbackService: rollbackService,
	}
}

// GetConfig handles GET /api/v1/configs/{envId}/{key}
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("envId")
	key := r.PathValue("key")

	if envID == "" || key == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID and key are required")
		return
	}

	config, err := h.configService.GetConfig(r.Context(), envID, key)
	if err != nil {
		if err == service.ErrConfigKeyNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get config")
		return
	}

	resp := dto.ConfigResponse{
		ID:              config.ID,
		EnvironmentID:   config.EnvironmentID,
		Key:             config.Key,
		Value:           config.Value,
		ValueType:       config.ValueType,
		ActiveVersionID: config.ActiveVersionID,
		CreatedAt:       config.CreatedAt,
		UpdatedAt:       config.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// SetConfig handles POST /api/v1/configs/{envId}/{key}
func (h *ConfigHandler) SetConfig(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("envId")
	key := r.PathValue("key")

	if envID == "" || key == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID and key are required")
		return
	}

	var req dto.SetConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Default value type to string if not provided
	valueType := req.ValueType
	if valueType == "" {
		valueType = "string"
	}

	// Get performed_by from request or default
	performedBy := req.PerformedBy
	if performedBy == "" {
		performedBy = "system" // TODO: Extract from auth context
	}

	config, err := h.configService.SetConfig(r.Context(), envID, key, req.Value, valueType, performedBy)
	if err != nil {
		if err == service.ErrEnvironmentNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to set config")
		return
	}

	resp := dto.ConfigResponse{
		ID:              config.ID,
		EnvironmentID:   config.EnvironmentID,
		Key:             config.Key,
		Value:           config.Value,
		ValueType:       config.ValueType,
		ActiveVersionID: config.ActiveVersionID,
		CreatedAt:       config.CreatedAt,
		UpdatedAt:       config.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListConfigs handles GET /api/v1/configs/{envId}
func (h *ConfigHandler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("envId")
	if envID == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID is required")
		return
	}

	configs, err := h.configService.ListConfigs(r.Context(), envID)
	if err != nil {
		if err == service.ErrEnvironmentNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list configs")
		return
	}

	var configResponses []dto.ConfigResponse
	for _, c := range configs {
		configResponses = append(configResponses, dto.ConfigResponse{
			ID:              c.ID,
			EnvironmentID:   c.EnvironmentID,
			Key:             c.Key,
			Value:           c.Value,
			ValueType:       c.ValueType,
			ActiveVersionID: c.ActiveVersionID,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
		})
	}

	resp := dto.ConfigListResponse{
		Configs: configResponses,
		Total:   len(configResponses),
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteConfig handles DELETE /api/v1/configs/{envId}/{key}
func (h *ConfigHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("envId")
	key := r.PathValue("key")

	if envID == "" || key == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID and key are required")
		return
	}

	// First get the config to get its ID
	config, err := h.configService.GetConfig(r.Context(), envID, key)
	if err != nil {
		if err == service.ErrConfigKeyNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get config")
		return
	}

	performedBy := "system" // TODO: Extract from auth context

	err = h.configService.DeleteConfig(r.Context(), config.ID, performedBy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete config")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetVersionHistory handles GET /api/v1/configs/{envId}/{key}/versions
func (h *ConfigHandler) GetVersionHistory(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("envId")
	key := r.PathValue("key")

	if envID == "" || key == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID and key are required")
		return
	}

	// First get the config to get its ID
	config, err := h.configService.GetConfig(r.Context(), envID, key)
	if err != nil {
		if err == service.ErrConfigKeyNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get config")
		return
	}

	versions, err := h.configService.GetVersionHistory(r.Context(), config.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get version history")
		return
	}

	var versionResponses []dto.VersionResponse
	for _, v := range versions {
		versionResponses = append(versionResponses, dto.VersionResponse{
			ID:        v.ID,
			Version:   v.Version,
			Value:     v.Value,
			CreatedAt: v.CreatedAt,
			CreatedBy: v.CreatedBy,
		})
	}

	resp := dto.VersionHistoryResponse{
		Versions: versionResponses,
		Total:    len(versionResponses),
	}

	writeJSON(w, http.StatusOK, resp)
}

// RollbackConfig handles POST /api/v1/configs/{envId}/{key}/rollback
func (h *ConfigHandler) RollbackConfig(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("envId")
	key := r.PathValue("key")

	if envID == "" || key == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Environment ID and key are required")
		return
	}

	var req dto.RollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Version <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "Version must be positive")
		return
	}

	// First get the config to get its ID
	config, err := h.configService.GetConfig(r.Context(), envID, key)
	if err != nil {
		if err == service.ErrConfigKeyNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get config")
		return
	}

	performedBy := req.PerformedBy
	if performedBy == "" {
		performedBy = "system" // TODO: Extract from auth context
	}

	updatedConfig, err := h.rollbackService.RollbackToVersion(r.Context(), config.ID, req.Version, performedBy)
	if err != nil {
		if err == service.ErrVersionNotFound {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to rollback config")
		return
	}

	resp := dto.ConfigResponse{
		ID:              updatedConfig.ID,
		EnvironmentID:   updatedConfig.EnvironmentID,
		Key:             updatedConfig.Key,
		Value:           updatedConfig.Value,
		ValueType:       updatedConfig.ValueType,
		ActiveVersionID: updatedConfig.ActiveVersionID,
		CreatedAt:       updatedConfig.CreatedAt,
		UpdatedAt:       updatedConfig.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}
