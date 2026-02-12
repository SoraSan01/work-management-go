package middleware

import (
	"net/http"
	"os"
	"strings"
	"work-management-system/repositories"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthRequired checks JWT from cookie
func AuthRequired(customerRepo *repositories.CustomerRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil || accessToken == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Required fields
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		roleVal, ok := claims["role"].(string)
		if !ok || roleVal == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		role := strings.ToLower(roleVal)
		c.Set("user_id", sub)
		c.Set("role", role)

		// Set customer_id if role is customer
		if role == "customer" {
			customer, err := customerRepo.FindByUserID(sub)
			if err != nil {
				c.Redirect(http.StatusSeeOther, "/login")
				c.Abort()
				return
			}

			// ✅ REAL customer ID (FK-safe)
			c.Set("customer_id", customer.ID.String())
		}

		// Name
		name := ""
		if n, ok := claims["name"].(string); ok && n != "" {
			name = n
		} else {
			fn, _ := claims["first_name"].(string)
			ln, _ := claims["last_name"].(string)
			name = strings.TrimSpace(fn + " " + ln)
		}
		c.Set("user_name", name)

		// Email
		if email, ok := claims["email"].(string); ok {
			c.Set("user_email", email)
		}

		c.Set("user_initials", generateInitials(name))

		// Permissions
		perms := []string{}
		if p, ok := claims["permissions"].([]interface{}); ok {
			for _, v := range p {
				if s, ok := v.(string); ok {
					perms = append(perms, s)
				}
			}
		}
		c.Set("Permissions", perms)

		c.Next()
	}
}

// Helper function to generate initials from name
func generateInitials(name string) string {
	if name == "" {
		return "??"
	}

	// Remove extra spaces
	name = strings.TrimSpace(name)

	// Split by spaces
	parts := strings.Fields(name)

	if len(parts) == 0 {
		return "??"
	}

	if len(parts) == 1 {
		// Single word name, take first two letters if available
		if len(name) >= 2 {
			return strings.ToUpper(name[:2])
		}
		return strings.ToUpper(name[:1])
	}

	// Multiple words, take first letter of first and last word
	first := string(parts[0][0])
	last := string(parts[len(parts)-1][0])
	return strings.ToUpper(first + last)
}

// RequireRole allows a single role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := strings.ToLower(c.GetString("role"))
		if userRole != strings.ToLower(role) {
			c.Redirect(http.StatusSeeOther, "/dashboard")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRoles allows multiple roles
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := strings.ToLower(c.GetString("role"))
		for _, r := range roles {
			if userRole == strings.ToLower(r) {
				c.Next()
				return
			}
		}
		c.Redirect(http.StatusSeeOther, "/dashboard")
		c.Abort()
	}
}

// RequirePermission checks fine-grained permissions
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permsAny, exists := c.Get("Permissions")
		if !exists {
			c.Redirect(http.StatusSeeOther, "/dashboard")
			c.Abort()
			return
		}

		perms, ok := permsAny.([]string)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/dashboard")
			c.Abort()
			return
		}

		for _, p := range perms {
			if p == permission {
				c.Next()
				return
			}
		}

		c.Redirect(http.StatusSeeOther, "/dashboard")
		c.Abort()
	}
}

// Logger logs requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		println(c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}
