package repositories

import (
	"time"
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskReviewCommentRepository struct {
	DB *gorm.DB
}

func NewTaskReviewCommentRepository(db *gorm.DB) *TaskReviewCommentRepository {
	return &TaskReviewCommentRepository{DB: db}
}

func (r *TaskReviewCommentRepository) Create(comment *models.TaskReviewComment) error {
	comment.ID = uuid.New()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	return r.DB.Create(comment).Error
}

func (r *TaskReviewCommentRepository) GetByTask(taskID uuid.UUID) ([]models.TaskReviewComment, error) {
	var comments []models.TaskReviewComment
	err := r.DB.
		Preload("Creator").
		Where("task_id = ?", taskID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}
