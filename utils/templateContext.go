package utils

import (
	"fmt"
	"time"
	"work-management-system/config"
	"work-management-system/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TemplateContext merges user info and extra template data for sidebar and pages
func TemplateContext(c *gin.Context, extra gin.H) gin.H {
	if extra == nil {
		extra = gin.H{}
	}

	// Get user info from context
	role := c.GetString("role")
	userName := c.GetString("user_name")
	userEmail := c.GetString("user_email")
	userInitials := c.GetString("user_initials")

	// Get user permissions
	userPermsAny, _ := c.Get("Permissions")
	userPerms, _ := userPermsAny.([]string)

	// Merge into result
	data := gin.H{
		"Role":         role,
		"UserName":     userName,
		"UserEmail":    userEmail,
		"UserInitials": userInitials,
		"Notifications":    []gin.H{},
		"NotificationCount": int64(0),
	}

	userIDStr := c.GetString("user_id")
	if userID, err := uuid.Parse(userIDStr); err == nil {
		var notifications []models.Notification
		_ = config.DB.
			Where("user_id = ?", userID).
			Order("created_at DESC").
			Limit(5).
			Find(&notifications).Error

		var unreadCount int64
		_ = config.DB.
			Model(&models.Notification{}).
			Where("user_id = ? AND is_read = ?", userID, false).
			Count(&unreadCount).Error

		items := make([]gin.H, 0, len(notifications))
		for _, n := range notifications {
			items = append(items, gin.H{
				"ID":      n.ID,
				"Message": n.Message,
				"Read":    n.IsRead,
				"Link":    fallbackLink(n.Link),
				"Icon":    "bell",
				"Color":   "blue",
				"Time":    formatNotificationTime(n.CreatedAt),
			})
		}

		data["Notifications"] = items
		data["NotificationCount"] = unreadCount
	}

	// Only set "Permissions" if not already provided by controller
	if _, ok := extra["Permissions"]; !ok {
		data["Permissions"] = userPerms
	}

	// Merge extra fields from controller
	for k, v := range extra {
		data[k] = v
	}

	return data
}

func fallbackLink(link string) string {
	if link == "" {
		return "#"
	}
	return link
}

func formatNotificationTime(t time.Time) string {
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return pluralizeTime(int(diff.Minutes()), "minute")
	case diff < 24*time.Hour:
		return pluralizeTime(int(diff.Hours()), "hour")
	case diff < 7*24*time.Hour:
		return pluralizeTime(int(diff.Hours()/24), "day")
	default:
		return t.Format("Jan 02, 2006")
	}
}

func pluralizeTime(value int, unit string) string {
	if value <= 1 {
		return "1 " + unit + " ago"
	}
	return fmt.Sprintf("%d %ss ago", value, unit)
}
