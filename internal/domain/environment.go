package domain

import "time"

// Environment represents a deployment environment (e.g., dev, staging, prod) and also the service to whihc it depends 
type Environment struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"service_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
