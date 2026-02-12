package repositories

import (
	"time"
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamRepository struct {
	DB *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{DB: db}
}

// GetAll returns all teams with members and their departments
func (r *TeamRepository) GetAll() ([]models.Team, error) {
	var teams []models.Team
	err := r.DB.
		Preload("Supervisor").
		Preload("Members").
		Preload("Members.Department").
		Order("created_at DESC").Find(&teams).Error
	return teams, err
}

// GetByID returns a single team by ID with members and department info
func (r *TeamRepository) GetByID(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.DB.
		Preload("Supervisor").
		Preload("Members").
		Preload("Members.Department").
		First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// Create adds a new team and its members
func (r *TeamRepository) Create(team *models.Team) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Create team
		if err := tx.Create(team).Error; err != nil {
			return err
		}

		// Add members
		if len(team.Members) > 0 {
			if err := tx.Model(team).Association("Members").Replace(team.Members); err != nil {
				return err
			}
		}

		return nil
	})
}

// Update modifies team info and members
func (r *TeamRepository) Update(team *models.Team) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Update team basic info
		team.UpdatedAt = time.Now()
		if err := tx.Model(&models.Team{}).Where("id = ?", team.ID).Updates(map[string]interface{}{
			"name":          team.Name,
			"supervisor_id": team.SupervisorID,
			"updated_at":    team.UpdatedAt,
		}).Error; err != nil {
			return err
		}

		// Replace members
		if err := tx.Model(team).Association("Members").Replace(team.Members); err != nil {
			return err
		}

		return nil
	})
}

// Delete removes a team and all member associations
func (r *TeamRepository) Delete(id uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Clear members
		if err := tx.Exec("DELETE FROM team_members WHERE team_id = ?", id).Error; err != nil {
			return err
		}
		// Delete team
		if err := tx.Delete(&models.Team{}, "id = ?", id).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetAllUsers returns all active users with department info
func (r *TeamRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := r.DB.Preload("Department").Where("is_active = ?", true).
		Order("first_name, last_name").Find(&users).Error
	return users, err
}

func (r *TeamRepository) GetSupervisors() ([]models.User, error) {
	var users []models.User

	err := r.DB.
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", "Supervisor").
		Where("users.is_active = ?", true).
		Order("users.first_name, users.last_name").
		Find(&users).Error

	return users, err
}

func (r *TeamRepository) GetEmployees() ([]models.User, error) {
	var users []models.User

	err := r.DB.
		Preload("Department").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", "Employee").
		Where("users.is_active = ?", true).
		Order("users.first_name, users.last_name").
		Find(&users).Error

	return users, err
}

func (r *TeamRepository) GetByUserID(userID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team

	err := r.DB.
		Joins("JOIN team_members tm ON tm.team_id = teams.id").
		Where("tm.user_id = ?", userID).
		Preload("Members").
		Preload("Supervisor").
		Find(&teams).Error

	return teams, err
}
