package main

import (
	"context"
	"finalproject/config"
	"finalproject/database"
	"finalproject/router"
	"finalproject/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	cfg := config.Load()
	redisClient, redisErr := services.NewRedisClient(cfg)
	if redisClient != nil {
		restoreRedisStore := services.SetRedisStore(redisClient)
		defer restoreRedisStore()
		defer func() {
			if err := redisClient.Close(); err != nil {
				log.Printf("failed to close Redis client: %v", err)
			}
		}()
	}
	if redisErr != nil {
		log.Printf("Redis unavailable at startup; cache and rate limiting will fail open: %v", redisErr)
	}

	r := router.StartApp()
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	serverErrors := make(chan error, 1)

	fmt.Printf("Server starting on port %s\n", cfg.Port)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case signal := <-shutdownSignals:
		log.Printf("received %s; shutting down", signal)
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run server: %v", err)
		}
		return
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("graceful HTTP shutdown failed: %v", err)
	}
}

func healthCheck() {
	// Simple health check - can be enhanced to check DB connectivity
	fmt.Println("Health check passed")
	os.Exit(0)
}
