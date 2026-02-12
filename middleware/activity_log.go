package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"work-management-system/models"
	"work-management-system/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ActivityLogger(repo *repositories.ActivityLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			return
		}

		userID, err := uuid.Parse(c.GetString("user_id"))
		if err != nil {
			return
		}

		entityID := extractEntityID(c)
		entity := extractEntity(c.Request.URL.Path)
		action := strings.ToLower(c.Request.Method) + "_" + entity
		details := fmt.Sprintf("%s %s -> %d", c.Request.Method, c.FullPath(), c.Writer.Status())

		_ = repo.Create(&models.ActivityLog{
			UserID:   userID,
			Action:   action,
			Entity:   entity,
			EntityID: entityID,
			Details:  details,
		})
	}
}

func extractEntity(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	if len(parts) == 1 && parts[0] != "" {
		return parts[0]
	}
	return "unknown"
}

func extractEntityID(c *gin.Context) uuid.UUID {
	for _, p := range c.Params {
		if id, err := uuid.Parse(p.Value); err == nil {
			return id
		}
	}
	return uuid.Nil
}
