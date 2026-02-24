package approval

import (
	"time"
)

type ApprovalStatus string

const (
	StatusPending  ApprovalStatus = "pending"
	StatusApproved ApprovalStatus = "approved"
	StatusRejected ApprovalStatus = "rejected"
)

type PendingApproval struct {
	ID          string         `json:"id"`
	RequestedBy string         `json:"requested_by"`
	Action      string         `json:"action"`
	Resource    string         `json:"resource"`
	ResourceID  *string        `json:"resource_id"`
	Data        any            `json:"data"`
	Status      ApprovalStatus `json:"status"`
	ReviewedBy  *string        `json:"reviewed_by"`
	ReviewNotes *string        `json:"review_notes"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
