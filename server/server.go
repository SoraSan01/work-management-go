package server

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"work-management-system/routes"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// NewServer sets up the Gin engine with templates, static files, sessions, and routes
func NewServer() *gin.Engine {
	r := gin.Default()

	// --------------------
	// Static files
	// --------------------
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")

	// --------------------
	// Sessions
	// --------------------
	store := cookie.NewStore([]byte("Sora@09089831215"))
	r.Use(sessions.Sessions("workms_session", store))

	// --------------------
	// FuncMap
	// --------------------
	funcMap := template.FuncMap{
		"hasPerm": HasPerm,
	}

	// --------------------
	// Load templates recursively
	// --------------------
	tmpl := template.New("").Funcs(funcMap)

	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".html") {
			// Get template name relative to templates folder
			name := strings.TrimPrefix(path, "templates/")
			name = strings.ReplaceAll(name, "\\", "/") // Windows support
			_, err := tmpl.New(name).ParseFiles(path)
			if err != nil {
				return fmt.Errorf("error parsing template %s: %w", path, err)
			}
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to load templates: %v", err))
	}

	r.SetHTMLTemplate(tmpl)

	// --------------------
	// Routes
	// --------------------
	routes.SetupRoutes(r)

	return r
}

// HasPerm checks if a permission exists in a list
func HasPerm(perms []string, perm string) bool {
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
