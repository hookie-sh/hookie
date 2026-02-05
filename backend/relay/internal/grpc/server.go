package grpc

import (
	"context"
	"log"
	"net"

	"github.com/hookie/relay/internal/auth"
	"github.com/hookie/relay/internal/redis"
	"github.com/hookie/relay/internal/supabase"
	"github.com/hookie/relay/proto"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	service    *Service
}

func NewServer(subscriber *redis.Subscriber, verifier *auth.Verifier, supabaseClient *supabase.Client) *Server {
	grpcServer := grpc.NewServer()
	
	service := NewService(subscriber, verifier, supabaseClient)
	proto.RegisterRelayServiceServer(grpcServer, service)

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


