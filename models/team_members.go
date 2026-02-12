package models

import "github.com/google/uuid"

type TeamMember struct {
	TeamID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
}
