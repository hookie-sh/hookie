package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hookie-sh/hookie/backend/ingest/internal/handlers"
	"github.com/hookie-sh/hookie/backend/ingest/internal/middleware"
	"github.com/hookie-sh/hookie/backend/ingest/internal/ratelimit"
	"github.com/hookie-sh/hookie/backend/ingest/internal/redis"
	"github.com/hookie-sh/hookie/backend/ingest/internal/supabase"
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

	// Configure timeouts - make them configurable for high load scenarios
	readTimeout := 10 * time.Second
	if readTimeoutStr := os.Getenv("HTTP_READ_TIMEOUT"); readTimeoutStr != "" {
		if duration, err := time.ParseDuration(readTimeoutStr); err == nil {
			readTimeout = duration
		}
	}

	writeTimeout := 10 * time.Second
	if writeTimeoutStr := os.Getenv("HTTP_WRITE_TIMEOUT"); writeTimeoutStr != "" {
		if duration, err := time.ParseDuration(writeTimeoutStr); err == nil {
			writeTimeout = duration
		}
	}

	idleTimeout := 120 * time.Second
	if idleTimeoutStr := os.Getenv("HTTP_IDLE_TIMEOUT"); idleTimeoutStr != "" {
		if duration, err := time.ParseDuration(idleTimeoutStr); err == nil {
			idleTimeout = duration
		}
	}

	// For high concurrency (100+ concurrent connections), increase timeouts
	// This prevents connection exhaustion under load
	maxHeaderBytes := 1 << 20 // 1MB
	if maxHeaderBytesStr := os.Getenv("HTTP_MAX_HEADER_BYTES"); maxHeaderBytesStr != "" {
		if mb, err := strconv.Atoi(maxHeaderBytesStr); err == nil && mb > 0 {
			maxHeaderBytes = mb << 20 // Convert MB to bytes
		}
	}

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: maxHeaderBytes,
		ReadTimeout:    readTimeout,
		WriteTimeout:  writeTimeout,
		IdleTimeout:    idleTimeout,
	}

	log.Printf("HTTP server configured: ReadTimeout=%v, WriteTimeout=%v, IdleTimeout=%v, MaxHeaderBytes=%d", 
		readTimeout, writeTimeout, idleTimeout, maxHeaderBytes)

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

