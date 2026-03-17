package dto

import "time"

// ServiceResponse represents a service response
type ServiceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ServiceListResponse represents a list of services
type ServiceListResponse struct {
	Services []ServiceResponse `json:"services"`
	Total    int               `json:"total"`
}

// EnvironmentResponse represents an environment response
type EnvironmentResponse struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"service_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EnvironmentListResponse represents a list of environments
type EnvironmentListResponse struct {
	Environments []EnvironmentResponse `json:"environments"`
	Total        int                   `json:"total"`
}

// ConfigResponse represents a config key response
type ConfigResponse struct {
	ID              string      `json:"id"`
	EnvironmentID   string      `json:"environment_id"`
	Key             string      `json:"key"`
	Value           interface{} `json:"value"`
	ValueType       string      `json:"value_type"`
	ActiveVersionID *string     `json:"active_version_id,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// ConfigListResponse represents a list of configs
type ConfigListResponse struct {
	Configs []ConfigResponse `json:"configs"`
	Total   int              `json:"total"`
}

// VersionResponse represents a config version
type VersionResponse struct {
	ID        string      `json:"id"`
	Version   int         `json:"version"`
	Value     interface{} `json:"value"`
	CreatedAt time.Time   `json:"created_at"`
	CreatedBy string      `json:"created_by"`
}

// VersionHistoryResponse represents version history
type VersionHistoryResponse struct {
	Versions []VersionResponse `json:"versions"`
	Total    int               `json:"total"`
}

// AuditLogResponse represents an audit log entry
type AuditLogResponse struct {
	ID          string      `json:"id"`
	ServiceID   string      `json:"service_id"`
	ConfigKeyID string      `json:"config_key_id"`
	Action      string      `json:"action"`
	OldValue    interface{} `json:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value,omitempty"`
	PerformedBy string      `json:"performed_by"`
	PerformedAt time.Time   `json:"performed_at"`
}

// AuditLogListResponse represents a list of audit logs
type AuditLogListResponse struct {
	AuditLogs []AuditLogResponse `json:"audit_logs"`
	Total     int                `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
