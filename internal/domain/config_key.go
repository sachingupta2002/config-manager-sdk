package domain

import "time"

// ConfigKey represents a configuration key-value pair
type ConfigKey struct {
	ID              string      `json:"id"`
	EnvironmentID   string      `json:"environment_id"`
	Key             string      `json:"key"`
	Value           interface{} `json:"value"`      // Denormalized: cached active value (typed based on ValueType)
	ValueType       string      `json:"value_type"` // string, int, bool, json
	ActiveVersionID *string     `json:"active_version_id,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}
