package models

import (
	"time"

	"github.com/google/uuid"
)

type CalendarEvent struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`

	UserID *uuid.UUID `json:"user_id"` // optional
	User   *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`

	ProjectID *uuid.UUID `json:"project_id"` // optional
	Project   *Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	TaskID *uuid.UUID `json:"task_id"`                                 // optional
	Task   *Task      `gorm:"foreignKey:TaskID" json:"task,omitempty"` // add this line

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
