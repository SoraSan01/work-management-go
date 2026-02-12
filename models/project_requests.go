package models

import (
	"time"

	"github.com/google/uuid"
)

type ProjectRequest struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerID  uuid.UUID `gorm:"type:uuid" json:"customer_id"`
	Customer    Customer  `gorm:"foreignKey:CustomerID" json:"customer"`
	Title       string    `gorm:"not null" json:"title" form:"title"` // Added form tag
	Description string    `json:"description" form:"description"`     // Added form tag
	Deadline    time.Time `form:"deadline"`                           // Added form tag
	Status      string    `gorm:"default:'pending'" json:"status"`    // pending, approved, rejected
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
