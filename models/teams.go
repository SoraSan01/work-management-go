package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	Name string

	SupervisorID uuid.UUID
	Supervisor   User `gorm:"foreignKey:SupervisorID"`

	Members []User `gorm:"many2many:team_members"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
