package models

import (
	"time"

	"github.com/google/uuid"
)

type ProjectFinalFile struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	ProjectID uuid.UUID `gorm:"type:uuid;index;not null" json:"project_id"`
	Project   Project   `gorm:"foreignKey:ProjectID;references:ID" json:"project"`

	UploadedBy uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	Uploader   User      `gorm:"foreignKey:UploadedBy;references:ID" json:"uploader"`

	FilePath string `gorm:"not null" json:"file_path"`
	FileName string `gorm:"not null" json:"file_name"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
