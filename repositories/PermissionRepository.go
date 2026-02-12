package repositories

import (
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	DB *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{DB: db}
}

func (r *PermissionRepository) GetAll() ([]models.Permission, error) {
	var permissions []models.Permission
	if err := r.DB.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepository) Create(permission *models.Permission) error {
	return r.DB.Create(permission).Error
}

func (r *PermissionRepository) GetByID(id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	if err := r.DB.First(&permission, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&models.Permission{}, "id = ?", id).Error
}

func (r *PermissionRepository) Update(permission *models.Permission) error {
	return r.DB.Save(permission).Error
}
