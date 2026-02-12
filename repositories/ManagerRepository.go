package repositories

import "gorm.io/gorm"

type ManagerRepository struct {
	DB *gorm.DB
}

func NewManagerRepository(db *gorm.DB) *ManagerRepository {
	return &ManagerRepository{DB: db}
}
