package controllers

import (
	"net/http"
	"strings"
	"time"
	"work-management-system/models"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventController struct {
	EventRepo        *repositories.EventRepository
	NotificationRepo *repositories.NotificationRepository
}

func NewEventController(repo *repositories.EventRepository, notificationRepo *repositories.NotificationRepository) *EventController {
	return &EventController{
		EventRepo:        repo,
		NotificationRepo: notificationRepo,
	}
}

/* ---------- PAGES ---------- */

func (ec *EventController) Calendar(c *gin.Context) {
	role := c.GetString("role")

	var templatePath string

	switch role { // tagged switch on the value of role
	case "admin":
		templatePath = "admin/events/calendar.html"
	case "supervisor":
		templatePath = "supervisor/events/calendar.html"
	case "manager":
		templatePath = "manager/events/calendar.html"
	case "employee":
		templatePath = "employee/events/calendar.html"
	case "qualityassurance":
		templatePath = "qa/events/calendar.html"
	default:
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.HTML(http.StatusOK, templatePath, utils.TemplateContext(c, gin.H{
		"ActivePage": "events",
	}))
}

/* ---------- ACTIONS ---------- */

func (c *EventController) Store(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		respondEventError(ctx, http.StatusUnauthorized, "Invalid user")
		return
	}

	title, start, end, description, ok := parseEventForm(ctx)
	if !ok {
		return
	}

	event := models.CalendarEvent{
		Title:       title,
		Description: description,
		StartTime:   start,
		EndTime:     end,
		UserID:      &userID,
	}

	if err := c.EventRepo.Create(&event); err != nil {
		respondEventError(ctx, http.StatusInternalServerError, "Failed to create event")
		return
	}

	_ = c.NotificationRepo.CreateForUserWithLink(userID, "Event created: "+event.Title, roleBasePath(ctx.GetString("role"))+"/events")
	if wantsJSON(ctx) {
		ctx.JSON(http.StatusCreated, gin.H{
			"message": "Event created successfully",
			"event":   eventResponse(event),
		})
		return
	}
	ctx.Redirect(http.StatusSeeOther, roleBasePath(ctx.GetString("role"))+"/events")
}

func (c *EventController) Update(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		respondEventError(ctx, http.StatusUnauthorized, "Invalid user")
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		respondEventError(ctx, http.StatusBadRequest, "Invalid event ID")
		return
	}
	event, err := c.EventRepo.GetByIDForUser(id, userID)
	if err != nil {
		respondEventError(ctx, http.StatusNotFound, "Event not found")
		return
	}

	title, start, end, description, ok := parseEventForm(ctx)
	if !ok {
		return
	}
	event.Title = title
	event.Description = description
	event.StartTime = start
	event.EndTime = end

	if err := c.EventRepo.Update(event); err != nil {
		respondEventError(ctx, http.StatusInternalServerError, "Failed to update event")
		return
	}

	_ = c.NotificationRepo.CreateForUserWithLink(userID, "Event updated: "+event.Title, roleBasePath(ctx.GetString("role"))+"/events")
	if wantsJSON(ctx) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Event updated successfully",
			"event":   eventResponse(*event),
		})
		return
	}
	ctx.Redirect(http.StatusSeeOther, roleBasePath(ctx.GetString("role"))+"/events")
}

func (c *EventController) Delete(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		respondEventError(ctx, http.StatusUnauthorized, "Invalid user")
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		respondEventError(ctx, http.StatusBadRequest, "Invalid event ID")
		return
	}
	event, err := c.EventRepo.GetByIDForUser(id, userID)
	if err != nil {
		respondEventError(ctx, http.StatusNotFound, "Event not found")
		return
	}

	if err := c.EventRepo.Delete(id); err != nil {
		respondEventError(ctx, http.StatusInternalServerError, "Failed to delete event")
		return
	}

	_ = c.NotificationRepo.CreateForUserWithLink(userID, "Event deleted: "+event.Title, roleBasePath(ctx.GetString("role"))+"/events")
	if wantsJSON(ctx) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Event deleted successfully",
			"id":      id,
		})
		return
	}
	ctx.Redirect(http.StatusSeeOther, roleBasePath(ctx.GetString("role"))+"/events")
}

/* ---------- JSON ---------- */

func (c *EventController) GetEventsJSON(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	events, err := c.EventRepo.GetAllEventsByUser(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load events"})
		return
	}

	var response []map[string]interface{}
	for _, e := range events {
		response = append(response, eventResponse(e))
	}

	ctx.JSON(http.StatusOK, response)
}

func roleBasePath(role string) string {
	switch role {
	case "admin":
		return "/admin"
	case "supervisor":
		return "/supervisor"
	case "manager":
		return "/manager"
	case "employee":
		return "/employee"
	case "qualityassurance":
		return "/qa"
	case "customer":
		return "/customer"
	default:
		return ""
	}
}

func parseEventForm(ctx *gin.Context) (string, time.Time, time.Time, string, bool) {
	title := strings.TrimSpace(ctx.PostForm("title"))
	description := strings.TrimSpace(ctx.PostForm("description"))
	if title == "" {
		respondEventError(ctx, http.StatusBadRequest, "Title is required")
		return "", time.Time{}, time.Time{}, "", false
	}

	start, err := time.ParseInLocation("2006-01-02T15:04", strings.TrimSpace(ctx.PostForm("start_time")), time.Local)
	if err != nil {
		respondEventError(ctx, http.StatusBadRequest, "Invalid start time")
		return "", time.Time{}, time.Time{}, "", false
	}
	end, err := time.ParseInLocation("2006-01-02T15:04", strings.TrimSpace(ctx.PostForm("end_time")), time.Local)
	if err != nil {
		respondEventError(ctx, http.StatusBadRequest, "Invalid end time")
		return "", time.Time{}, time.Time{}, "", false
	}
	if !end.After(start) {
		respondEventError(ctx, http.StatusBadRequest, "End time must be after start time")
		return "", time.Time{}, time.Time{}, "", false
	}

	return title, start, end, description, true
}

func wantsJSON(ctx *gin.Context) bool {
	return strings.EqualFold(strings.TrimSpace(ctx.GetHeader("X-Requested-With")), "XMLHttpRequest") ||
		strings.Contains(strings.ToLower(ctx.GetHeader("Accept")), "application/json")
}

func respondEventError(ctx *gin.Context, status int, message string) {
	if wantsJSON(ctx) {
		ctx.JSON(status, gin.H{"error": message})
		return
	}
	ctx.String(status, message)
}

func eventResponse(event models.CalendarEvent) map[string]interface{} {
	return map[string]interface{}{
		"id":          event.ID,
		"title":       event.Title,
		"start":       event.StartTime.Format(time.RFC3339),
		"end":         event.EndTime.Format(time.RFC3339),
		"description": event.Description,
	}
}
