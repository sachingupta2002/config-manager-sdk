package repository

import (
	"context"

	"github.com/sachin/config-manager/internal/domain"
)

// AuditRepository defines the interface for audit log data access
type AuditRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	GetByID(ctx context.Context, id string) (*domain.AuditLog, error)
	ListByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.AuditLog, error)
	ListByConfigKey(ctx context.Context, configKeyID string, limit, offset int) ([]*domain.AuditLog, error)
	CountByService(ctx context.Context, serviceID string) (int, error)
}
