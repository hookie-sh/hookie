package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

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

// GetTopicApplicationID looks up the application_id for a given topic
func (c *Client) GetTopicApplicationID(ctx context.Context, topicID string) (string, error) {
	var result struct {
		ApplicationID string `json:"application_id"`
	}

	data, _, err := c.client.From("topics").
		Select("application_id", "exact", false).
		Eq("id", topicID).
		Single().
		Execute()

	if err != nil {
		return "", fmt.Errorf("topic not found: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse topic data: %w", err)
	}

	return result.ApplicationID, nil
}

// GetApplicationOrgID looks up the org_id for a given application
// Returns empty string if the application is user-owned (not org-owned)
func (c *Client) GetApplicationOrgID(ctx context.Context, appID string) (string, error) {
	var result struct {
		OrgID *string `json:"org_id"` // Use pointer to detect NULL vs empty string
	}

	data, _, err := c.client.From("applications").
		Select("org_id", "exact", false).
		Eq("id", appID).
		Single().
		Execute()

	if err != nil {
		log.Printf("[GetApplicationOrgID] ERROR querying appID=%s: %v", appID, err)
		return "", fmt.Errorf("application not found: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("[GetApplicationOrgID] ERROR unmarshaling appID=%s: %v", appID, err)
		return "", fmt.Errorf("failed to parse application data: %w", err)
	}

	// Return empty string if org_id is NULL (user-owned app)
	if result.OrgID == nil {
		return "", nil
	}

	// Return the org_id value
	return *result.OrgID, nil
}

// ensureUserExists checks if a user exists, and if not, creates a minimal user record
func (c *Client) ensureUserExists(ctx context.Context, userID string) error {
	// Check if user exists
	data, _, err := c.client.From("users").
		Select("id", "exact", false).
		Eq("id", userID).
		Single().
		Execute()

	if err == nil && len(data) > 0 {
		// User exists, no need to create
		return nil
	}

	// User doesn't exist, create a minimal user record
	// Use a placeholder email since we don't have the actual email in the relay service
	// The webhook will update this with the real email later
	userData := map[string]interface{}{
		"id":    userID,
		"email": fmt.Sprintf("%s@placeholder.hookie", userID), // Placeholder email
	}

	_, _, err = c.client.From("users").
		Insert(userData, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("failed to ensure user exists: %w", err)
	}

	return nil
}

// UpsertConnectedClient creates or updates a connected client record for a machine+org context
// machineID is the mach_<ksuid> value from the CLI, which becomes the id (primary key)
// Returns the id (which is the machineID)
func (c *Client) UpsertConnectedClient(ctx context.Context, userID, machineID, orgID string) (string, error) {
	// Ensure user exists before inserting
	if err := c.ensureUserExists(ctx, userID); err != nil {
		// Log but don't fail - the insert might still work if user was created
		log.Printf("Warning: failed to ensure user exists: %v", err)
	}

	// Build query to check if record exists (regardless of disconnected_at status)
	// Note: For empty org_id, we use empty string (not NULL)
	// IMPORTANT: We must filter by ALL three primary key components: id, user_id, org_id
	query := c.client.From("connected_clients").
		Select("id,user_id,org_id", "exact", false).
		Eq("id", machineID).
		Eq("user_id", userID)
	
	// Handle empty string for org_id (personal accounts)
	if orgID == "" {
		query = query.Eq("org_id", "")
	} else {
		query = query.Eq("org_id", orgID)
	}

	// Query without Single() to avoid error when no record exists
	data, _, err := query.Execute()
	
	// Parse JSON to check if we actually got records (not just empty array)
	// This fixes the bug where empty JSON array "[]" (2 bytes) was incorrectly
	// treated as "record exists" when checking len(data) > 0
	var records []map[string]interface{}
	hasRecords := false
	if err == nil && len(data) > 0 {
		// Check if data is not just an empty JSON array "[]"
		dataStr := string(data)
		if dataStr != "[]" && dataStr != "null" {
			if err := json.Unmarshal(data, &records); err == nil {
				hasRecords = len(records) > 0
			}
		}
	}
	
	recordExists := hasRecords
	
	if recordExists {
		// Record exists, update connected_at and clear disconnected_at
		log.Printf("[UpsertConnectedClient] Record EXISTS, updating: id=%s user_id=%s org_id=%q", machineID, userID, orgID)
		updateQuery := c.client.From("connected_clients").
			Update(map[string]interface{}{
				"connected_at":   time.Now().Format(time.RFC3339),
				"disconnected_at": nil,
			}, "", "").
			Eq("id", machineID).
			Eq("user_id", userID).
			Eq("org_id", orgID) // orgID is empty string for personal accounts

		_, _, err := updateQuery.Execute()

		if err != nil {
			return "", fmt.Errorf("failed to update connected client: %w", err)
		}

		return machineID, nil
	}
	
	if err != nil {
		log.Printf("[UpsertConnectedClient] Query error (proceeding to insert): %v", err)
	}

	// Record doesn't exist, insert new one
	log.Printf("[UpsertConnectedClient] Record DOES NOT EXIST, creating new: id=%s user_id=%s org_id=%q", machineID, userID, orgID)
	clientData := map[string]interface{}{
		"id":      machineID, // The mach_<ksuid> becomes part of the composite primary key
		"user_id": userID,
		"org_id":  orgID, // Empty string for personal accounts, org ID for organizations
	}

	_, _, err = c.client.From("connected_clients").
		Insert(clientData, false, "", "", "").
		Execute()

	if err != nil {
		log.Printf("[UpsertConnectedClient] Insert ERROR: %v", err)
		return "", fmt.Errorf("failed to create connected client: %w", err)
	}

	log.Printf("[UpsertConnectedClient] Successfully created record: id=%s user_id=%s org_id=%q", machineID, userID, orgID)
	return machineID, nil
}

// DisconnectClient marks a connected client as disconnected using id (machine ksuid)
// Note: This should only be called when connection_count is already 0 (last subscription)
func (c *Client) DisconnectClient(ctx context.Context, userID, machineID, orgID string) error {
	query := c.client.From("connected_clients").
		Update(map[string]interface{}{
			"disconnected_at": time.Now().Format(time.RFC3339),
			// Also ensure connection_count is 0 when disconnecting
			// This ensures the UPDATE event includes connection_count for frontend
			"connection_count": 0,
		}, "", "").
		Eq("id", machineID).
		Eq("user_id", userID).
		Eq("org_id", orgID). // orgID is empty string for personal accounts
		Is("disconnected_at", "null")

	_, _, err := query.Execute()

	if err != nil {
		return fmt.Errorf("failed to disconnect client: %w", err)
	}

	return nil
}

// DisconnectClientByMachineID marks a connected client as disconnected using only the machine ID
// This is used when a relay crashes and we need to disconnect clients without knowing userID/orgID
func (c *Client) DisconnectClientByMachineID(ctx context.Context, machineID string) error {
	query := c.client.From("connected_clients").
		Update(map[string]interface{}{
			"disconnected_at": time.Now().Format(time.RFC3339),
		}, "", "").
		Eq("id", machineID).
		Is("disconnected_at", "null")

	_, _, err := query.Execute()

	if err != nil {
		return fmt.Errorf("failed to disconnect client by machine ID: %w", err)
	}

	return nil
}

// UpdateConnectionCount updates the connection_count for a specific machine+user+org combination
func (c *Client) UpdateConnectionCount(ctx context.Context, machineID, userID, orgID string, count int) error {
	query := c.client.From("connected_clients").
		Update(map[string]interface{}{
			"connection_count": count,
		}, "", "").
		Eq("id", machineID).
		Eq("user_id", userID).
		Eq("org_id", orgID) // orgID is empty string for personal accounts

	_, _, err := query.Execute()

	if err != nil {
		return fmt.Errorf("failed to update connection count: %w", err)
	}

	return nil
}

