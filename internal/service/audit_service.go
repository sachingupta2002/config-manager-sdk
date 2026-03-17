package service

import (
	"context"

	"github.com/sachin/config-manager/internal/domain"
	"github.com/sachin/config-manager/internal/repository"
)

// AuditService handles audit log business logic
type AuditService struct {
	auditRepo repository.AuditRepository
}

// NewAuditService creates a new AuditService instance
func NewAuditService(auditRepo repository.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// GetAuditLog retrieves an audit log by ID
func (s *AuditService) GetAuditLog(ctx context.Context, id string) (*domain.AuditLog, error) {
	return s.auditRepo.GetByID(ctx, id)
}

// ListAuditLogsByService returns audit logs for a service with pagination
func (s *AuditService) ListAuditLogsByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.AuditLog, int, error) {
	logs, err := s.auditRepo.ListByService(ctx, serviceID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.auditRepo.CountByService(ctx, serviceID)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListAuditLogsByConfigKey returns audit logs for a specific config key
func (s *AuditService) ListAuditLogsByConfigKey(ctx context.Context, configKeyID string, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditRepo.ListByConfigKey(ctx, configKeyID, limit, offset)
}
