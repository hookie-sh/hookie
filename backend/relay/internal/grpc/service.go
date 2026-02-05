package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hookie/relay/internal/auth"
	"github.com/hookie/relay/internal/redis"
	"github.com/hookie/relay/internal/supabase"
	"github.com/hookie/relay/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Service struct {
	proto.UnimplementedRelayServiceServer
	
	subscriber *redis.Subscriber
	verifier   *auth.Verifier
	supabase   *supabase.Client
	
	// Map of client connections to their subscription channels
	clients sync.Map // map[string]*clientSubscription
}

type clientSubscription struct {
	appID    string
	topicID  string
	eventsCh chan redis.StreamEvent
	cancel   context.CancelFunc
}

func NewService(subscriber *redis.Subscriber, verifier *auth.Verifier, supabaseClient *supabase.Client) *Service {
	return &Service{
		subscriber: subscriber,
		verifier:   verifier,
		supabase:   supabaseClient,
	}
}

func (s *Service) extractTokenInfo(ctx context.Context) (*auth.TokenInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	tokenInfo, err := s.verifier.VerifyToken(ctx, tokens[0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
	}

	return tokenInfo, nil
}

func (s *Service) Subscribe(req *proto.SubscribeRequest, stream proto.RelayService_SubscribeServer) error {
	ctx := stream.Context()

	// Verify authentication
	tokenInfo, err := s.extractTokenInfo(ctx)
	if err != nil {
		return err
	}

	// Use org_id from request if provided, otherwise use from token
	orgID := req.OrgId
	if orgID == "" {
		orgID = tokenInfo.OrgID
	}

	var appID string
	var topicIDs []string

	if req.TopicId != "" {
		// Topic ID provided - look up application_id from topic
		appID, err = s.supabase.GetTopicApplicationID(ctx, req.TopicId)
		if err != nil {
			return status.Error(codes.NotFound, fmt.Sprintf("topic not found: %v", err))
		}

		// Verify user has access to the application
		if err := s.supabase.VerifyApplicationAccess(ctx, tokenInfo.UserID, appID, orgID); err != nil {
			return status.Error(codes.PermissionDenied, fmt.Sprintf("access denied: %v", err))
		}

		// Verify the topic belongs to the application (double-check)
		if err := s.supabase.VerifyTopicAccess(ctx, tokenInfo.UserID, appID, req.TopicId, orgID); err != nil {
			return status.Error(codes.PermissionDenied, fmt.Sprintf("access denied: %v", err))
		}

		topicIDs = []string{req.TopicId}
	} else {
		// No topic ID - subscribe to all topics for the application
		appID = req.AppId
		if appID == "" {
			return status.Error(codes.InvalidArgument, "app_id is required when topic_id is not specified")
		}

		// Verify user has access to the application
		if err := s.supabase.VerifyApplicationAccess(ctx, tokenInfo.UserID, appID, orgID); err != nil {
			return status.Error(codes.PermissionDenied, fmt.Sprintf("access denied: %v", err))
		}

		// Query all topics for the application
		var topics []struct {
			ID string `json:"id"`
		}

		data, _, err := s.supabase.GetClient().From("topics").
			Select("id", "exact", false).
			Eq("application_id", appID).
			Execute()

		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to fetch topics: %v", err))
		}

		if err := json.Unmarshal(data, &topics); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to parse topics: %v", err))
		}

		topicIDs = make([]string, 0, len(topics))
		for _, topic := range topics {
			topicIDs = append(topicIDs, topic.ID)
		}
	}

	// Create event channel
	eventsCh := make(chan redis.StreamEvent, 100)
	clientID := fmt.Sprintf("%s-%d", tokenInfo.UserID, stream.Context().Value("stream_id"))

	// Subscribe to Redis streams
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var subscribeErr error
	if len(topicIDs) == 1 {
		subscribeErr = s.subscriber.SubscribeToTopic(topicIDs[0], eventsCh)
	} else {
		subscribeErr = s.subscriber.SubscribeToApplication(topicIDs, eventsCh)
	}

	if subscribeErr != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to subscribe: %v", subscribeErr))
	}

	// Store client subscription
	s.clients.Store(clientID, &clientSubscription{
		appID:    appID,
		topicID:  req.TopicId,
		eventsCh: eventsCh,
		cancel:   cancel,
	})
	defer s.clients.Delete(clientID)

	// Stream events to client
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-eventsCh:
			if !ok {
				return status.Error(codes.Internal, "event channel closed")
			}

			// Convert Redis event to proto Event
			protoEvent, err := s.convertToProtoEvent(ctx, event)
			if err != nil {
				log.Printf("Error converting event: %v", err)
				continue // Skip this event but continue processing
			}
			if err := stream.Send(protoEvent); err != nil {
				log.Printf("Error sending event to client %s: %v", clientID, err)
				return err
			}
		}
	}
}

func (s *Service) convertToProtoEvent(ctx context.Context, event redis.StreamEvent) (*proto.Event, error) {
	// Parse stream key to extract topic_id: webhook:events:{topicId}
	streamKey := event.StreamKey
	prefix := "webhook:events:"
	if !strings.HasPrefix(streamKey, prefix) {
		return nil, fmt.Errorf("invalid stream key format: %s", streamKey)
	}

	topicID := streamKey[len(prefix):]
	if topicID == "" {
		return nil, fmt.Errorf("topic ID not found in stream key: %s", streamKey)
	}

	// Look up application_id from topic
	appID, err := s.supabase.GetTopicApplicationID(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application ID for topic %s: %w", topicID, err)
	}

	return &proto.Event{
		Method:        event.Fields["method"],
		Url:           event.Fields["url"],
		Path:          event.Fields["path"],
		Query:         event.Fields["query"],
		Headers:       event.Fields["headers"],
		Body:          event.Fields["body"],
		ContentType:   event.Fields["content_type"],
		ContentLength: event.Fields["content_length"],
		RemoteAddr:    event.Fields["remote_addr"],
		Timestamp:     s.parseTimestamp(event.Fields["timestamp"]),
		AppId:         appID,
		TopicId:       topicID,
	}, nil
}

func (s *Service) parseTimestamp(ts string) int64 {
	if ts == "" {
		return 0
	}
	
	val, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func (s *Service) ListApplications(ctx context.Context, req *proto.ListApplicationsRequest) (*proto.ListApplicationsResponse, error) {
	log.Printf("[ListApplications] Starting request with org_id=%q", req.OrgId)
	
	tokenInfo, err := s.extractTokenInfo(ctx)
	if err != nil {
		log.Printf("[ListApplications] Failed to extract token info: %v", err)
		return nil, err
	}
	log.Printf("[ListApplications] Token info extracted - user_id=%q, org_id=%q", tokenInfo.UserID, tokenInfo.OrgID)

	// Query user-owned applications
	// Note: Using service role key should bypass RLS, but let's verify the query works
	query := s.supabase.GetClient().From("applications").
		Select("id,name,description,created_at,updated_at,user_id,org_id", "exact", false).
		Eq("user_id", tokenInfo.UserID)
	
	log.Printf("[ListApplications] About to execute query for user_id=%q", tokenInfo.UserID)
	userOwnedData, count, err := query.Execute()
	
	log.Printf("[ListApplications] User-owned apps query - data length=%d bytes, count=%d, raw data=%s, error=%v", len(userOwnedData), count, string(userOwnedData), err)
	if len(userOwnedData) > 0 {
		log.Printf("[ListApplications] Raw JSON response: %s", string(userOwnedData))
	} else {
		log.Printf("[ListApplications] WARNING: Query returned empty data, but no error. This might indicate RLS is being applied.")
	}
	if err != nil {
		log.Printf("[ListApplications] User-owned apps query failed: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to fetch user-owned applications: %v", err))
	}

	var userOwnedApps []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		UserID      string `json:"user_id"`
		OrgID       string `json:"org_id"`
	}

	if err := json.Unmarshal(userOwnedData, &userOwnedApps); err != nil {
		log.Printf("[ListApplications] Failed to parse user-owned apps: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to parse user-owned applications: %v", err))
	}

	log.Printf("[ListApplications] Found %d user-owned applications", len(userOwnedApps))

	// Use a map to deduplicate applications by ID
	appMap := make(map[string]struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		UserID      string `json:"user_id"`
		OrgID       string `json:"org_id"`
	})

	// Add user-owned apps to map
	for _, app := range userOwnedApps {
		appMap[app.ID] = app
	}

	// Query org applications for organizations the user has access to
	// TODO: In the future, query Clerk API to get all orgs the user belongs to
	// For now, use the active org from the token
	if tokenInfo.OrgID != "" {
		log.Printf("[ListApplications] Querying org applications for org_id=%q", tokenInfo.OrgID)
		orgData, _, err := s.supabase.GetClient().From("applications").
			Select("id,name,description,created_at,updated_at,user_id,org_id", "exact", false).
			Eq("org_id", tokenInfo.OrgID).
			Execute()
		
		log.Printf("[ListApplications] Org apps query - data length=%d bytes, error=%v", len(orgData), err)
		if err != nil {
			log.Printf("[ListApplications] Org apps query failed: %v", err)
			// Don't fail completely, just log and continue with user-owned apps
		} else {
			var orgApps []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				CreatedAt   string `json:"created_at"`
				UpdatedAt   string `json:"updated_at"`
				UserID      string `json:"user_id"`
				OrgID       string `json:"org_id"`
			}

			if err := json.Unmarshal(orgData, &orgApps); err != nil {
				log.Printf("[ListApplications] Failed to parse org apps: %v", err)
			} else {
				log.Printf("[ListApplications] Found %d org applications", len(orgApps))
				// Add org apps to map (will overwrite duplicates, which is fine)
				for _, app := range orgApps {
					appMap[app.ID] = app
				}
			}
		}
	}

	// Convert map back to slice
	applications := make([]struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		UserID      string `json:"user_id"`
		OrgID       string `json:"org_id"`
	}, 0, len(appMap))
	for _, app := range appMap {
		applications = append(applications, app)
	}

	log.Printf("[ListApplications] Total unique applications after merge: %d", len(applications))

	// If org_id filter is provided in request, filter to only that organization
	if req.OrgId != "" {
		log.Printf("[ListApplications] Filtering applications by org_id=%q", req.OrgId)
		filtered := make([]struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
			UserID      string `json:"user_id"`
			OrgID       string `json:"org_id"`
		}, 0)
		for _, app := range applications {
			if app.OrgID == req.OrgId {
				filtered = append(filtered, app)
			}
		}
		applications = filtered
		log.Printf("[ListApplications] After org_id filter: %d applications", len(applications))
	}

	log.Printf("[ListApplications] Parsed %d applications from response", len(applications))
	for i, app := range applications {
		log.Printf("[ListApplications] Application[%d]: id=%q, name=%q, user_id=%q, org_id=%q", i, app.ID, app.Name, app.UserID, app.OrgID)
	}

	// Convert to proto messages
	protoApps := make([]*proto.Application, 0, len(applications))
	for _, app := range applications {
		protoApps = append(protoApps, &proto.Application{
			Id:          app.ID,
			Name:        app.Name,
			Description: app.Description,
			CreatedAt:   s.parseTimestampFromISO(app.CreatedAt),
			UpdatedAt:   s.parseTimestampFromISO(app.UpdatedAt),
		})
	}

	log.Printf("[ListApplications] Returning %d applications", len(protoApps))
	return &proto.ListApplicationsResponse{
		Applications: protoApps,
	}, nil
}

func (s *Service) ListTopics(ctx context.Context, req *proto.ListTopicsRequest) (*proto.ListTopicsResponse, error) {
	tokenInfo, err := s.extractTokenInfo(ctx)
	if err != nil {
		return nil, err
	}

	appID := req.AppId
	if appID == "" {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	// Verify user has access to the application
	if err := s.supabase.VerifyApplicationAccess(ctx, tokenInfo.UserID, appID, tokenInfo.OrgID); err != nil {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("access denied: %v", err))
	}

	// Query topics for the application
	var topics []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}

	data, _, err := s.supabase.GetClient().From("topics").
		Select("id,name,description,created_at,updated_at", "exact", false).
		Eq("application_id", appID).
		Execute()

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to fetch topics: %v", err))
	}

	if err := json.Unmarshal(data, &topics); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to parse topics: %v", err))
	}

	// Convert to proto messages
	protoTopics := make([]*proto.Topic, 0, len(topics))
	for _, topic := range topics {
		createdAt := s.parseTimestampFromISO(topic.CreatedAt)
		updatedAt := s.parseTimestampFromISO(topic.UpdatedAt)

		protoTopics = append(protoTopics, &proto.Topic{
			Id:          topic.ID,
			Name:        topic.Name,
			Description: topic.Description,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	return &proto.ListTopicsResponse{
		Topics: protoTopics,
	}, nil
}

func (s *Service) parseTimestampFromISO(iso string) int64 {
	// Parse ISO 8601 timestamp to nanoseconds
	if iso == "" {
		return 0
	}
	
	// Try parsing common ISO 8601 formats
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05Z",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, iso); err == nil {
			return t.UnixNano()
		}
	}
	
	return 0
}

