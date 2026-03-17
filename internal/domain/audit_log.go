package domain

import "time"

// AuditLog represents an audit trail entry for config changes
type AuditLog struct {
	ID          string      `json:"id"`
	ServiceID   string      `json:"service_id"`
	ConfigKeyID string      `json:"config_key_id"`
	Action      string      `json:"action"` // create, update, delete, rollback
	OldValue    interface{} `json:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value,omitempty"`
	PerformedBy string      `json:"performed_by"` // this will be extracted from the auth context further down the line. by default it will be the user ID
	PerformedAt time.Time   `json:"performed_at"` // current time when the action was performed
}
