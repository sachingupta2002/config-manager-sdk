package service

import (
	"context"
	"errors"

	"github.com/sachin/config-manager/internal/domain"
	"github.com/sachin/config-manager/internal/repository"
)

var (
	ErrVersionNotFound = errors.New("version not found")
)

// RollbackService handles config rollback business logic
type RollbackService struct {
	configRepo repository.ConfigRepository
	envRepo    repository.EnvironmentRepository
	auditRepo  repository.AuditRepository
}

// NewRollbackService creates a new RollbackService instance
func NewRollbackService(
	configRepo repository.ConfigRepository,
	envRepo repository.EnvironmentRepository,
	auditRepo repository.AuditRepository,
) *RollbackService {
	return &RollbackService{
		configRepo: configRepo,
		envRepo:    envRepo,
		auditRepo:  auditRepo,
	}
}

// RollbackToVersion rolls back a config key to a specific version
func (s *RollbackService) RollbackToVersion(ctx context.Context, keyID string, version int, performedBy string) (*domain.ConfigKey, error) {
	// Get the config key
	configKey, err := s.configRepo.GetConfigKeyByID(ctx, keyID)
	if err != nil {
		return nil, ErrConfigKeyNotFound
	}

	// Get environment for service ID
	env, err := s.envRepo.GetByID(ctx, configKey.EnvironmentID)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	// Get the target version
	targetVersion, err := s.configRepo.GetConfigVersionByNumber(ctx, keyID, version)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	oldValue := configKey.Value

	// Update config key to point to target version
	configKey.Value = targetVersion.Value
	configKey.ActiveVersionID = &targetVersion.ID

	if err := s.configRepo.UpdateConfigKey(ctx, configKey); err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		ServiceID:   env.ServiceID,
		ConfigKeyID: configKey.ID,
		Action:      "rollback",
		OldValue:    oldValue,
		NewValue:    targetVersion.Value,
		PerformedBy: performedBy,
	}
	s.auditRepo.Create(ctx, auditLog)

	return configKey, nil
}

// GetVersionHistory returns the version history for a config key
func (s *RollbackService) GetVersionHistory(ctx context.Context, keyID string) ([]*domain.ConfigVersion, error) {
	_, err := s.configRepo.GetConfigKeyByID(ctx, keyID)
	if err != nil {
		return nil, ErrConfigKeyNotFound
	}

	return s.configRepo.GetConfigVersionsByKey(ctx, keyID)
}
