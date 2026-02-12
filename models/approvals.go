package models

import (
	"time"

	"github.com/google/uuid"
)

type Approval struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid" json:"project_id"`
	Project   Project   `gorm:"foreignKey:ProjectID" json:"project"`

	ApproverID uuid.UUID `gorm:"type:uuid" json:"approver_id"`
	Approver   User      `gorm:"foreignKey:ApproverID" json:"approver"`

	Stage   string `gorm:"not null" json:"stage"`  // initial, milestone_1, final, etc.
	Status  string `gorm:"not null" json:"status"` // pending, approved, rejected
	Comment string `json:"comment"`

	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ApprovedAt time.Time `json:"approved_at"`
	RejectedAt time.Time `json:"rejected_at"`
}
