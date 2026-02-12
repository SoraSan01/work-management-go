package repositories

import (
	"time"
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	DB *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{DB: db}
}

func (r *NotificationRepository) Create(notification *models.Notification) error {
	return r.DB.Create(notification).Error
}

func (r *NotificationRepository) CreateForUser(userID uuid.UUID, message string) error {
	return r.CreateForUserWithLink(userID, message, "")
}

func (r *NotificationRepository) CreateForUserWithLink(userID uuid.UUID, message, link string) error {
	notification := models.Notification{
		UserID:    userID,
		Message:   message,
		Link:      link,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	return r.DB.Create(&notification).Error
}

func (r *NotificationRepository) CreateForAllUsers(message string) error {
	return r.CreateForAllUsersWithLink(message, "")
}

func (r *NotificationRepository) CreateForAllUsersWithLink(message, link string) error {
	var users []models.User
	if err := r.DB.Select("id").Where("is_active = ?", true).Find(&users).Error; err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}

	now := time.Now()
	notifications := make([]models.Notification, 0, len(users))
	for _, user := range users {
		notifications = append(notifications, models.Notification{
			UserID:    user.ID,
			Message:   message,
			Link:      link,
			IsRead:    false,
			CreatedAt: now,
		})
	}

	return r.DB.Create(&notifications).Error
}

func (r *NotificationRepository) RecentByUser(userID uuid.UUID, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	q := r.DB.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) CountUnreadByUser(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepository) MarkRead(userID, notificationID uuid.UUID) error {
	return r.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllRead(userID uuid.UUID) error {
	return r.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
