package repositories

import (
	"errors"
	"time"
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepository struct {
	DB *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{DB: db}
}

// Get all tasks
func (r *TaskRepository) GetAll() ([]models.Task, error) {
	var tasks []models.Task
	err := r.DB.Preload("Project").Preload("Assignee").Find(&tasks).Error
	return tasks, err
}

// Get task by ID
func (r *TaskRepository) GetByID(id uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.DB.Preload("Project").Preload("Assignee").First(&task, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Create a new task
func (r *TaskRepository) Create(task *models.Task) error {
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	return r.DB.Create(task).Error
}

// Update a task
func (r *TaskRepository) Update(task *models.Task) error {
	task.UpdatedAt = time.Now()
	return r.DB.Save(task).Error
}

// Delete a task
func (r *TaskRepository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&models.Task{}, "id = ?", id).Error
}

func (r *TaskRepository) StartTask(taskID uuid.UUID) error {
	now := time.Now()
	result := r.DB.
		Model(&models.Task{}).
		Where("id = ? AND status = ?", taskID, "todo").
		Updates(map[string]interface{}{
			"status":     "in_progress",
			"started_at": now,
			"is_paused":  false,
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task can only be started from the to do state")
	}
	return nil
}

func (r *TaskRepository) GetByAssignee(userID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task

	err := r.DB.
		Preload("Project").
		Preload("Assignee").
		Where("assigned_to = ?", userID).
		Find(&tasks).Error

	return tasks, err
}

func (r *TaskRepository) MarkForReviewAfterUpload(taskID uuid.UUID) error {
	now := time.Now()
	return r.DB.
		Model(&models.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":     "for_review",
			"started_at": nil,
			"is_paused":  false,
			"updated_at": now,
		}).
		Error
}

func (r *TaskRepository) PauseTask(taskID uuid.UUID) error {
	now := time.Now()
	result := r.DB.
		Model(&models.Task{}).
		Where("id = ? AND status = ? AND started_at IS NOT NULL", taskID, "in_progress").
		Updates(map[string]interface{}{
			"started_at": nil,
			"is_paused":  true,
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task is not currently running")
	}
	return nil
}

func (r *TaskRepository) ResumeTask(taskID uuid.UUID) error {
	now := time.Now()
	result := r.DB.
		Model(&models.Task{}).
		Where("id = ? AND status = ? AND is_paused = ?", taskID, "in_progress", true).
		Updates(map[string]interface{}{
			"started_at": now,
			"is_paused":  false,
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task is not currently paused")
	}
	return nil
}

func (r *TaskRepository) AddElapsedDuration(taskID uuid.UUID, seconds int64) error {
	if seconds <= 0 {
		return nil
	}
	return r.DB.
		Model(&models.Task{}).
		Where("id = ?", taskID).
		Update("duration_seconds", gorm.Expr("duration_seconds + ?", seconds)).Error
}

func (r *TaskRepository) GetBySupervisorID(supervisorID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	err := r.DB.
		Joins("JOIN projects ON projects.id = tasks.project_id AND projects.deleted_at IS NULL").
		Joins("JOIN teams ON teams.id = projects.team_id").
		Where("teams.supervisor_id = ?", supervisorID).
		Preload("Project").
		Preload("Assignee").
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) GetByQATeamMemberID(qaID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	err := r.DB.
		Joins("JOIN projects ON projects.id = tasks.project_id AND projects.deleted_at IS NULL").
		Joins("JOIN team_members ON team_members.team_id = projects.team_id").
		Where("team_members.user_id = ?", qaID).
		Preload("Project").
		Preload("Assignee").
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) CanQAReviewTask(taskID uuid.UUID, qaID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.
		Table("tasks").
		Joins("JOIN projects ON projects.id = tasks.project_id AND projects.deleted_at IS NULL").
		Joins("JOIN team_members ON team_members.team_id = projects.team_id").
		Where("tasks.id = ? AND team_members.user_id = ?", taskID, qaID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TaskRepository) GetByManagerID(managerID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	err := r.DB.
		Joins("JOIN projects ON projects.id = tasks.project_id AND projects.deleted_at IS NULL").
		Where("projects.manager_id = ?", managerID).
		Preload("Project").
		Preload("Assignee").
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) CanManagerReviewTask(taskID uuid.UUID, managerID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.
		Table("tasks").
		Joins("JOIN projects ON projects.id = tasks.project_id AND projects.deleted_at IS NULL").
		Where("tasks.id = ? AND projects.manager_id = ?", taskID, managerID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
