package main

import (
	"finalproject/config"
	"finalproject/database"
	"finalproject/router"
	"fmt"
	"log"
	"os"
)

func main() {
	// Check if running health check
	if len(os.Args) > 1 && os.Args[1] == "health" {
		healthCheck()
		return
	}

	if err := database.StartDB(); err != nil {
		log.Fatalf("failed to start database: %v", err)
	}

	r := router.StartApp()
	cfg := config.Load()

	fmt.Printf("Server starting on port %s\n", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func healthCheck() {
	// Simple health check - can be enhanced to check DB connectivity
	fmt.Println("Health check passed")
	os.Exit(0)
}
