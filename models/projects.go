package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectRequestID *uuid.UUID      `gorm:"type:uuid" json:"project_request_id"`
	ProjectRequest   *ProjectRequest `gorm:"foreignKey:ProjectRequestID" json:"project_request"`

	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`

	ManagerID *uuid.UUID `gorm:"type:uuid" json:"manager_id"`
	Manager   *User      `gorm:"foreignKey:ManagerID;references:ID" json:"manager"`

	ApproverID *uuid.UUID `gorm:"type:uuid" json:"approver_id"`
	Approver   *User      `gorm:"foreignKey:ApproverID;references:ID" json:"approver"`

	TeamID *uuid.UUID `gorm:"type:uuid" json:"team_id"`
	Team   Team       `gorm:"foreignKey:TeamID" json:"team"`

	// Add approval workflow fields
	ApprovalStatus string `gorm:"default:'pending_manager_assignment'" json:"approval_status"`
	FinalQAStatus  string `gorm:"default:'not_submitted'" json:"final_qa_status"` // not_submitted, pending_qa_review, approved, rejected
	FinalQAComment string `gorm:"type:text" json:"final_qa_comment"`

	FinalQASubmittedBy *uuid.UUID `gorm:"type:uuid" json:"final_qa_submitted_by"`
	FinalQASubmitter   *User      `gorm:"foreignKey:FinalQASubmittedBy;references:ID" json:"final_qa_submitter"`
	FinalQASubmittedAt *time.Time `json:"final_qa_submitted_at"`

	FinalQAReviewedBy *uuid.UUID `gorm:"type:uuid" json:"final_qa_reviewed_by"`
	FinalQAReviewer   *User      `gorm:"foreignKey:FinalQAReviewedBy;references:ID" json:"final_qa_reviewer"`
	FinalQAReviewedAt *time.Time `json:"final_qa_reviewed_at"`
	FinalQASent       bool       `gorm:"default:false" json:"final_qa_sent"`
	FinalQASentAt     *time.Time `json:"final_qa_sent_at"`
	FinalQASentBy     *uuid.UUID `gorm:"type:uuid" json:"final_qa_sent_by"`
	FinalQASender     *User      `gorm:"foreignKey:FinalQASentBy;references:ID" json:"final_qa_sender"`

	Status    string    `gorm:"default:'draft'" json:"status"` // draft, planning, in_progress, completed, cancelled
	StartDate time.Time `json:"start_date"`
	DueDate   time.Time `json:"due_date"`
	Tasks     []Task    `gorm:"foreignKey:ProjectID" json:"tasks"`

	// Track approvals
	Approvals []Approval `gorm:"foreignKey:ProjectID" json:"approvals"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
