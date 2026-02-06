package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	supabasego "github.com/supabase-community/supabase-go"
)

type Client struct {
	client *supabasego.Client
}

func NewClient() (*Client, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_SECRET_KEY")

	if url == "" || key == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SECRET_KEY must be set")
	}

	client, err := supabasego.NewClient(url, key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase client: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

// TopicTierResult holds the resolved tier information
type TopicTierResult struct {
	TierName        string
	IsEnterprise    bool
	CustomOverrides *EnterpriseOverride
}

// EnterpriseOverride holds custom rate limits for Enterprise organizations
type EnterpriseOverride struct {
	BurstPerMinute int64 `json:"burst_per_minute"`
	DailyQuota     int64 `json:"daily_quota"`
	MaxPayloadSize int64 `json:"max_payload_size"`
}

// ToTier converts an enterprise override into a Tier.
// Falls back to Scale defaults for any zero/unset field.
// Note: This method is used by the resolver after importing ratelimit package.
// We can't define it here due to import cycle, so it's defined in resolver.go

// LookupTopicTier resolves a topic ID to its tier by:
// 1. Looking up topic in topics table
// 2. Getting the application_id
// 3. Getting user_id or org_id from application
// 4. Looking up subscription (if org_id, check org subscription; if user_id, check user subscription)
// 5. Getting Stripe product name from subscription
// 6. Mapping product name to tier
// 7. If Enterprise, checking for custom overrides
func (c *Client) LookupTopicTier(ctx context.Context, topicID string) (*TopicTierResult, error) {
	// Step 1: Get topic
	var topic struct {
		ID            string `json:"id"`
		ApplicationID string `json:"application_id"`
	}

	data, _, err := c.client.From("topics").
		Select("id,application_id", "exact", false).
		Eq("id", topicID).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("topic not found: %w", err)
	}

	if err := json.Unmarshal(data, &topic); err != nil {
		return nil, fmt.Errorf("failed to parse topic data: %w", err)
	}

	// Step 2: Get application
	var app struct {
		ID     string  `json:"id"`
		UserID *string `json:"user_id"`
		OrgID  *string `json:"org_id"`
	}

	appData, _, err := c.client.From("applications").
		Select("id,user_id,org_id", "exact", false).
		Eq("id", topic.ApplicationID).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	if err := json.Unmarshal(appData, &app); err != nil {
		return nil, fmt.Errorf("failed to parse application data: %w", err)
	}

	// Step 3: Get subscription (prefer org over user)
	var orgID string
	if app.OrgID != nil && *app.OrgID != "" {
		orgID = *app.OrgID
	} else if app.UserID != nil {
		// For user-owned apps, we'd need to check user subscriptions
		// For now, default to Starter if no org subscription
		orgID = ""
	} else {
		return nil, fmt.Errorf("application has no owner")
	}

	// Step 4: Look up subscription
	// Note: We'll need to query Stripe API or store product name in subscription table
	// For now, assume we query subscriptions table and get stripe_subscription_id
	// Then we'd need to call Stripe API to get product name
	// As a temporary solution, we'll check if subscription exists and default to Starter
	// TODO: Integrate with Stripe API to get actual product name

	result := &TopicTierResult{
		TierName: "starter", // default (will be resolved via config in resolver)
	}

	if orgID == "" {
		// No org, default to Starter
		return result, nil
	}

	// Check for subscription
	var subscriptions []struct {
		ID                   string `json:"id"`
		StripeSubscriptionID string `json:"stripe_subscription_id"`
		Subscribed           bool   `json:"subscribed"`
	}

	subData, _, err := c.client.From("subscriptions").
		Select("id,stripe_subscription_id,subscribed", "exact", false).
		Eq("org_id", orgID).
		Eq("subscribed", "true").
		Execute()

	if err != nil || len(subData) == 0 || string(subData) == "null" || string(subData) == "[]" {
		// No subscription, default to Starter
		return result, nil
	}

	if err := json.Unmarshal(subData, &subscriptions); err != nil || len(subscriptions) == 0 {
		// Parse error or empty result, default to Starter
		return result, nil
	}

	// TODO: Use subscriptions[0] to get Stripe subscription ID and call Stripe API to get product name
	_ = subscriptions[0] // Will be used when Stripe integration is complete
	// For now, we'll need to store product name in subscription table or call Stripe
	// As a placeholder, check for Enterprise override first

	// Check for Enterprise custom overrides
	override, err := c.GetEnterpriseOverrides(ctx, orgID)
	if err == nil && override != nil {
		result.IsEnterprise = true
		result.CustomOverrides = override
		return result, nil
	}

	// Default to Starter until Stripe integration is complete
	return result, nil
}

// GetEnterpriseOverrides fetches custom rate limit overrides for an Enterprise organization.
// Returns nil if no overrides exist (org is not Enterprise or no custom limits set).
func (c *Client) GetEnterpriseOverrides(ctx context.Context, orgID string) (*EnterpriseOverride, error) {
	// Check if organizations table has settings column
	// For now, we'll assume a simple approach: check if org exists and has custom settings
	// In production, this would be in organization_settings JSONB column or dedicated table

	// Placeholder: return nil (no overrides)
	// TODO: Implement actual lookup from organization_settings table
	return nil, nil
}

// IncrementAnonTopicCount increments the request_count for an anonymous topic.
// This is called asynchronously after each webhook publish.
func (c *Client) IncrementAnonTopicCount(ctx context.Context, topicID string) error {
	// Read current count
	var topic struct {
		RequestCount int64 `json:"request_count"`
	}

	data, _, err := c.client.From("anonymous_topics").
		Select("request_count", "exact", false).
		Eq("id", topicID).
		Single().
		Execute()

	if err != nil {
		log.Printf("Failed to read anonymous topic for increment: %v", err)
		return err
	}

	if err := json.Unmarshal(data, &topic); err != nil {
		log.Printf("Failed to parse anonymous topic data: %v", err)
		return err
	}

	// Update with incremented count
	_, _, err = c.client.From("anonymous_topics").
		Update(map[string]interface{}{
			"request_count": topic.RequestCount + 1,
			"last_used_at":  time.Now().Format(time.RFC3339),
		}, "", "").
		Eq("id", topicID).
		Execute()

	if err != nil {
		log.Printf("Failed to increment anonymous topic count: %v", err)
		return err
	}

	return nil
}
