package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectRequestID *uuid.UUID      `gorm:"type:uuid" json:"project_request_id"`
	ProjectRequest   *ProjectRequest `gorm:"foreignKey:ProjectRequestID" json:"project_request"`

	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	Priority    string `gorm:"default:'low'" json:"priority"` // low, medium, high

	ManagerID *uuid.UUID `gorm:"type:uuid" json:"manager_id"`
	Manager   *User      `gorm:"foreignKey:ManagerID;references:ID" json:"manager"`

	AssignedEmployeeID *uuid.UUID `gorm:"type:uuid" json:"assigned_employee_id"`
	AssignedEmployee   *User      `gorm:"foreignKey:AssignedEmployeeID;references:ID" json:"assigned_employee"`

	ApproverID *uuid.UUID `gorm:"type:uuid" json:"approver_id"`
	Approver   *User      `gorm:"foreignKey:ApproverID;references:ID" json:"approver"`

	TeamID *uuid.UUID `gorm:"type:uuid" json:"team_id"`
	Team   Team       `gorm:"foreignKey:TeamID" json:"team"`

	// Add approval workflow fields
	ApprovalStatus string `gorm:"default:'pending_manager_assignment'" json:"approval_status"`

	Status    string    `gorm:"default:'draft'" json:"status"` // draft, planning, in_progress, completed, cancelled
	StartDate time.Time `json:"start_date"`
	DueDate   time.Time `json:"due_date"`
	Tasks     []Task    `gorm:"foreignKey:ProjectID" json:"tasks"`

	// Track approvals
	Approvals []Approval `gorm:"foreignKey:ProjectID" json:"approvals"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
