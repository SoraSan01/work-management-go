package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskReviewComment struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	TaskID uuid.UUID `gorm:"type:uuid;index;not null" json:"task_id"`
	Task   Task      `gorm:"foreignKey:TaskID;references:ID" json:"task"`

	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	Creator   User      `gorm:"foreignKey:CreatedBy;references:ID" json:"creator"`

	Message string `gorm:"type:text;not null" json:"message"`
	Action  string `gorm:"type:varchar(32);default:'commented'" json:"action"` // commented, approved, rejected

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
