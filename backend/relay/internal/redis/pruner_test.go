package redis

import (
	"testing"
)

func TestParseStreamIDTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTs    int64
		wantError bool
	}{
		{
			name:      "valid",
			input:     "1234567890123-0",
			wantTs:    1234567890123,
			wantError: false,
		},
		{
			name:      "valid with sequence",
			input:     "1234567890123-456",
			wantTs:    1234567890123,
			wantError: false,
		},
		{
			name:      "empty",
			input:     "",
			wantTs:    0,
			wantError: true,
		},
		{
			name:      "invalid",
			input:     "invalid",
			wantTs:    0,
			wantError: true,
		},
		{
			name:      "no hyphen",
			input:     "1234567890123",
			wantTs:    0,
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStreamIDTimestamp(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("parseStreamIDTimestamp() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got != tt.wantTs {
				t.Errorf("parseStreamIDTimestamp() = %d, want %d", got, tt.wantTs)
			}
		})
	}
}
