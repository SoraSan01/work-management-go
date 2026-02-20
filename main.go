// package main

// import (
// 	"log"

// 	"work-management-system/config"
// 	"work-management-system/server"

// 	"github.com/joho/godotenv"
// )

// func main() {
// 	// 1. Load environment variables
// 	if err := godotenv.Load(); err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// 	// 2. Connect to PostgreSQL
// 	config.Connect()
// 	config.Migrate() // optional but recommended

// 	// 3. Create and start Gin server
// 	r := server.NewServer()
// 	if err := r.Run(":8080"); err != nil {
// 		log.Fatal(err)
// 	}
// }

package main

import (
	"log"
	"net/http"
	"os"

	"work-management-system/config"
	"work-management-system/server"

	"github.com/joho/godotenv"
)

func main() {
	// In cloud, .env may not exist. Don't crash if missing.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found (ok in production):", err)
	}

	// OPTIONAL: don't do migrations on every boot in production unless you really want that.
	config.Connect()
	// config.Migrate()

	r := server.NewServer()

	// Add a fast health endpoint BEFORE heavy middleware if possible
	r.GET("/kaithhealth", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/kaithhealthcheck", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/kaithheathcheck", func(c *gin.Context) { c.String(http.StatusOK, "ok") }) // matches your logs

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Bind to 0.0.0.0 so Leapcell proxy can reach it
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal(err)
	}
}
