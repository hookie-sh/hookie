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
	"github.com/hookie/ingest/internal/ratelimit"
	"github.com/hookie/ingest/internal/redis"
	"github.com/hookie/ingest/internal/supabase"
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

	supabaseClient, err := supabase.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize supabase client: %v", err)
	}

	// Create rate limiter and resolver
	limiter := ratelimit.NewLimiter(redisClient.GetRedisClient())
	resolver := ratelimit.NewResolver(supabaseClient, redisClient.GetRedisClient())

	// Create handlers
	topicsHandler := handlers.NewTopicsHandler(redisClient)
	anonHandler := handlers.NewAnonHandler(redisClient, supabaseClient)

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Authenticated topics endpoint
	mux.Handle("/topics/{topicID}",
		middleware.ForTopics(resolver, limiter)(
			http.HandlerFunc(topicsHandler.HandleTopicWebhook),
		),
	)

	// Anonymous topics endpoint
	mux.Handle("/anon/{anonTopicID}",
		middleware.ForAnon(redisClient, limiter, supabaseClient)(
			http.HandlerFunc(anonHandler.HandleAnonWebhook),
		),
	)

	handler := middleware.Logger(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:   120 * time.Second,
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

