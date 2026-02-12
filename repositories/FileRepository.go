package repositories

import (
	"time"
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileRepository struct {
	DB *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{DB: db}
}

func (r *FileRepository) Create(file *models.File) error {
	file.ID = uuid.New()
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()
	return r.DB.Create(file).Error
}

func (r *FileRepository) GetByTask(taskID uuid.UUID) ([]models.File, error) {
	var files []models.File
	err := r.DB.Where("task_id = ?", taskID).Find(&files).Error
	return files, err
}
