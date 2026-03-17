package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/sachin/config-manager/internal/domain"
)

type ConfigRepositoryImpl struct {
	db *sql.DB
}

func NewConfigRepository(db *sql.DB) *ConfigRepositoryImpl {
	return &ConfigRepositoryImpl{db: db}
}

// ConfigKey operations

func (r *ConfigRepositoryImpl) CreateConfigKey(ctx context.Context, key *domain.ConfigKey) error {
	query := `
		INSERT INTO config_keys (environment_id, key, value, value_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	key.CreatedAt = now
	key.UpdatedAt = now

	// Serialize value to string for database storage
	serializedValue, err := domain.SerializeValue(key.Value)
	if err != nil {
		return err
	}

	return r.db.QueryRowContext(ctx, query,
		key.EnvironmentID, key.Key, serializedValue, key.ValueType, key.CreatedAt, key.UpdatedAt,
	).Scan(&key.ID)
}

func (r *ConfigRepositoryImpl) GetConfigKeyByID(ctx context.Context, id string) (*domain.ConfigKey, error) {
	query := `
		SELECT id, environment_id, key, value, value_type, active_version_id, created_at, updated_at 
		FROM config_keys WHERE id = $1`

	key := &domain.ConfigKey{}
	var rawValue string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&key.ID, &key.EnvironmentID, &key.Key, &rawValue, &key.ValueType,
		&key.ActiveVersionID, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse value from string to appropriate type
	key.Value, err = domain.ParseValue(rawValue, key.ValueType)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (r *ConfigRepositoryImpl) GetConfigKeyByName(ctx context.Context, envID, keyName string) (*domain.ConfigKey, error) {
	query := `
		SELECT id, environment_id, key, value, value_type, active_version_id, created_at, updated_at 
		FROM config_keys WHERE environment_id = $1 AND key = $2`

	key := &domain.ConfigKey{}
	var rawValue string
	err := r.db.QueryRowContext(ctx, query, envID, keyName).Scan(
		&key.ID, &key.EnvironmentID, &key.Key, &rawValue, &key.ValueType,
		&key.ActiveVersionID, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse value from string to appropriate type
	key.Value, err = domain.ParseValue(rawValue, key.ValueType)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (r *ConfigRepositoryImpl) ListConfigKeysByEnvironment(ctx context.Context, envID string) ([]*domain.ConfigKey, error) {
	query := `
		SELECT id, environment_id, key, value, value_type, active_version_id, created_at, updated_at 
		FROM config_keys WHERE environment_id = $1 ORDER BY key`

	rows, err := r.db.QueryContext(ctx, query, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.ConfigKey
	for rows.Next() {
		key := &domain.ConfigKey{}
		var rawValue string
		if err := rows.Scan(
			&key.ID, &key.EnvironmentID, &key.Key, &rawValue, &key.ValueType,
			&key.ActiveVersionID, &key.CreatedAt, &key.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Parse value from string to appropriate type
		key.Value, err = domain.ParseValue(rawValue, key.ValueType)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (r *ConfigRepositoryImpl) UpdateConfigKey(ctx context.Context, key *domain.ConfigKey) error {
	query := `
		UPDATE config_keys 
		SET value = $1, value_type = $2, active_version_id = $3, updated_at = $4
		WHERE id = $5`

	key.UpdatedAt = time.Now()

	// Serialize value to string for database storage
	serializedValue, err := domain.SerializeValue(key.Value)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, serializedValue, key.ValueType, key.ActiveVersionID, key.UpdatedAt, key.ID)
	return err
}

func (r *ConfigRepositoryImpl) DeleteConfigKey(ctx context.Context, id string) error {
	query := `DELETE FROM config_keys WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ConfigVersion operations

func (r *ConfigRepositoryImpl) CreateConfigVersion(ctx context.Context, version *domain.ConfigVersion) error {
	query := `
		INSERT INTO config_versions (config_key_id, value, version, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	version.CreatedAt = time.Now()

	// Serialize value to string for database storage
	serializedValue, err := domain.SerializeValue(version.Value)
	if err != nil {
		return err
	}

	return r.db.QueryRowContext(ctx, query,
		version.ConfigKeyID, serializedValue, version.Version, version.CreatedAt, version.CreatedBy,
	).Scan(&version.ID)
}

func (r *ConfigRepositoryImpl) GetConfigVersionByID(ctx context.Context, id string) (*domain.ConfigVersion, error) {
	query := `
		SELECT cv.id, cv.config_key_id, cv.value, cv.version, cv.created_at, cv.created_by, ck.value_type
		FROM config_versions cv
		JOIN config_keys ck ON ck.id = cv.config_key_id
		WHERE cv.id = $1`

	v := &domain.ConfigVersion{}
	var rawValue, valueType string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.ConfigKeyID, &rawValue, &v.Version, &v.CreatedAt, &v.CreatedBy, &valueType,
	)
	if err != nil {
		return nil, err
	}

	// Parse value from string to appropriate type
	v.Value, err = domain.ParseValue(rawValue, valueType)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (r *ConfigRepositoryImpl) GetConfigVersionsByKey(ctx context.Context, keyID string) ([]*domain.ConfigVersion, error) {
	// First get the value type from config_keys
	var valueType string
	err := r.db.QueryRowContext(ctx, `SELECT value_type FROM config_keys WHERE id = $1`, keyID).Scan(&valueType)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, config_key_id, value, version, created_at, created_by 
		FROM config_versions WHERE config_key_id = $1 ORDER BY version DESC`

	rows, err := r.db.QueryContext(ctx, query, keyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*domain.ConfigVersion
	for rows.Next() {
		v := &domain.ConfigVersion{}
		var rawValue string
		if err := rows.Scan(&v.ID, &v.ConfigKeyID, &rawValue, &v.Version, &v.CreatedAt, &v.CreatedBy); err != nil {
			return nil, err
		}

		// Parse value from string to appropriate type
		v.Value, err = domain.ParseValue(rawValue, valueType)
		if err != nil {
			return nil, err
		}

		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *ConfigRepositoryImpl) GetConfigVersionByNumber(ctx context.Context, keyID string, version int) (*domain.ConfigVersion, error) {
	// First get the value type from config_keys
	var valueType string
	err := r.db.QueryRowContext(ctx, `SELECT value_type FROM config_keys WHERE id = $1`, keyID).Scan(&valueType)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, config_key_id, value, version, created_at, created_by 
		FROM config_versions WHERE config_key_id = $1 AND version = $2`

	v := &domain.ConfigVersion{}
	var rawValue string
	err = r.db.QueryRowContext(ctx, query, keyID, version).Scan(
		&v.ID, &v.ConfigKeyID, &rawValue, &v.Version, &v.CreatedAt, &v.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Parse value from string to appropriate type
	v.Value, err = domain.ParseValue(rawValue, valueType)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (r *ConfigRepositoryImpl) GetMaxVersion(ctx context.Context, keyID string) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) FROM config_versions WHERE config_key_id = $1`

	var maxVersion int
	err := r.db.QueryRowContext(ctx, query, keyID).Scan(&maxVersion)
	return maxVersion, err
}
