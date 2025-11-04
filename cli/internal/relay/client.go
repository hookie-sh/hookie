package relay

import (
	"context"
	"fmt"
	"os"

	"github.com/hookie/cli/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn   *grpc.ClientConn
	client proto.RelayServiceClient
	token  string
}

func NewClient(relayURL, token string) (*Client, error) {
	if relayURL == "" {
		relayURL = os.Getenv("HOOKIE_RELAY_URL")
		if relayURL == "" {
			relayURL = "localhost:50051"
		}
	}

	conn, err := grpc.NewClient(relayURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to relay: %w", err)
	}

	return &Client{
		conn:   conn,
		client: proto.NewRelayServiceClient(conn),
		token:  token,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) createContext(ctx context.Context) context.Context {
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

func (c *Client) Subscribe(ctx context.Context, appID, topicID, orgID string) (proto.RelayService_SubscribeClient, error) {
	req := &proto.SubscribeRequest{
		AppId:   appID,
		TopicId: topicID,
		OrgId:   orgID,
	}
	return c.client.Subscribe(c.createContext(ctx), req)
}

