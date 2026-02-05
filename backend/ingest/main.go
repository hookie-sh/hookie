package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hookie/ingest/internal/handlers"
	"github.com/hookie/ingest/internal/middleware"
	"github.com/hookie/ingest/internal/redis"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file if it exists (ignore errors)
	_ = godotenv.Load()
}

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient, err := redis.NewClient(redisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize redis client: %v", err)
	}
	defer redisClient.Close()

	webhookHandler := handlers.NewWebhookHandler(redisClient)

	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	
	// Webhook endpoint - register all HTTP methods
	mux.HandleFunc("/webhooks/{topicId}", webhookHandler.HandleWebhook)

	handler := middleware.Logger(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

