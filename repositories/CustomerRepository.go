package repositories

import (
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	DB *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{DB: db}
}

func (r *CustomerRepository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Role").Preload("Permissions").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *CustomerRepository) FindByUserID(userID string) (*models.Customer, error) {
	var customer models.Customer
	err := r.DB.Where("user_id = ?", userID).First(&customer).Error
	return &customer, err
}
