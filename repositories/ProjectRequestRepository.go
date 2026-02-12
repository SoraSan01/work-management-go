package repositories

import (
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRequestRepository struct {
	DB *gorm.DB
}

func NewProjectRequestRepository(db *gorm.DB) *ProjectRequestRepository {
	return &ProjectRequestRepository{DB: db}
}

func (repo *ProjectRequestRepository) GetAll() ([]models.ProjectRequest, error) {
	var projects []models.ProjectRequest
	err := repo.DB.Find(&projects).Error
	return projects, err
}

func (repo *ProjectRequestRepository) Create(project *models.ProjectRequest) error {
	return repo.DB.Create(project).Error
}

func (repo *ProjectRequestRepository) GetAllByCustomerID(customerID uuid.UUID) ([]models.ProjectRequest, error) {
	var projects []models.ProjectRequest
	err := repo.DB.Where("customer_id = ?", customerID).Find(&projects).Error
	return projects, err
}
