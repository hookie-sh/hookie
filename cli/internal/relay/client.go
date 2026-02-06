package relay

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hookie/cli/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn      *grpc.ClientConn
	client    proto.RelayServiceClient
	token     string
	channelID string  // anonymous channel ID
	anonymous bool    // anonymous mode flag
}

func NewClient(token string, debug bool) (*Client, error) {
	relayURL := os.Getenv("HOOKIE_RELAY_URL")
	if relayURL == "" {
		relayURL = GetRelayURL()
	}
	// if relayURL == "" {
	// 	relayURL = os.Getenv("HOOKIE_RELAY_URL")
	// 	log.Println("relayURL from env", relayURL)
	// 	if relayURL == "" {
	// 		relayURL = "localhost:50051"
	// 	}
	// }

	// Determine transport credentials based on URL
	var creds credentials.TransportCredentials
	if isLocalhost(relayURL) && os.Getenv("HOOKIE_INSECURE_TLS") == "" {
		// Use insecure credentials for localhost (dev convenience)
		creds = insecure.NewCredentials()
	} else {
		// Use TLS for remote connections (production)
		creds = credentials.NewTLS(nil) // nil means use system root CAs
	}

	conn, err := grpc.NewClient(relayURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to relay: %w", err)
	}

	return &Client{
		conn:   conn,
		client: proto.NewRelayServiceClient(conn),
		token:  token,
	}, nil
}

// isLocalhost checks if the URL is pointing to localhost or 127.0.0.1
func isLocalhost(url string) bool {
	// Remove scheme if present
	host := strings.TrimPrefix(url, "grpc://")
	host = strings.TrimPrefix(host, "grpcs://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	
	// Extract host:port and check host
	host, _, err := net.SplitHostPort(host)
	if err != nil {
		// No port, use entire string as host
		host = url
	}
	
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || host == ""
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) createContext(ctx context.Context) context.Context {
	if c.anonymous {
		md := metadata.New(map[string]string{
			"x-channel-type": "anonymous",
		})
		return metadata.NewOutgoingContext(ctx, md)
	}
	md := metadata.New(map[string]string{
		"authorization": c.token,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *Client) ListApplications(ctx context.Context, orgID string) ([]*proto.Application, error) {
	req := &proto.ListApplicationsRequest{
		OrgId: orgID,
	}
	resp, err := c.client.ListApplications(c.createContext(ctx), req)
	if err != nil {
		return nil, err
	}
	return resp.Applications, nil
}

func (c *Client) ListTopics(ctx context.Context, appID string) ([]*proto.Topic, error) {
	req := &proto.ListTopicsRequest{
		AppId: appID,
	}
	resp, err := c.client.ListTopics(c.createContext(ctx), req)
	if err != nil {
		return nil, err
	}
	return resp.Topics, nil
}

func (c *Client) Subscribe(ctx context.Context, appID, topicID, orgID, machineID string) (proto.RelayService_SubscribeClient, error) {
	req := &proto.SubscribeRequest{
		AppId:     appID,
		TopicId:   topicID,
		OrgId:     orgID,
		MachineId: machineID,
	}
	return c.client.Subscribe(c.createContext(ctx), req)
}

// NewAnonymousClient creates a new relay client for anonymous usage (no auth)
func NewAnonymousClient(debug bool) (*Client, error) {
	relayURL := os.Getenv("HOOKIE_RELAY_URL")
	if relayURL == "" {
		relayURL = GetRelayURL()
	}

	// Determine transport credentials based on URL
	var creds credentials.TransportCredentials
	if isLocalhost(relayURL) && os.Getenv("HOOKIE_INSECURE_TLS") == "" {
		// Use insecure credentials for localhost (dev convenience)
		creds = insecure.NewCredentials()
	} else {
		// Use TLS for remote connections (production)
		creds = credentials.NewTLS(nil) // nil means use system root CAs
	}

	conn, err := grpc.NewClient(relayURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to relay: %w", err)
	}

	return &Client{
		conn:      conn,
		client:    proto.NewRelayServiceClient(conn),
		anonymous: true,
	}, nil
}

// CreateAnonymousChannel creates an anonymous ephemeral channel
func (c *Client) CreateAnonymousChannel(ctx context.Context) (*proto.CreateAnonymousChannelResponse, error) {
	return c.client.CreateAnonymousChannel(c.createContext(ctx), &proto.CreateAnonymousChannelRequest{})
}

// SetChannelID sets the anonymous channel ID
func (c *Client) SetChannelID(channelID string) {
	c.channelID = channelID
}

