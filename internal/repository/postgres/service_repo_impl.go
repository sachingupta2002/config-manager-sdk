package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/sachin/config-manager/internal/domain"
)

type ServiceRepositoryImpl struct {
	db *sql.DB
}

func NewServiceRepository(db *sql.DB) *ServiceRepositoryImpl {
	return &ServiceRepositoryImpl{db: db}
}

func (r *ServiceRepositoryImpl) Create(ctx context.Context, service *domain.Service) error {
	query := `
		INSERT INTO services (name, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id`

	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now

	return r.db.QueryRowContext(ctx, query, service.Name, service.CreatedAt, service.UpdatedAt).Scan(&service.ID)
}

func (r *ServiceRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.Service, error) {
	query := `SELECT id, name, created_at, updated_at FROM services WHERE id = $1`

	service := &domain.Service{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&service.ID, &service.Name, &service.CreatedAt, &service.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (r *ServiceRepositoryImpl) GetByName(ctx context.Context, name string) (*domain.Service, error) {
	query := `SELECT id, name, created_at, updated_at FROM services WHERE name = $1`

	service := &domain.Service{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&service.ID, &service.Name, &service.CreatedAt, &service.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (r *ServiceRepositoryImpl) List(ctx context.Context) ([]*domain.Service, error) {
	query := `SELECT id, name, created_at, updated_at FROM services ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*domain.Service
	for rows.Next() {
		service := &domain.Service{}
		if err := rows.Scan(&service.ID, &service.Name, &service.CreatedAt, &service.UpdatedAt); err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, rows.Err()
}

func (r *ServiceRepositoryImpl) Update(ctx context.Context, service *domain.Service) error {
	query := `
		UPDATE services 
		SET name = $1, updated_at = $2
		WHERE id = $3`

	service.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, service.Name, service.UpdatedAt, service.ID)
	return err
}

func (r *ServiceRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM services WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
