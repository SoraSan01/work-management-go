package repositories

import (
	"work-management-system/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository struct {
	DB *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{DB: db}
}

func (r *EventRepository) GetAllEvents() ([]models.CalendarEvent, error) {
	var events []models.CalendarEvent
	return events, r.DB.Find(&events).Error
}

func (r *EventRepository) GetAllEventsByUser(userID uuid.UUID) ([]models.CalendarEvent, error) {
	var events []models.CalendarEvent
	err := r.DB.
		Where("user_id = ?", userID).
		Order("start_time ASC").
		Find(&events).Error
	return events, err
}

func (r *EventRepository) GetByID(id uuid.UUID) (*models.CalendarEvent, error) {
	var event models.CalendarEvent
	err := r.DB.First(&event, "id = ?", id).Error
	return &event, err
}

func (r *EventRepository) GetByIDForUser(id uuid.UUID, userID uuid.UUID) (*models.CalendarEvent, error) {
	var event models.CalendarEvent
	err := r.DB.Where("id = ? AND user_id = ?", id, userID).First(&event).Error
	return &event, err
}

func (r *EventRepository) Create(event *models.CalendarEvent) error {
	return r.DB.Create(event).Error
}

func (r *EventRepository) Update(event *models.CalendarEvent) error {
	return r.DB.Save(event).Error
}

func (r *EventRepository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&models.CalendarEvent{}, "id = ?", id).Error
}
