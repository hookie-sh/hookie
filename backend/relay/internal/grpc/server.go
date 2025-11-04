package grpc

import (
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
}

func NewServer(subscriber *redis.Subscriber, verifier *auth.Verifier, supabaseClient *supabase.Client) *Server {
	grpcServer := grpc.NewServer()
	
	service := NewService(subscriber, verifier, supabaseClient)
	proto.RegisterRelayServiceServer(grpcServer, service)

	return &Server{
		grpcServer: grpcServer,
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


