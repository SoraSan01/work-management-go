package middleware

import (
	"github.com/gin-gonic/gin"
)

// InjectTemplateData ensures user data is available in all templates
func InjectTemplateData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user data from context
		role := c.GetString("role")
		userName := c.GetString("user_name")
		userEmail := c.GetString("user_email")
		userInitials := c.GetString("user_initials")
		permissions, _ := c.Get("Permissions")

		// Set template data
		c.Set("Role", role)
		c.Set("UserName", userName)
		c.Set("UserEmail", userEmail)
		c.Set("UserInitials", userInitials)
		c.Set("Permissions", permissions)

		c.Next()
	}
}
