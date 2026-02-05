package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hookie/relay/internal/auth"
	"github.com/hookie/relay/internal/grpc"
	realtime "github.com/hookie/relay/internal/realtime"
	"github.com/hookie/relay/internal/redis"
	"github.com/hookie/relay/internal/supabase"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file if it exists (ignore errors)
	_ = godotenv.Load()
}

func main() {
	var grpcServer *grpc.Server
	
	// Handle panics to ensure cleanup
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered: %v", r)
			// Try to mark clients as disconnected even on panic
			// Note: This might not work if the panic happened before initialization
			if grpcServer != nil {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cleanupCancel()
				log.Println("Attempting to disconnect all clients after panic...")
				grpcServer.DisconnectAllClients(cleanupCtx)
			}
			panic(r) // Re-panic to maintain original behavior
		}
	}()

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
	grpcServer = grpc.NewServer(subscriber, verifier, supabaseClient)
	
	// Start broadcast listener for force disconnect
	broadcastListener, err := realtime.NewBroadcastListener(grpcServer)
	if err != nil {
		log.Fatalf("Failed to initialize broadcast listener: %v", err)
	}
	
	// Set broadcast listener in the service
	grpcServer.SetBroadcastListener(broadcastListener)
	
	// Start broadcast listener in background
	go func() {
		ctx := context.Background()
		if err := broadcastListener.Start(ctx); err != nil {
			log.Printf("Broadcast listener error: %v", err)
		}
	}()
	log.Printf("Broadcast listener started")

	// Start Postgres change listener for connected_clients table
	postgresListener, err := realtime.NewListener(grpcServer)
	if err != nil {
		log.Fatalf("Failed to initialize postgres listener: %v", err)
	}
	
	// Start postgres listener in background
	go func() {
		ctx := context.Background()
		if err := postgresListener.Start(ctx); err != nil {
			log.Printf("Postgres listener error: %v", err)
		}
	}()
	log.Printf("Postgres change listener started")

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
	
	// Create a context with timeout for cleanup operations
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()
	
	// Mark all active clients as disconnected before shutting down
	log.Println("Marking all active clients as disconnected...")
	grpcServer.DisconnectAllClients(cleanupCtx)
	
	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
	log.Println("Relay service stopped")
}


