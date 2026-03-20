package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	ProjectID uuid.UUID `gorm:"type:uuid" json:"project_id"`
	Project   Project   `gorm:"foreignKey:ProjectID;references:ID" json:"project"`

	TaskID uuid.UUID `gorm:"type:uuid" json:"task_id"`
	Task   Task      `gorm:"foreignKey:TaskID;references:ID" json:"task"`

	UploadedBy uuid.UUID `gorm:"type:uuid" json:"uploaded_by"`
	Uploader   User      `gorm:"foreignKey:UploadedBy;references:ID" json:"uploader"`

	FilePath string `gorm:"not null" json:"file_path"`
	FileType string `json:"file_type"`
	Version  int    `gorm:"default:1" json:"version"`
	IsLatest bool   `gorm:"default:true" json:"is_latest"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
