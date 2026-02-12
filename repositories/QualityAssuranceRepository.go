package repositories

import "gorm.io/gorm"

type QualityAssuranceRepository struct {
	DB *gorm.DB
}

func NewQualityAssuranceRepository(db *gorm.DB) *QualityAssuranceRepository {
	return &QualityAssuranceRepository{DB: db}
}
