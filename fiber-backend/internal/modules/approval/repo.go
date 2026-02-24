package approval

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, a *PendingApproval) error
	GetByID(ctx context.Context, id string) (*PendingApproval, error)
	UpdateStatus(ctx context.Context, id string, status ApprovalStatus, reviewerID string, notes *string) error
	ListPending(ctx context.Context) ([]PendingApproval, error)
}

type PgRepo struct {
	DB *pgxpool.Pool
}

func (r PgRepo) Create(ctx context.Context, a *PendingApproval) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO pending_approvals (requested_by, action, resource, resource_id, data, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		a.RequestedBy, a.Action, a.Resource, a.ResourceID, a.Data, a.ExpiresAt,
	)
	return err
}

func (r PgRepo) GetByID(ctx context.Context, id string) (*PendingApproval, error) {
	var a PendingApproval
	err := r.DB.QueryRow(ctx,
		`SELECT id, requested_by, action, resource, resource_id, data, status, reviewed_by, review_notes, expires_at, created_at, updated_at
		 FROM pending_approvals WHERE id = $1`, id,
	).Scan(&a.ID, &a.RequestedBy, &a.Action, &a.Resource, &a.ResourceID, &a.Data, &a.Status, &a.ReviewedBy, &a.ReviewNotes, &a.ExpiresAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r PgRepo) UpdateStatus(ctx context.Context, id string, status ApprovalStatus, reviewerID string, notes *string) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE pending_approvals SET status = $1, reviewed_by = $2, review_notes = $3, updated_at = now() WHERE id = $4`,
		status, reviewerID, notes, id,
	)
	return err
}

func (r PgRepo) ListPending(ctx context.Context) ([]PendingApproval, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, requested_by, action, resource, resource_id, data, status, reviewed_by, review_notes, expires_at, created_at, updated_at
		 FROM pending_approvals WHERE status = 'pending' ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PendingApproval
	for rows.Next() {
		var a PendingApproval
		err := rows.Scan(&a.ID, &a.RequestedBy, &a.Action, &a.Resource, &a.ResourceID, &a.Data, &a.Status, &a.ReviewedBy, &a.ReviewNotes, &a.ExpiresAt, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, nil
}
