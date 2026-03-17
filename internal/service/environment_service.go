package service

import (
	"context"
	"errors"

	"github.com/sachin/config-manager/internal/domain"
	"github.com/sachin/config-manager/internal/repository"
)

var (
	ErrEnvironmentNotFound      = errors.New("environment not found")
	ErrEnvironmentAlreadyExists = errors.New("environment with this name already exists for this service")
)

// EnvironmentService handles environment-related business logic
type EnvironmentService struct {
	envRepo     repository.EnvironmentRepository
	serviceRepo repository.ServiceRepository
}

// NewEnvironmentService creates a new EnvironmentService instance
func NewEnvironmentService(envRepo repository.EnvironmentRepository, serviceRepo repository.ServiceRepository) *EnvironmentService {
	return &EnvironmentService{
		envRepo:     envRepo,
		serviceRepo: serviceRepo,
	}
}

// CreateEnvironment creates a new environment for a service
func (s *EnvironmentService) CreateEnvironment(ctx context.Context, serviceID, name string) (*domain.Environment, error) {
	// Verify service exists
	_, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	// Check if environment already exists for this service
	existing, _ := s.envRepo.GetByServiceAndName(ctx, serviceID, name)
	if existing != nil {
		return nil, ErrEnvironmentAlreadyExists
	}

	env := &domain.Environment{
		ServiceID: serviceID,
		Name:      name,
	}

	if err := s.envRepo.Create(ctx, env); err != nil {
		return nil, err
	}

	return env, nil
}

// GetEnvironment retrieves an environment by ID
func (s *EnvironmentService) GetEnvironment(ctx context.Context, id string) (*domain.Environment, error) {
	env, err := s.envRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}
	return env, nil
}

// GetEnvironmentByName retrieves an environment by service ID and name
func (s *EnvironmentService) GetEnvironmentByName(ctx context.Context, serviceID, name string) (*domain.Environment, error) {
	env, err := s.envRepo.GetByServiceAndName(ctx, serviceID, name)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}
	return env, nil
}

// ListEnvironments returns all environments for a service
func (s *EnvironmentService) ListEnvironments(ctx context.Context, serviceID string) ([]*domain.Environment, error) {
	// Verify service exists
	_, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	return s.envRepo.ListByService(ctx, serviceID)
}

// UpdateEnvironment updates an environment
func (s *EnvironmentService) UpdateEnvironment(ctx context.Context, id, name string) (*domain.Environment, error) {
	env, err := s.envRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	// Check if another environment with the same name exists for this service
	existing, _ := s.envRepo.GetByServiceAndName(ctx, env.ServiceID, name)
	if existing != nil && existing.ID != id {
		return nil, ErrEnvironmentAlreadyExists
	}

	env.Name = name
	if err := s.envRepo.Update(ctx, env); err != nil {
		return nil, err
	}

	return env, nil
}

// DeleteEnvironment deletes an environment
func (s *EnvironmentService) DeleteEnvironment(ctx context.Context, id string) error {
	_, err := s.envRepo.GetByID(ctx, id)
	if err != nil {
		return ErrEnvironmentNotFound
	}

	return s.envRepo.Delete(ctx, id)
}
