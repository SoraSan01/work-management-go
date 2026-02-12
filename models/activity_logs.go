package models

import (
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Action    string    `gorm:"not null" json:"action"`     // e.g., "created_task", "approved_project"
	Entity    string    `gorm:"not null" json:"entity"`     // e.g., "project", "task", "comment"
	EntityID  uuid.UUID `gorm:"type:uuid" json:"entity_id"` // the id of the entity
	Details   string    `json:"details"`                    // optional extra info
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
