package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/hookie/relay/internal/auth"
	"github.com/hookie/relay/internal/grpc"
	"github.com/hookie/relay/internal/redis"
	"github.com/hookie/relay/internal/supabase"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file if it exists (ignore errors)
	_ = godotenv.Load()
}

func main() {
	// Get configuration from environment
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":50051"
	}

	clerkSecretKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecretKey == "" {
		log.Fatal("CLERK_SECRET_KEY environment variable is required")
	}

	// Initialize Redis subscriber
	subscriber, err := redis.NewSubscriber(redisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize redis subscriber: %v", err)
	}
	defer subscriber.Close()

	// Initialize Clerk verifier
	verifier, err := auth.NewVerifier(clerkSecretKey)
	if err != nil {
		log.Fatalf("Failed to initialize clerk verifier: %v", err)
	}

	// Initialize Supabase client
	supabaseClient, err := supabase.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize supabase client: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(subscriber, verifier, supabaseClient)

	// Start listening
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}

	// Start server in goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	log.Printf("Relay service started on %s", grpcAddr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down relay service...")
	grpcServer.GracefulStop()
	log.Println("Relay service stopped")
}


