package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/sachin/config-manager/internal/domain"
)

type EnvironmentRepositoryImpl struct {
	db *sql.DB
}

func NewEnvironmentRepository(db *sql.DB) *EnvironmentRepositoryImpl {
	return &EnvironmentRepositoryImpl{db: db}
}

func (r *EnvironmentRepositoryImpl) Create(ctx context.Context, env *domain.Environment) error {
	query := `
		INSERT INTO environments (service_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	now := time.Now()
	env.CreatedAt = now
	env.UpdatedAt = now

	return r.db.QueryRowContext(ctx, query, env.ServiceID, env.Name, env.CreatedAt, env.UpdatedAt).Scan(&env.ID)
}

func (r *EnvironmentRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.Environment, error) {
	query := `SELECT id, service_id, name, created_at, updated_at FROM environments WHERE id = $1`

	env := &domain.Environment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&env.ID, &env.ServiceID, &env.Name, &env.CreatedAt, &env.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (r *EnvironmentRepositoryImpl) GetByServiceAndName(ctx context.Context, serviceID, name string) (*domain.Environment, error) {
	query := `SELECT id, service_id, name, created_at, updated_at FROM environments WHERE service_id = $1 AND name = $2`

	env := &domain.Environment{}
	err := r.db.QueryRowContext(ctx, query, serviceID, name).Scan(
		&env.ID, &env.ServiceID, &env.Name, &env.CreatedAt, &env.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (r *EnvironmentRepositoryImpl) ListByService(ctx context.Context, serviceID string) ([]*domain.Environment, error) {
	query := `SELECT id, service_id, name, created_at, updated_at FROM environments WHERE service_id = $1 ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []*domain.Environment
	for rows.Next() {
		env := &domain.Environment{}
		if err := rows.Scan(&env.ID, &env.ServiceID, &env.Name, &env.CreatedAt, &env.UpdatedAt); err != nil {
			return nil, err
		}
		envs = append(envs, env)
	}
	return envs, rows.Err()
}

func (r *EnvironmentRepositoryImpl) Update(ctx context.Context, env *domain.Environment) error {
	query := `
		UPDATE environments 
		SET name = $1, updated_at = $2
		WHERE id = $3`

	env.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, env.Name, env.UpdatedAt, env.ID)
	return err
}

func (r *EnvironmentRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM environments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
