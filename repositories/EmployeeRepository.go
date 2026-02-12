package repositories

import (
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeRepository struct {
	DB *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{DB: db}
}

func (r *EmployeeRepository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Role").Preload("Permissions").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
