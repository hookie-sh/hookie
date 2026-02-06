package config

import "time"

// Redis key prefixes
const (
	RedisPrefixRateLimit = "rl"
	RedisPrefixTierCache = "tier"
	RedisPrefixAnon      = "anon"
	RedisPrefixTopics    = "topics"
)

// Redis key patterns
const (
	RedisKeyAnonChannels = "anon:channels"
	RedisKeyAnonMeta     = "anon:meta:"
	RedisKeyAnonIP       = "anon:ip:"
)

// Stream key prefixes
const (
	StreamPrefixTopics    = "topics:"
	StreamPrefixAnonTopics = "anon:topics:"
)

// Rate limit window names
const (
	WindowMinute = "minute"
	WindowDay    = "day"
)

// Tier names
const (
	TierAnon      = "anon"
	TierStarter   = "starter"
	TierPro       = "pro"
	TierScale     = "scale"
	TierEnterprise = "enterprise"
)

// Cache configuration
const (
	TierCacheTTL = 5 * time.Minute
)

// Rate limit retry after values
const (
	RetryAfterMinute = 60 // seconds
)

// Error messages
const (
	ErrInvalidPathParams = "Invalid path parameters"
	ErrTopicNotFound     = "Topic not found"
	ErrTopicDisabled     = "Topic disabled"
	ErrReadBody          = "Failed to read request body"
	ErrPublishWebhook    = "Failed to process webhook"
	ErrInternalServer    = "Internal server error"
)

// Rate limit error messages
const (
	ErrRateLimitExceededBurst = "Rate limit exceeded. Please slow down."
	ErrRateLimitExceededQuota = "Daily quota exceeded. Upgrade at https://hookie.sh/pricing"
)

// Upgrade URL
const (
	UpgradeURL = "https://hookie.sh/pricing"
)

// BuildRateLimitKey builds a Redis rate limit key.
// Format: rl:{prefix}:{topicID}
func BuildRateLimitKey(prefix string, topicID string) string {
	return RedisPrefixRateLimit + ":" + prefix + ":" + topicID
}

// BuildTierCacheKey builds a Redis tier cache key.
// Format: tier:{topicID}
func BuildTierCacheKey(topicID string) string {
	return RedisPrefixTierCache + ":" + topicID
}

// BuildAnonMetaKey builds a Redis anonymous metadata key.
// Format: anon:meta:{topicID}
func BuildAnonMetaKey(topicID string) string {
	return RedisKeyAnonMeta + topicID
}

// BuildStreamKey builds a Redis stream key for topics.
// Format: topics:{topicID} or anon:topics:{topicID}
func BuildStreamKey(isAnon bool, topicID string) string {
	if isAnon {
		return StreamPrefixAnonTopics + topicID
	}
	return StreamPrefixTopics + topicID
}
