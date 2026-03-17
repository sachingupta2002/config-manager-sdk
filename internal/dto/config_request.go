package dto

// CreateServiceRequest represents a request to create a new service
type CreateServiceRequest struct {
	Name string `json:"name"`
}

// UpdateServiceRequest represents a request to update a service
type UpdateServiceRequest struct {
	Name string `json:"name"`
}

// CreateEnvironmentRequest represents a request to create a new environment
type CreateEnvironmentRequest struct {
	Name string `json:"name"`
}

// UpdateEnvironmentRequest represents a request to update an environment
type UpdateEnvironmentRequest struct {
	Name string `json:"name"`
}

// SetConfigRequest represents a request to set a config value
type SetConfigRequest struct {
	Value       interface{} `json:"value"`
	ValueType   string      `json:"value_type,omitempty"` // string, int, bool, json
	PerformedBy string      `json:"performed_by,omitempty"`
}

// RollbackRequest represents a request to rollback to a specific version
type RollbackRequest struct {
	Version     int    `json:"version"`
	PerformedBy string `json:"performed_by,omitempty"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
