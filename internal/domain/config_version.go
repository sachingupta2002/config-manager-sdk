package domain

import "time"

// ConfigVersion represents a historical version of a config key
type ConfigVersion struct {
	ID          string      `json:"id"`
	ConfigKeyID string      `json:"config_key_id"`
	Value       interface{} `json:"value"` // Typed value based on config key's ValueType
	Version     int         `json:"version"`
	CreatedAt   time.Time   `json:"created_at"`
	CreatedBy   string      `json:"created_by"`
}
