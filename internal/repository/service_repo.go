package repository

import (
	"context"

	"github.com/sachin/config-manager/internal/domain"
)

// ServiceRepository defines the interface for service data access
type ServiceRepository interface {
	Create(ctx context.Context, service *domain.Service) error
	GetByID(ctx context.Context, id string) (*domain.Service, error)
	GetByName(ctx context.Context, name string) (*domain.Service, error)
	List(ctx context.Context) ([]*domain.Service, error)
	Update(ctx context.Context, service *domain.Service) error
	Delete(ctx context.Context, id string) error
}

// EnvironmentRepository defines the interface for environment data access
type EnvironmentRepository interface {
	Create(ctx context.Context, env *domain.Environment) error
	GetByID(ctx context.Context, id string) (*domain.Environment, error)
	GetByServiceAndName(ctx context.Context, serviceID, name string) (*domain.Environment, error)
	ListByService(ctx context.Context, serviceID string) ([]*domain.Environment, error)
	Update(ctx context.Context, env *domain.Environment) error
	Delete(ctx context.Context, id string) error
}
