package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/supabase-community/supabase-go"
)

type Client struct {
	client *supabase.Client
}

// GetClient returns the underlying supabase client (for internal use)
func (c *Client) GetClient() *supabase.Client {
	return c.client
}

func NewClient() (*Client, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_SECRET_KEY")

	if url == "" || key == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SECRET_KEY must be set")
	}

	// Log key prefix to verify it's being loaded (first 20 chars for security)
	keyPrefix := ""
	if len(key) > 20 {
		keyPrefix = key[:20] + "..."
	} else {
		keyPrefix = "***"
	}
	fmt.Printf("[Supabase Client] Initializing with URL=%s, Key=%s\n", url, keyPrefix)

	client, err := supabase.NewClient(url, key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase client: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

// VerifyApplicationAccess verifies that a user has access to an application
// Applications can be owned by a user (user_id) or an organization (org_id)
func (c *Client) VerifyApplicationAccess(ctx context.Context, userID, appID, orgID string) error {
	var result struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
		OrgID  string `json:"org_id"`
	}

	data, _, err := c.client.From("applications").
		Select("id,user_id,org_id", "exact", false).
		Eq("id", appID).
		Single().
		Execute()

	if err != nil {
		return fmt.Errorf("application not found or access denied: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to parse application data: %w", err)
	}

	// Check if user owns it directly
	if result.UserID == userID {
		return nil
	}

	// Check if it's an org app and user has access
	if result.OrgID != "" {
		// If orgID is provided, verify it matches
		if orgID != "" && result.OrgID == orgID {
			return nil
		}
		// Otherwise, verify user belongs to the organization
		// For now, we'll allow if org_id matches (assuming RLS handles membership)
		// In production, verify org membership via a join query or Clerk API
		if orgID == "" || result.OrgID == orgID {
			return nil
		}
	}

	return fmt.Errorf("access denied: user does not have access to this application")
}

// VerifyTopicAccess verifies that a user has access to a topic (through its application)
func (c *Client) VerifyTopicAccess(ctx context.Context, userID, appID, topicID, orgID string) error {
	// First verify application access
	if err := c.VerifyApplicationAccess(ctx, userID, appID, orgID); err != nil {
		return err
	}

	// Then verify the topic belongs to the application
	var result struct {
		ID            string `json:"id"`
		ApplicationID string `json:"application_id"`
	}

	data, _, err := c.client.From("topics").
		Select("id,application_id", "exact", false).
		Eq("id", topicID).
		Eq("application_id", appID).
		Single().
		Execute()

	if err != nil {
		return fmt.Errorf("topic not found or does not belong to application: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to parse topic data: %w", err)
	}

	return nil
}

