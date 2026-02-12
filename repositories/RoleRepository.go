package repositories

import (
	"errors"
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository struct {
	DB *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{DB: db}
}

func (r *RoleRepository) All() ([]models.Role, error) {
	var roles []models.Role
	// Preload Permissions for each role
	if err := r.DB.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// Create a new role with permissions
func (r *RoleRepository) Create(role *models.Role, permissionIDs []string) error {
	var permissions []models.Permission
	if len(permissionIDs) > 0 {
		if err := r.DB.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
			return err
		}
	}

	role.Permissions = permissions

	if err := r.DB.Create(role).Error; err != nil {
		return err
	}
	return nil
}

// List all permissions for multi-select
func (r *RoleRepository) AllPermissions() ([]models.Permission, error) {
	var perms []models.Permission
	if err := r.DB.Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// Check if role exists by name
func (r *RoleRepository) ExistsByName(name string) (bool, error) {
	var role models.Role
	if err := r.DB.Where("name = ?", name).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *RoleRepository) Update(role *models.Role, permissionIDs []string) error {
	// Load existing role with permissions
	var existing models.Role
	if err := r.DB.Preload("Permissions").First(&existing, "id = ?", role.ID).Error; err != nil {
		return err
	}

	// Update role name
	existing.Name = role.Name

	// Update permissions
	var permissions []models.Permission
	if len(permissionIDs) > 0 {
		if err := r.DB.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
			return err
		}
	}
	// Use GORM association to replace permissions
	if err := r.DB.Model(&existing).Association("Permissions").Replace(permissions); err != nil {
		return err
	}

	// Save changes
	return r.DB.Save(&existing).Error
}

func (r *RoleRepository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&models.Role{}, "id = ?", id).Error
}
