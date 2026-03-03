package redis

import (
	"testing"
)

func TestStreamKey(t *testing.T) {
	tests := []struct {
		name      string
		topicID   string
		anonymous bool
		want      string
	}{
		{
			name:      "topic",
			topicID:   "topic1",
			anonymous: false,
			want:      "topics:topic1",
		},
		{
			name:      "anon",
			topicID:   "anon_xxx",
			anonymous: true,
			want:      "anon:topics:anon_xxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StreamKey(tt.topicID, tt.anonymous)
			if got != tt.want {
				t.Errorf("StreamKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
