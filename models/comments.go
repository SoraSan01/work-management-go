package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID  uuid.UUID `gorm:"type:uuid" json:"project_id"`
	UserID     uuid.UUID `gorm:"type:uuid" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user"`
	Message    string    `gorm:"not null" json:"message"`
	IsInternal bool      `gorm:"default:false" json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
}
