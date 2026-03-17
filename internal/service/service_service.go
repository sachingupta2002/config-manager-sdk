package service

import (
	"context"
	"errors"

	"github.com/sachin/config-manager/internal/domain"
	"github.com/sachin/config-manager/internal/repository"
)

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrServiceAlreadyExists = errors.New("service with this name already exists")
)

// ServiceService handles service-related business logic
type ServiceService struct {
	serviceRepo repository.ServiceRepository
}

// NewServiceService creates a new ServiceService instance
func NewServiceService(serviceRepo repository.ServiceRepository) *ServiceService {
	return &ServiceService{
		serviceRepo: serviceRepo,
	}
}

// CreateService creates a new service
func (s *ServiceService) CreateService(ctx context.Context, name string) (*domain.Service, error) {
	// Check if service already exists
	existing, _ := s.serviceRepo.GetByName(ctx, name)
	if existing != nil {
		return nil, ErrServiceAlreadyExists
	}

	svc := &domain.Service{
		Name: name,
	}

	if err := s.serviceRepo.Create(ctx, svc); err != nil {
		return nil, err
	}

	return svc, nil
}

// GetService retrieves a service by ID
func (s *ServiceService) GetService(ctx context.Context, id string) (*domain.Service, error) {
	svc, err := s.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServiceNotFound
	}
	return svc, nil
}

// GetServiceByName retrieves a service by name
func (s *ServiceService) GetServiceByName(ctx context.Context, name string) (*domain.Service, error) {
	svc, err := s.serviceRepo.GetByName(ctx, name)
	if err != nil {
		return nil, ErrServiceNotFound
	}
	return svc, nil
}

// ListServices returns all services
func (s *ServiceService) ListServices(ctx context.Context) ([]*domain.Service, error) {
	return s.serviceRepo.List(ctx)
}

// UpdateService updates a service
func (s *ServiceService) UpdateService(ctx context.Context, id, name string) (*domain.Service, error) {
	svc, err := s.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	// Check if another service with the same name exists
	existing, _ := s.serviceRepo.GetByName(ctx, name)
	if existing != nil && existing.ID != id {
		return nil, ErrServiceAlreadyExists
	}

	svc.Name = name
	if err := s.serviceRepo.Update(ctx, svc); err != nil {
		return nil, err
	}

	return svc, nil
}

// DeleteService deletes a service
func (s *ServiceService) DeleteService(ctx context.Context, id string) error {
	_, err := s.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return ErrServiceNotFound
	}

	return s.serviceRepo.Delete(ctx, id)
}
