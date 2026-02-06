package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/hookie/relay/internal/auth"
	"github.com/hookie/relay/internal/redis"
	"github.com/hookie/relay/internal/supabase"
	"github.com/hookie/relay/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	grpcServer *grpc.Server
	service    *Service
}

func NewServer(subscriber *redis.Subscriber, verifier *auth.Verifier, supabaseClient *supabase.Client) *Server {
	// Configure gRPC server options for backpressure handling
	maxConcurrentStreams := 1000
	if maxStreamsStr := os.Getenv("GRPC_MAX_CONCURRENT_STREAMS"); maxStreamsStr != "" {
		if maxStreams, err := strconv.Atoi(maxStreamsStr); err == nil && maxStreams > 0 {
			maxConcurrentStreams = maxStreams
		}
	}

	keepaliveParams := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Minute,
		MaxConnectionAge:      30 * time.Minute,
		MaxConnectionAgeGrace:  5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               1 * time.Second,
	}

	keepaliveEnforcement := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(uint32(maxConcurrentStreams)),
		grpc.KeepaliveParams(keepaliveParams),
		grpc.KeepaliveEnforcementPolicy(keepaliveEnforcement),
	}

	grpcServer := grpc.NewServer(opts...)
	
	service := NewService(subscriber, verifier, supabaseClient)
	proto.RegisterRelayServiceServer(grpcServer, service)

	log.Printf("gRPC server configured: MaxConcurrentStreams=%d", maxConcurrentStreams)

	return &Server{
		grpcServer: grpcServer,
		service:    service,
	}
}

func (s *Server) Serve(lis net.Listener) error {
	log.Printf("gRPC server starting on %s", lis.Addr().String())
	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

func (s *Server) Stop() {
	s.grpcServer.Stop()
}

// DisconnectAllClients marks all active clients as disconnected in the database
func (s *Server) DisconnectAllClients(ctx context.Context) {
	s.service.DisconnectAllClients(ctx)
}

// SetBroadcastListener sets the broadcast listener for the service
func (s *Server) SetBroadcastListener(listener interface {
	SubscribeToMachineID(ctx context.Context, machineID string) error
}) {
	s.service.SetBroadcastListener(listener)
}

// DisconnectClientByMachineID disconnects all active connections for a given database machine ID
func (s *Server) DisconnectClientByMachineID(dbMachineID string) {
	s.service.DisconnectClientByMachineID(dbMachineID)
}


