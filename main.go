package main

import (
	"log"

	"work-management-system/config"
	"work-management-system/server"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Connect to PostgreSQL
	config.Connect()
	config.Migrate() // optional but recommended

	// 3. Create and start Gin server
	r := server.NewServer()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
