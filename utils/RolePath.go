package utils

import "github.com/gin-gonic/gin"

func RolePath(c *gin.Context, path string) string {
	role := c.GetString("role")
	if role == "" {
		role = "admin" // default fallback
	}
	return "/" + role + path
}
