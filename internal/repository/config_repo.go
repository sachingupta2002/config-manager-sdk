package repository

import (
	"context"

	"github.com/sachin/config-manager/internal/domain"
)

// ConfigRepository defines the interface for config data access
type ConfigRepository interface {
	// ConfigKey operations
	CreateConfigKey(ctx context.Context, key *domain.ConfigKey) error
	GetConfigKeyByID(ctx context.Context, id string) (*domain.ConfigKey, error)
	GetConfigKeyByName(ctx context.Context, envID, keyName string) (*domain.ConfigKey, error)
	ListConfigKeysByEnvironment(ctx context.Context, envID string) ([]*domain.ConfigKey, error)
	UpdateConfigKey(ctx context.Context, key *domain.ConfigKey) error
	DeleteConfigKey(ctx context.Context, id string) error

	// ConfigVersion operations
	CreateConfigVersion(ctx context.Context, version *domain.ConfigVersion) error
	GetConfigVersionByID(ctx context.Context, id string) (*domain.ConfigVersion, error)
	GetConfigVersionsByKey(ctx context.Context, keyID string) ([]*domain.ConfigVersion, error)
	GetConfigVersionByNumber(ctx context.Context, keyID string, version int) (*domain.ConfigVersion, error)
	GetMaxVersion(ctx context.Context, keyID string) (int, error)
}
