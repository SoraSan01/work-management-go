package repositories

import (
	"work-management-system/models"

	"gorm.io/gorm"
)

type ActivityLogRepository struct {
	DB *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) *ActivityLogRepository {
	return &ActivityLogRepository{DB: db}
}

func (r *ActivityLogRepository) Create(log *models.ActivityLog) error {
	return r.DB.Create(log).Error
}

func (r *ActivityLogRepository) ListRecent(limit int) ([]models.ActivityLog, error) {
	if limit <= 0 {
		limit = 200
	}

	var logs []models.ActivityLog
	err := r.DB.
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
