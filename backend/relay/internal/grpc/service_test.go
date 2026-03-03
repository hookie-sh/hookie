package grpc

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/hookie/relay/internal/redis"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func TestParseTimestamp(t *testing.T) {
	s := &Service{}
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{"empty", "", 0},
		{"valid", "1234567890", 1234567890},
		{"invalid", "invalid", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.parseTimestamp(tt.input)
			if got != tt.want {
				t.Errorf("parseTimestamp(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTimestampFromISO(t *testing.T) {
	s := &Service{}
	tests := []struct {
		name     string
		input    string
		wantZero bool
	}{
		{"empty", "", true},
		{"valid RFC3339", "2024-01-15T12:00:00Z", false},
		{"valid RFC3339Nano", "2024-01-15T12:00:00.123456789Z", false},
		{"invalid", "invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.parseTimestampFromISO(tt.input)
			if tt.wantZero && got != 0 {
				t.Errorf("parseTimestampFromISO(%q) = %d, want 0", tt.input, got)
			}
			if !tt.wantZero && got == 0 {
				t.Errorf("parseTimestampFromISO(%q) = 0, want non-zero", tt.input)
			}
		})
	}
}

func TestGetChannelBufferSize(t *testing.T) {
	envVar := "RELAY_EVENTS_CHANNEL_BUFFER"
	// Restore original value after test
	orig := os.Getenv(envVar)
	defer func() {
		if orig != "" {
			os.Setenv(envVar, orig)
		} else {
			os.Unsetenv(envVar)
		}
	}()

	tests := []struct {
		name        string
		envValue    string
		defaultSize int
		want        int
	}{
		{"unset", "", 5000, 5000},
		{"valid", "1000", 5000, 1000},
		{"invalid", "invalid", 5000, 5000},
		{"zero", "0", 5000, 5000},
		{"negative", "-1", 5000, 5000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(envVar, tt.envValue)
			} else {
				os.Unsetenv(envVar)
			}
			got := getChannelBufferSize(envVar, tt.defaultSize)
			if got != tt.want {
				t.Errorf("getChannelBufferSize(env=%q, default=%d) = %d, want %d", tt.envValue, tt.defaultSize, got, tt.want)
			}
		})
	}
}

// mockTopicAppLookup for testing convertToProtoEvent
type mockTopicAppLookup struct {
	appIDs map[string]string
	err    error
}

func (m *mockTopicAppLookup) GetTopicApplicationID(_ context.Context, topicID string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.appIDs[topicID], nil
}

// mockDisconnectClient for testing DisconnectAllClients
type mockDisconnectClient struct {
	calls []struct {
		userID   string
		machineID string
		orgID    string
	}
	err error
}

func (m *mockDisconnectClient) DisconnectClient(_ context.Context, userID, machineID, orgID string) error {
	m.calls = append(m.calls, struct {
		userID    string
		machineID string
		orgID     string
	}{userID, machineID, orgID})
	return m.err
}

func TestConvertToProtoEvent(t *testing.T) {
	ctx := context.Background()
	baseFields := map[string]string{
		"method": "POST", "path": "/webhook", "timestamp": "1234567890",
	}

	t.Run("valid", func(t *testing.T) {
		s := &Service{
			topicAppLookup: &mockTopicAppLookup{appIDs: map[string]string{"topic1": "app1"}},
		}
		event := redis.StreamEvent{StreamKey: "topics:topic1", ID: "1", Fields: baseFields}
		got, err := s.convertToProtoEvent(ctx, event)
		if err != nil {
			t.Fatalf("convertToProtoEvent() error = %v", err)
		}
		if got.AppId != "app1" || got.TopicId != "topic1" {
			t.Errorf("convertToProtoEvent() AppId=%q TopicId=%q, want app1 topic1", got.AppId, got.TopicId)
		}
		if got.EventType != "webhook" {
			t.Errorf("convertToProtoEvent() EventType=%q, want webhook", got.EventType)
		}
	})

	t.Run("invalid prefix", func(t *testing.T) {
		s := &Service{topicAppLookup: &mockTopicAppLookup{}}
		event := redis.StreamEvent{StreamKey: "invalid:topic1", Fields: baseFields}
		_, err := s.convertToProtoEvent(ctx, event)
		if err == nil {
			t.Error("convertToProtoEvent() expected error for invalid prefix")
		}
	})

	t.Run("empty topic ID", func(t *testing.T) {
		s := &Service{topicAppLookup: &mockTopicAppLookup{}}
		event := redis.StreamEvent{StreamKey: "topics:", Fields: baseFields}
		_, err := s.convertToProtoEvent(ctx, event)
		if err == nil {
			t.Error("convertToProtoEvent() expected error for empty topic ID")
		}
	})

	t.Run("app lookup error", func(t *testing.T) {
		s := &Service{
			topicAppLookup: &mockTopicAppLookup{err: errors.New("not found")},
		}
		event := redis.StreamEvent{StreamKey: "topics:topic1", Fields: baseFields}
		_, err := s.convertToProtoEvent(ctx, event)
		if err == nil {
			t.Error("convertToProtoEvent() expected error when app lookup fails")
		}
	})

	t.Run("cache hit", func(t *testing.T) {
		mock := &mockTopicAppLookup{appIDs: map[string]string{"topic1": "app1"}}
		s := &Service{topicAppLookup: mock}
		event := redis.StreamEvent{StreamKey: "topics:topic1", Fields: baseFields}
		got1, err := s.convertToProtoEvent(ctx, event)
		if err != nil {
			t.Fatalf("first call: %v", err)
		}
		got2, err := s.convertToProtoEvent(ctx, event)
		if err != nil {
			t.Fatalf("second call: %v", err)
		}
		if got1.AppId != "app1" || got2.AppId != "app1" {
			t.Errorf("AppId: got1=%q got2=%q", got1.AppId, got2.AppId)
		}
		// Second call uses cache - mock would only be called once if we tracked it
		// For now just verify both return correct app1
	})
}

func TestGetTotalConnectionsForDBMachine(t *testing.T) {
	t.Run("empty ID", func(t *testing.T) {
		s := &Service{}
		got := s.getTotalConnectionsForDBMachine("")
		if got != 0 {
			t.Errorf("getTotalConnectionsForDBMachine(\"\") = %d, want 0", got)
		}
	})

	t.Run("unknown machine", func(t *testing.T) {
		s := &Service{}
		got := s.getTotalConnectionsForDBMachine("unknown")
		if got != 0 {
			t.Errorf("getTotalConnectionsForDBMachine(\"unknown\") = %d, want 0", got)
		}
	})

	t.Run("one context two subs", func(t *testing.T) {
		s := &Service{}
		ctx1 := "user1:mach1:null"
		ms := &machineSubscriptions{
			subscriptions: []*clientSubscription{{}, {}},
			machineID:     "mach1",
			orgID:         "",
		}
		s.machines.Store(ctx1, ms)
		s.dbMachineToContexts.Store("mach1", []string{ctx1})
		got := s.getTotalConnectionsForDBMachine("mach1")
		if got != 2 {
			t.Errorf("getTotalConnectionsForDBMachine(\"mach1\") = %d, want 2", got)
		}
	})

	t.Run("two contexts", func(t *testing.T) {
		s := &Service{}
		ctx1, ctx2 := "user1:mach1:null", "user2:mach1:org1"
		ms1 := &machineSubscriptions{subscriptions: []*clientSubscription{{}}, machineID: "mach1"}
		ms2 := &machineSubscriptions{subscriptions: []*clientSubscription{{}}, machineID: "mach1"}
		s.machines.Store(ctx1, ms1)
		s.machines.Store(ctx2, ms2)
		s.dbMachineToContexts.Store("mach1", []string{ctx1, ctx2})
		got := s.getTotalConnectionsForDBMachine("mach1")
		if got != 2 {
			t.Errorf("getTotalConnectionsForDBMachine(\"mach1\") = %d, want 2", got)
		}
	})
}

func TestDisconnectClientByMachineID(t *testing.T) {
	var cancelCalled bool
	cancelFn := func() { cancelCalled = true }
	sub := &clientSubscription{cancel: cancelFn}
	ms := &machineSubscriptions{
		subscriptions: []*clientSubscription{sub},
		machineID:     "mach1",
	}
	s := &Service{}
	s.machines.Store("ctx1", ms)
	s.dbMachineToContexts.Store("mach1", []string{"ctx1"})

	s.DisconnectClientByMachineID("mach1")

	if !cancelCalled {
		t.Error("DisconnectClientByMachineID: cancel was not called")
	}
	if _, ok := s.machines.Load("ctx1"); ok {
		t.Error("DisconnectClientByMachineID: machines still contains ctx1")
	}
	if _, ok := s.dbMachineToContexts.Load("mach1"); ok {
		t.Error("DisconnectClientByMachineID: dbMachineToContexts still contains mach1")
	}
}

func TestDisconnectAllClients(t *testing.T) {
	mock := &mockDisconnectClient{}
	s := &Service{disconnectClient: mock}
	ctx1, ctx2 := "user1:mach1:null", "user2:mach2:org2"
	ms1 := &machineSubscriptions{machineID: "mach1", orgID: ""}
	ms2 := &machineSubscriptions{machineID: "mach2", orgID: "org2"}
	s.machines.Store(ctx1, ms1)
	s.machines.Store(ctx2, ms2)
	s.dbMachineToContexts.Store("mach1", []string{ctx1})
	s.dbMachineToContexts.Store("mach2", []string{ctx2})

	s.DisconnectAllClients(context.Background())

	if len(mock.calls) != 2 {
		t.Errorf("DisconnectAllClients: expected 2 calls, got %d", len(mock.calls))
	}
	// Check "null" orgID is converted to ""
	var foundNull, foundOrg2 bool
	for _, c := range mock.calls {
		if c.userID == "user1" && c.orgID == "" {
			foundNull = true
		}
		if c.userID == "user2" && c.orgID == "org2" {
			foundOrg2 = true
		}
	}
	if !foundNull {
		t.Error("DisconnectAllClients: expected call with orgID=\"\" for user1:mach1:null")
	}
	if !foundOrg2 {
		t.Error("DisconnectAllClients: expected call with orgID=org2 for user2:mach2:org2")
	}
}

func TestExtractClientIP(t *testing.T) {
	t.Run("x-forwarded-for", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-forwarded-for", "1.2.3.4"))
		got := extractClientIP(ctx)
		if got != "1.2.3.4" {
			t.Errorf("extractClientIP() = %q, want 1.2.3.4", got)
		}
	})

	t.Run("x-forwarded-for multiple", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-forwarded-for", "1.2.3.4, 5.6.7.8"))
		got := extractClientIP(ctx)
		if got != "1.2.3.4" {
			t.Errorf("extractClientIP() = %q, want 1.2.3.4", got)
		}
	})

	t.Run("peer fallback", func(t *testing.T) {
		addr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 12345}
		ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: addr})
		got := extractClientIP(ctx)
		if got != "192.168.1.1" {
			t.Errorf("extractClientIP() = %q, want 192.168.1.1", got)
		}
	})

	t.Run("no metadata no peer", func(t *testing.T) {
		got := extractClientIP(context.Background())
		if got != "" {
			t.Errorf("extractClientIP() = %q, want empty", got)
		}
	})
}
