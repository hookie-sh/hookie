package ratelimit

import "testing"

func TestTierByName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Tier
	}{
		{"starter", "starter", Starter},
		{"pro", "pro", Pro},
		{"scale", "scale", Scale},
		{"unknown", "unknown", Starter}, // defaults to Starter
		{"empty", "", Starter},
		{"case insensitive", "PRO", Pro},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TierByName(tt.input)
			if result.Name != tt.expected.Name {
				t.Errorf("TierByName(%q) = %v, want %v", tt.input, result.Name, tt.expected.Name)
			}
		})
	}
}

func TestRateLimitKey(t *testing.T) {
	tests := []struct {
		prefix   string
		topicID  string
		expected string
	}{
		{"anon", "anon_abc123", "rl:anon:anon_abc123"},
		{"topics", "topic_xyz789", "rl:topics:topic_xyz789"},
	}

	for _, tt := range tests {
		t.Run(tt.prefix+"_"+tt.topicID, func(t *testing.T) {
			result := RateLimitKey(tt.prefix, tt.topicID)
			if result != tt.expected {
				t.Errorf("RateLimitKey(%q, %q) = %q, want %q", tt.prefix, tt.topicID, result, tt.expected)
			}
		})
	}
}
