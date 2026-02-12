package models

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID" json:"user"`
	CompanyName   string    `json:"company_name"`
	ContactNumber string    `json:"contact_number"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
