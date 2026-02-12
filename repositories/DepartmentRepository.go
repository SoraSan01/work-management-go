package repositories

import (
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepartmentRepository struct {
	DB *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{DB: db}
}

func (r *DepartmentRepository) All() ([]models.Department, error) {
	var departments []models.Department
	if err := r.DB.Find(&departments).Error; err != nil {
		return nil, err
	}
	return departments, nil
}

// Create a new department
func (r *DepartmentRepository) Create(department *models.Department) error {
	return r.DB.Create(department).Error
}

// Delete a department by ID
func (r *DepartmentRepository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&models.Department{}, "id = ?", id).Error
}
