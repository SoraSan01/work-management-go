package models

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID `gorm:"type:uuid" json:"project_id"`
	Project     Project   `gorm:"foreignKey:ProjectID" json:"project"`
	AssignedTo  uuid.UUID `gorm:"type:uuid" json:"assigned_to"`
	Assignee    User      `gorm:"foreignKey:AssignedTo" json:"assignee"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`

	Status          string     `gorm:"default:'todo'" json:"status"`  // todo, in_progress, for_review, done
	Priority        string     `gorm:"default:'low'" json:"priority"` // new field: low, medium, high
	StartedAt       *time.Time `json:"started_at"`
	DurationSeconds int64      `gorm:"default:0" json:"duration_seconds"`
	DueDate         time.Time  `json:"due_date"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
