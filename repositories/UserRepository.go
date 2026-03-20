package repositories

import (
	"errors"
	"time"
	"work-management-system/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) All() ([]models.User, error) {
	var users []models.User
	if err := r.DB.Preload("Role").Preload("Department").Preload("Manager").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Create(user *models.User, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)

	if err := r.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var user models.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	if err := r.DB.
		Preload("Role").
		Preload("Department").
		Preload("Manager").
		First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User, password string) error {
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.PasswordHash = string(hash)
	}

	updates := map[string]interface{}{
		"email":         user.Email,
		"password_hash": user.PasswordHash,
		"role_id":       user.RoleID,
		"department_id": user.DepartmentID,
		"manager_id":    user.ManagerID,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"nationality":   user.Nationality,
		"is_active":     user.IsActive,
		"refresh_token": user.RefreshToken,
		"refresh_exp":   user.RefreshExp,
		"reset_token":   user.ResetToken,
		"reset_exp":     user.ResetExp,
		"updated_at":    time.Now(),
	}

	return r.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error
}

func (r *UserRepository) Delete(id string) error {
	return r.DB.Delete(&models.User{}, "id = ?", id).Error
}

// Fetch users by role
func (r *UserRepository) GetUsersByRole(roleName string) ([]models.User, error) {
	var users []models.User

	err := r.DB.
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", roleName).
		Order("users.first_name, users.last_name").
		Find(&users).Error

	return users, err
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	err := r.DB.Find(&users).Error
	return users, err
}
