package audit

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, limit, offset int) ([]AuditLog, error)
	GetByResource(ctx context.Context, resource string, resourceID string) ([]AuditLog, error)
}

type PgRepo struct {
	DB *pgxpool.Pool
}

func (r PgRepo) Create(ctx context.Context, log *AuditLog) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO audit_logs (user_id, action, resource, resource_id, old_values, new_values, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.UserID, log.Action, log.Resource, log.ResourceID, log.OldValues, log.NewValues, log.IPAddress, log.UserAgent,
	)
	return err
}

func (r PgRepo) List(ctx context.Context, limit, offset int) ([]AuditLog, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, user_id, action, resource, resource_id, old_values, new_values, ip_address, user_agent, created_at
		 FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &l.OldValues, &l.NewValues, &l.IPAddress, &l.UserAgent, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r PgRepo) GetByResource(ctx context.Context, resource string, resourceID string) ([]AuditLog, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, user_id, action, resource, resource_id, old_values, new_values, ip_address, user_agent, created_at
		 FROM audit_logs WHERE resource = $1 AND resource_id = $2 ORDER BY created_at DESC`,
		resource, resourceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &l.OldValues, &l.NewValues, &l.IPAddress, &l.UserAgent, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
