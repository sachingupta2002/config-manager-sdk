package service

import (
	"context"
	"errors"

	"github.com/sachin/config-manager/internal/domain"
	"github.com/sachin/config-manager/internal/repository"
)

var (
	ErrConfigKeyNotFound      = errors.New("config key not found")
	ErrConfigKeyAlreadyExists = errors.New("config key already exists in this environment")
)

// ConfigService handles config-related business logic
type ConfigService struct {
	configRepo repository.ConfigRepository
	envRepo    repository.EnvironmentRepository
	auditRepo  repository.AuditRepository
}

// NewConfigService creates a new ConfigService instance
func NewConfigService(
	configRepo repository.ConfigRepository,
	envRepo repository.EnvironmentRepository,
	auditRepo repository.AuditRepository,
) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
		envRepo:    envRepo,
		auditRepo:  auditRepo,
	}
}

// GetConfig retrieves a config value by key
func (s *ConfigService) GetConfig(ctx context.Context, envID, key string) (*domain.ConfigKey, error) {
	configKey, err := s.configRepo.GetConfigKeyByName(ctx, envID, key)
	if err != nil {
		return nil, ErrConfigKeyNotFound
	}
	return configKey, nil
}

// GetConfigByID retrieves a config by ID
func (s *ConfigService) GetConfigByID(ctx context.Context, id string) (*domain.ConfigKey, error) {
	configKey, err := s.configRepo.GetConfigKeyByID(ctx, id)
	if err != nil {
		return nil, ErrConfigKeyNotFound
	}
	return configKey, nil
}

// SetConfig creates or updates a config key
func (s *ConfigService) SetConfig(ctx context.Context, envID, key string, value interface{}, valueType, performedBy string) (*domain.ConfigKey, error) {
	// Verify environment exists and get service ID for audit
	env, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	// Infer value type if not provided
	if valueType == "" {
		valueType = domain.InferValueType(value)
	}

	// Check if config key exists
	existing, err := s.configRepo.GetConfigKeyByName(ctx, envID, key)
	if err != nil {
		// Create new config key
		return s.createConfig(ctx, env, key, value, valueType, performedBy)
	}

	// Update existing config key
	return s.updateConfig(ctx, env, existing, value, valueType, performedBy)
}

// createConfig creates a new config key with initial version
func (s *ConfigService) createConfig(ctx context.Context, env *domain.Environment, key string, value interface{}, valueType, performedBy string) (*domain.ConfigKey, error) {
	// Create config key
	configKey := &domain.ConfigKey{
		EnvironmentID: env.ID,
		Key:           key,
		Value:         value,
		ValueType:     valueType,
	}

	if err := s.configRepo.CreateConfigKey(ctx, configKey); err != nil {
		return nil, err
	}

	// Create first version
	version := &domain.ConfigVersion{
		ConfigKeyID: configKey.ID,
		Value:       value,
		Version:     1,
		CreatedBy:   performedBy,
	}

	if err := s.configRepo.CreateConfigVersion(ctx, version); err != nil {
		return nil, err
	}

	// Update config key with active version
	configKey.ActiveVersionID = &version.ID
	if err := s.configRepo.UpdateConfigKey(ctx, configKey); err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		ServiceID:   env.ServiceID,
		ConfigKeyID: configKey.ID,
		Action:      "create",
		NewValue:    value,
		PerformedBy: performedBy,
	}
	s.auditRepo.Create(ctx, auditLog)

	return configKey, nil
}

// updateConfig updates an existing config key with new version
func (s *ConfigService) updateConfig(ctx context.Context, env *domain.Environment, configKey *domain.ConfigKey, value interface{}, valueType, performedBy string) (*domain.ConfigKey, error) {
	oldValue := configKey.Value

	// Get next version number
	maxVersion, err := s.configRepo.GetMaxVersion(ctx, configKey.ID)
	if err != nil {
		return nil, err
	}

	// Create new version
	version := &domain.ConfigVersion{
		ConfigKeyID: configKey.ID,
		Value:       value,
		Version:     maxVersion + 1,
		CreatedBy:   performedBy,
	}

	if err := s.configRepo.CreateConfigVersion(ctx, version); err != nil {
		return nil, err
	}

	// Update config key
	configKey.Value = value
	configKey.ValueType = valueType
	configKey.ActiveVersionID = &version.ID

	if err := s.configRepo.UpdateConfigKey(ctx, configKey); err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		ServiceID:   env.ServiceID,
		ConfigKeyID: configKey.ID,
		Action:      "update",
		OldValue:    oldValue,
		NewValue:    value,
		PerformedBy: performedBy,
	}
	s.auditRepo.Create(ctx, auditLog)

	return configKey, nil
}

// ListConfigs returns all configs for an environment
func (s *ConfigService) ListConfigs(ctx context.Context, envID string) ([]*domain.ConfigKey, error) {
	// Verify environment exists
	_, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	return s.configRepo.ListConfigKeysByEnvironment(ctx, envID)
}

// DeleteConfig removes a config key
func (s *ConfigService) DeleteConfig(ctx context.Context, keyID, performedBy string) error {
	configKey, err := s.configRepo.GetConfigKeyByID(ctx, keyID)
	if err != nil {
		return ErrConfigKeyNotFound
	}

	// Get environment for service ID
	env, err := s.envRepo.GetByID(ctx, configKey.EnvironmentID)
	if err != nil {
		return ErrEnvironmentNotFound
	}

	// Create audit log before deletion
	auditLog := &domain.AuditLog{
		ServiceID:   env.ServiceID,
		ConfigKeyID: configKey.ID,
		Action:      "delete",
		OldValue:    configKey.Value,
		PerformedBy: performedBy,
	}
	s.auditRepo.Create(ctx, auditLog)

	return s.configRepo.DeleteConfigKey(ctx, keyID)
}

// GetVersionHistory returns the version history for a config key
func (s *ConfigService) GetVersionHistory(ctx context.Context, keyID string) ([]*domain.ConfigVersion, error) {
	_, err := s.configRepo.GetConfigKeyByID(ctx, keyID)
	if err != nil {
		return nil, ErrConfigKeyNotFound
	}

	return s.configRepo.GetConfigVersionsByKey(ctx, keyID)
}
