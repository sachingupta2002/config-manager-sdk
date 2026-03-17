package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/sachin/config-manager/internal/domain"
)

type AuditRepositoryImpl struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepositoryImpl {
	return &AuditRepositoryImpl{db: db}
}

func (r *AuditRepositoryImpl) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (service_id, config_key_id, action, old_value, new_value, performed_by, performed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	log.PerformedAt = time.Now()

	// Serialize values to string for database storage
	var oldValue, newValue string
	var err error

	if log.OldValue != nil {
		oldValue, err = domain.SerializeValue(log.OldValue)
		if err != nil {
			return err
		}
	}

	if log.NewValue != nil {
		newValue, err = domain.SerializeValue(log.NewValue)
		if err != nil {
			return err
		}
	}

	return r.db.QueryRowContext(ctx, query,
		log.ServiceID, log.ConfigKeyID, log.Action, oldValue, newValue, log.PerformedBy, log.PerformedAt,
	).Scan(&log.ID)
}

func (r *AuditRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	query := `
		SELECT id, service_id, config_key_id, action, old_value, new_value, performed_by, performed_at 
		FROM audit_logs WHERE id = $1`

	log := &domain.AuditLog{}
	var oldValue, newValue sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.ServiceID, &log.ConfigKeyID, &log.Action,
		&oldValue, &newValue, &log.PerformedBy, &log.PerformedAt,
	)
	if err != nil {
		return nil, err
	}

	if oldValue.Valid {
		log.OldValue = oldValue.String
	}
	if newValue.Valid {
		log.NewValue = newValue.String
	}

	return log, nil
}

func (r *AuditRepositoryImpl) ListByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, service_id, config_key_id, action, old_value, new_value, performed_by, performed_at 
		FROM audit_logs 
		WHERE service_id = $1 
		ORDER BY performed_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, serviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log := &domain.AuditLog{}
		var oldValue, newValue sql.NullString
		if err := rows.Scan(
			&log.ID, &log.ServiceID, &log.ConfigKeyID, &log.Action,
			&oldValue, &newValue, &log.PerformedBy, &log.PerformedAt,
		); err != nil {
			return nil, err
		}
		if oldValue.Valid {
			log.OldValue = oldValue.String
		}
		if newValue.Valid {
			log.NewValue = newValue.String
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *AuditRepositoryImpl) ListByConfigKey(ctx context.Context, configKeyID string, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, service_id, config_key_id, action, old_value, new_value, performed_by, performed_at 
		FROM audit_logs 
		WHERE config_key_id = $1 
		ORDER BY performed_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, configKeyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log := &domain.AuditLog{}
		var oldValue, newValue sql.NullString
		if err := rows.Scan(
			&log.ID, &log.ServiceID, &log.ConfigKeyID, &log.Action,
			&oldValue, &newValue, &log.PerformedBy, &log.PerformedAt,
		); err != nil {
			return nil, err
		}
		if oldValue.Valid {
			log.OldValue = oldValue.String
		}
		if newValue.Valid {
			log.NewValue = newValue.String
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *AuditRepositoryImpl) CountByService(ctx context.Context, serviceID string) (int, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE service_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, serviceID).Scan(&count)
	return count, err
}
