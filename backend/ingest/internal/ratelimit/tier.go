package ratelimit

import (
	"strings"

	"github.com/hookie/ingest/internal/config"
)

type Tier struct {
	Name           string
	BurstPerMinute int64
	DailyQuota     int64
	MaxPayloadSize int64
}

// Default tiers — aligned with pricing plans
var (
	Anon    = Tier{config.TierAnon, 10, 100, 64 * 1024}           // anonymous usage via /anon route
	Starter = Tier{config.TierStarter, 60, 2_000, 256 * 1024}      // $9/mo — 50k webhooks/mo
	Pro     = Tier{config.TierPro, 200, 20_000, 1 << 20}            // $29/mo — 500k webhooks/mo
	Scale   = Tier{config.TierScale, 500, 200_000, 1 << 20}         // $99/mo — 5M webhooks/mo
	// Enterprise: custom limits per org, stored in Supabase
)

// TierByName resolves a tier string (from Stripe subscription) to a Tier.
// Authenticated users without a subscription default to Starter.
func TierByName(name string) Tier {
	switch strings.ToLower(name) {
	case config.TierStarter:
		return Starter
	case config.TierPro:
		return Pro
	case config.TierScale:
		return Scale
	default:
		return Starter
	}
}

// EnterpriseOverride holds custom rate limits set by the Hookie team
// for a specific organization. Stored in Supabase org settings.
type EnterpriseOverride struct {
	BurstPerMinute int64 `json:"burst_per_minute"`
	DailyQuota     int64 `json:"daily_quota"`
	MaxPayloadSize int64 `json:"max_payload_size"`
}

// ToTier converts an enterprise override into a Tier.
// Falls back to Scale defaults for any zero/unset field.
func (o EnterpriseOverride) ToTier() Tier {
	t := Tier{Name: config.TierEnterprise}
	t.BurstPerMinute = fallback(o.BurstPerMinute, Scale.BurstPerMinute)
	t.DailyQuota = fallback(o.DailyQuota, Scale.DailyQuota)
	t.MaxPayloadSize = fallback(o.MaxPayloadSize, Scale.MaxPayloadSize)
	return t
}

func fallback(val, defaultVal int64) int64 {
	if val == 0 {
		return defaultVal
	}
	return val
}

// RateLimitKey builds a Redis key for rate limiting.
// Format: rl:{prefix}:{topicID}
func RateLimitKey(prefix string, topicID string) string {
	return config.BuildRateLimitKey(prefix, topicID)
}
