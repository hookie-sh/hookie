package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hookie/relay/internal/auth"
	"github.com/hookie/relay/internal/redis"
	"github.com/hookie/relay/internal/supabase"
	"github.com/hookie/relay/proto"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Service struct {
	proto.UnimplementedRelayServiceServer
	
	subscriber         *redis.Subscriber
	verifier           *auth.Verifier
	supabase           *supabase.Client
	broadcastListener  interface {
		SubscribeToMachineID(ctx context.Context, machineID string) error
	}
	
	// Map of machine+org contexts to their subscription lists
	machines sync.Map // map[string]*machineSubscriptions
	
	// Map of database machine ID to machine+org contexts
	// This allows us to find all contexts for a given database machine ID
	dbMachineToContexts sync.Map // map[string][]string (machineID -> []machineContextKey)
}

type clientSubscription struct {
	machineID  string // Machine identifier (userID:machineIDFromRequest:orgID)
	eventsCh   chan redis.StreamEvent
	cancel     context.CancelFunc
	stream     proto.RelayService_SubscribeServer // gRPC stream for sending events
}

type machineSubscriptions struct {
	subscriptions []*clientSubscription
	machineID     string // The id (primary key) from the database - the mach_<ksuid> value
	orgID         string
	mu            sync.Mutex
}

func NewService(subscriber *redis.Subscriber, verifier *auth.Verifier, supabaseClient *supabase.Client) *Service {
	return &Service{
		subscriber: subscriber,
		verifier:   verifier,
		supabase:   supabaseClient,
	}
}

// SetBroadcastListener sets the broadcast listener for the service
func (s *Service) SetBroadcastListener(listener interface {
	SubscribeToMachineID(ctx context.Context, machineID string) error
}) {
	s.broadcastListener = listener
}

// getTotalConnectionsForDBMachine counts total active subscriptions across all contexts for a database machine ID
func (s *Service) getTotalConnectionsForDBMachine(dbMachineID string) int {
	if dbMachineID == "" {
		return 0
	}
	
	contextsRaw, ok := s.dbMachineToContexts.Load(dbMachineID)
	if !ok {
		return 0
	}
	
	contexts := contextsRaw.([]string)
	total := 0
	
	for _, machineContextKey := range contexts {
		machineSubsRaw, ok := s.machines.Load(machineContextKey)
		if ok {
			machineSubs := machineSubsRaw.(*machineSubscriptions)
			machineSubs.mu.Lock()
			total += len(machineSubs.subscriptions)
			machineSubs.mu.Unlock()
		}
	}
	
	return total
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

	log.Printf("[Subscribe] Starting subscription request: topic_id=%q app_id=%q machine_id=%q", req.TopicId, req.AppId, req.MachineId)

	// Check for anonymous channel subscription
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-channel-type"); len(vals) > 0 && vals[0] == "anonymous" {
			return s.handleAnonymousSubscribe(req, stream)
		}
	}

	// Verify authentication
	tokenInfo, err := s.extractTokenInfo(ctx)
	if err != nil {
		return err
	}
	
	log.Printf("[Subscribe] Authenticated user: userID=%s", tokenInfo.UserID)

	// Validate machine_id is provided
	if req.MachineId == "" {
		return status.Error(codes.InvalidArgument, "machine_id is required")
	}

	// Validate that exactly one of topic_id or app_id is provided
	provided := 0
	if req.TopicId != "" {
		provided++
	}
	if req.AppId != "" {
		provided++
	}
	if provided != 1 {
		return status.Error(codes.InvalidArgument, "exactly one of topic_id or app_id must be provided")
	}

	var appID string
	var topicIDs []string

	// Get appID first (from topic_id or use req.AppId)
	if req.TopicId != "" {
		// Topic ID provided - subscribe to a specific topic
		var err error
		appID, err = s.supabase.GetTopicApplicationID(ctx, req.TopicId)
		if err != nil {
			return status.Error(codes.NotFound, fmt.Sprintf("topic not found: %v", err))
		}
		topicIDs = []string{req.TopicId}
	} else if req.AppId != "" {
		// App ID provided - subscribe to all topics for the application
		appID = req.AppId
	}

	// Determine org_id from the application (not from request/token)
	// This ensures we use the actual org_id that owns the application
	orgID, err := s.supabase.GetApplicationOrgID(ctx, appID)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to get application org_id: %v", err))
	}
	// orgID will be empty string for user-owned apps, or the org ID for org-owned apps

	log.Printf("[Subscribe] Determined orgID=%q from application=%s (Request orgID=%q, Token orgID=%q), MachineID=%s", 
		orgID, appID, req.OrgId, tokenInfo.OrgID, req.MachineId)

	// Verify user has access to the application (using the org_id from the application)
	if err := s.supabase.VerifyApplicationAccess(ctx, tokenInfo.UserID, appID, orgID); err != nil {
		return status.Error(codes.PermissionDenied, fmt.Sprintf("access denied: %v", err))
	}

	if req.TopicId != "" {
		// Verify the topic belongs to the application (double-check)
		if err := s.supabase.VerifyTopicAccess(ctx, tokenInfo.UserID, appID, req.TopicId, orgID); err != nil {
			return status.Error(codes.PermissionDenied, fmt.Sprintf("access denied: %v", err))
		}
	} else if req.AppId != "" {
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

	// Create machine identifier: userID:machineIDFromRequest:orgID (use "null" string for empty org_id)
	orgIDStr := orgID
	if orgIDStr == "" {
		orgIDStr = "null"
	}
	machineID := fmt.Sprintf("%s:%s:%s", tokenInfo.UserID, req.MachineId, orgIDStr)
	
	log.Printf("[Subscribe] Created machineID key=%s for userID=%s machineID=%s orgID=%q", 
		machineID, tokenInfo.UserID, req.MachineId, orgID)

	// Create event channel
	eventsCh := make(chan redis.StreamEvent, 100)

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

	// Get or create machine subscriptions for this machine+org context
	// The machineID key includes orgID, so different orgs get different keys
	machineSubsRaw, loaded := s.machines.LoadOrStore(machineID, &machineSubscriptions{
		subscriptions: make([]*clientSubscription, 0),
		orgID:         orgID,
		mu:            sync.Mutex{},
	})
	machineSubs := machineSubsRaw.(*machineSubscriptions)
	
	// If we loaded an existing entry, verify it has the correct orgID
	if loaded {
		machineSubs.mu.Lock()
		if machineSubs.orgID != orgID {
			log.Printf("[Subscribe] WARNING: Existing machineSubs has orgID=%q but request has orgID=%q for machineID key=%s", 
				machineSubs.orgID, orgID, machineID)
		}
		machineSubs.mu.Unlock()
	}

	machineSubs.mu.Lock()
	isFirstSubscription := len(machineSubs.subscriptions) == 0
	currentSubCount := len(machineSubs.subscriptions)
	machineSubs.mu.Unlock()

	log.Printf("[Subscribe] machineID key=%s, loaded=%v, isFirstSubscription=%v, currentSubCount=%d, orgID=%q", 
		machineID, loaded, isFirstSubscription, currentSubCount, orgID)

	// Upsert database record for connected client (only if first subscription for this machine+org)
	// This ensures we create a record for each unique (machineID, userID, orgID) combination
	var dbMachineID string
	
	if isFirstSubscription {
		var err error
		log.Printf("[UpsertConnectedClient] Creating/updating record: userID=%s machineID=%s orgID=%q", 
			tokenInfo.UserID, req.MachineId, orgID)
		dbMachineID, err = s.supabase.UpsertConnectedClient(ctx, tokenInfo.UserID, req.MachineId, orgID)
		
		if err != nil {
			log.Printf("Warning: failed to upsert connected client record: %v", err)
			// Continue anyway - subscription is already established
			dbMachineID = req.MachineId // Use the machine_id from request as fallback
		} else {
			log.Printf("[UpsertConnectedClient] Success: dbMachineID=%s for userID=%s machineID=%s orgID=%s", dbMachineID, tokenInfo.UserID, req.MachineId, orgID)
			machineSubs.mu.Lock()
			machineSubs.machineID = dbMachineID
			machineSubs.mu.Unlock()
		}
		
		// Track mapping from database machine ID to machine context
		if dbMachineID != "" {
			contextsRaw, _ := s.dbMachineToContexts.LoadOrStore(dbMachineID, []string{})
			contexts := contextsRaw.([]string)
			// Check if this context is already in the list
			found := false
			for _, ctx := range contexts {
				if ctx == machineID {
					found = true
					break
				}
			}
			if !found {
				contexts = append(contexts, machineID)
				s.dbMachineToContexts.Store(dbMachineID, contexts)
			}
		}
	} else {
		// Use existing machineID
		machineSubs.mu.Lock()
		dbMachineID = machineSubs.machineID
		machineSubs.mu.Unlock()
		
		// Track mapping from database machine ID to machine context
		if dbMachineID != "" {
			contextsRaw, _ := s.dbMachineToContexts.LoadOrStore(dbMachineID, []string{})
			contexts := contextsRaw.([]string)
			// Check if this context is already in the list
			found := false
			for _, ctx := range contexts {
				if ctx == machineID {
					found = true
					break
				}
			}
			if !found {
				contexts = append(contexts, machineID)
				s.dbMachineToContexts.Store(dbMachineID, contexts)
			}
		}
	}

	// Create subscription
	subscription := &clientSubscription{
		machineID: machineID,
		eventsCh: eventsCh,
		cancel:   cancel,
		stream:   stream,
	}

	// Add subscription to machine's list
	machineSubs.mu.Lock()
	machineSubs.subscriptions = append(machineSubs.subscriptions, subscription)
	connectionCount := len(machineSubs.subscriptions)
	machineSubs.mu.Unlock()
	
	// Update connection count for this specific machine+user+org combination
	if dbMachineID != "" {
		if err := s.supabase.UpdateConnectionCount(ctx, dbMachineID, tokenInfo.UserID, orgID, connectionCount); err != nil {
			log.Printf("Warning: failed to update connection count: %v", err)
		}
		
		// Subscribe to broadcast channel for this machine_id (only if first subscription)
		if isFirstSubscription && s.broadcastListener != nil {
			if err := s.broadcastListener.SubscribeToMachineID(ctx, dbMachineID); err != nil {
				log.Printf("Warning: failed to subscribe to broadcast channel for machine %s: %v", dbMachineID, err)
			}
		}
	}
	
	// Log every connection
	target := req.TopicId
	if target == "" {
		target = req.AppId
	}
	log.Printf("[Client Connect] UserID=%s MachineID=%s DBMachineID=%s OrgID=%s Target=%s ConnectionCount=%d", 
		tokenInfo.UserID, req.MachineId, dbMachineID, orgID, target, connectionCount)

	defer func() {
		// Remove subscription from machine's list
		machineSubs.mu.Lock()
		remaining := make([]*clientSubscription, 0)
		for _, sub := range machineSubs.subscriptions {
			if sub != subscription {
				remaining = append(remaining, sub)
			}
		}
		machineSubs.subscriptions = remaining
		remainingCount := len(machineSubs.subscriptions)
		isLastSubscription := remainingCount == 0
		machineSubs.mu.Unlock()

		// Update connection count for this specific machine+user+org combination
		if dbMachineID != "" {
			if err := s.supabase.UpdateConnectionCount(ctx, dbMachineID, tokenInfo.UserID, orgID, remainingCount); err != nil {
				log.Printf("Warning: failed to update connection count: %v", err)
			}
		}

		// If this was the last subscription for this machine+org, mark as disconnected
		if isLastSubscription {
			s.machines.Delete(machineID)
			if dbMachineID != "" {
				// Remove from dbMachineToContexts mapping
				contextsRaw, ok := s.dbMachineToContexts.Load(dbMachineID)
				if ok {
					contexts := contextsRaw.([]string)
					remaining := make([]string, 0)
					for _, ctx := range contexts {
						if ctx != machineID {
							remaining = append(remaining, ctx)
						}
					}
					if len(remaining) == 0 {
						s.dbMachineToContexts.Delete(dbMachineID)
					} else {
						s.dbMachineToContexts.Store(dbMachineID, remaining)
					}
				}
				
				disconnectErr := s.supabase.DisconnectClient(ctx, tokenInfo.UserID, req.MachineId, orgID)
				if disconnectErr != nil {
					log.Printf("[Client Disconnect] UserID=%s MachineID=%s DBMachineID=%s OrgID=%s TotalConnections=%d Error=failed to mark disconnected: %v", 
						tokenInfo.UserID, req.MachineId, dbMachineID, orgID, remainingCount, disconnectErr)
				} else {
					log.Printf("[Client Disconnect] UserID=%s MachineID=%s DBMachineID=%s OrgID=%s TotalConnections=%d", 
						tokenInfo.UserID, req.MachineId, dbMachineID, orgID, remainingCount)
				}
			} else {
				log.Printf("[Client Disconnect] UserID=%s MachineID=%s DBMachineID=%s OrgID=%s TotalConnections=%d", 
					tokenInfo.UserID, req.MachineId, dbMachineID, orgID, remainingCount)
			}
		} else {
			// Log disconnection even if not the last subscription
			log.Printf("[Client Disconnect] UserID=%s MachineID=%s DBMachineID=%s OrgID=%s TotalConnections=%d", 
				tokenInfo.UserID, req.MachineId, dbMachineID, orgID, remainingCount)
		}
	}()

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
				log.Printf("Error sending event to machine %s: %v", machineID, err)
				return err
			}
		}
	}
}

func (s *Service) convertToProtoEvent(ctx context.Context, event redis.StreamEvent) (*proto.Event, error) {
	// Parse stream key to extract topic_id: topics:{topicId}
	streamKey := event.StreamKey
	prefix := "topics:"
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
		EventType:     "webhook", // Mark as webhook event
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

	// If app_id is provided, verify user has access to the application
	if appID != "" {
		if err := s.supabase.VerifyApplicationAccess(ctx, tokenInfo.UserID, appID, tokenInfo.OrgID); err != nil {
			return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("access denied: %v", err))
		}
	}

	// Query topics - RLS will automatically filter to only accessible topics
	var topics []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		ApplicationID string `json:"application_id"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}

	query := s.supabase.GetClient().From("topics").
		Select("id,name,description,application_id,created_at,updated_at", "exact", false)

	// If app_id is provided, filter by it
	if appID != "" {
		query = query.Eq("application_id", appID)
	}

	data, _, err := query.Execute()

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
			Id:            topic.ID,
			Name:          topic.Name,
			Description:   topic.Description,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
			ApplicationId: topic.ApplicationID,
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

// DisconnectClientByMachineID disconnects all active connections for a given database machine ID
// This is called when a disconnect event is received from Supabase Realtime
func (s *Service) DisconnectClientByMachineID(dbMachineID string) {
	log.Printf("[DisconnectClientByMachineID] Starting disconnect for machine ID: %s", dbMachineID)
	
	// Find all machine contexts for this database machine ID
	contextsRaw, ok := s.dbMachineToContexts.Load(dbMachineID)
	if !ok {
		log.Printf("[DisconnectClientByMachineID] No active connections found for machine ID: %s", dbMachineID)
		// Log all current mappings for debugging
		s.dbMachineToContexts.Range(func(key, value interface{}) bool {
			log.Printf("[DisconnectClientByMachineID] Active mapping: %s -> %v", key, value)
			return true
		})
		return
	}
	
	contexts := contextsRaw.([]string)
	log.Printf("[DisconnectClientByMachineID] Found %d active connections for machine ID: %s (contexts: %v)", len(contexts), dbMachineID, contexts)
	
	// Disconnect all contexts
	disconnectedCount := 0
	for _, machineID := range contexts {
		machineSubsRaw, ok := s.machines.Load(machineID)
		if !ok {
			log.Printf("[DisconnectClientByMachineID] Machine context %s not found in machines map", machineID)
			continue
		}
		
		machineSubs := machineSubsRaw.(*machineSubscriptions)
		machineSubs.mu.Lock()
		
		subscriptionCount := len(machineSubs.subscriptions)
		log.Printf("[DisconnectClientByMachineID] Cancelling %d subscriptions for context %s", subscriptionCount, machineID)
		
		// Send disconnect event and cancel all subscriptions for this machine
		for i, sub := range machineSubs.subscriptions {
			if sub.stream != nil {
				// Send disconnect event before canceling
				disconnectEvent := &proto.Event{
					EventType: "disconnect",
					Timestamp: time.Now().UnixNano(),
				}
				if err := sub.stream.Send(disconnectEvent); err != nil {
					log.Printf("[DisconnectClientByMachineID] Failed to send disconnect event to subscription %d: %v", i, err)
				} else {
					log.Printf("[DisconnectClientByMachineID] Sent disconnect event to subscription %d", i)
				}
			}
			
			if sub.cancel != nil {
				log.Printf("[DisconnectClientByMachineID] Cancelling subscription %d for context %s", i, machineID)
				// Cancel the context - this will cause the Subscribe function to return ctx.Err()
				// which will close the gRPC stream
				sub.cancel()
				disconnectedCount++
			}
		}
		
		machineSubs.mu.Unlock()
		
		// Remove from machines map
		s.machines.Delete(machineID)
		log.Printf("[DisconnectClientByMachineID] Removed context %s from machines map", machineID)
	}
	
	// Remove from dbMachineToContexts mapping
	s.dbMachineToContexts.Delete(dbMachineID)
	
	log.Printf("[DisconnectClientByMachineID] Successfully disconnected %d subscriptions for machine ID: %s", disconnectedCount, dbMachineID)
}

// DisconnectAllClients marks all active clients as disconnected in the database
// This should be called during graceful shutdown or when the relay crashes
func (s *Service) DisconnectAllClients(ctx context.Context) {
	log.Println("[DisconnectAllClients] Starting disconnect of all active clients...")
	
	disconnectedCount := 0
	errorCount := 0
	
	// Iterate through all database machine IDs
	s.dbMachineToContexts.Range(func(dbMachineIDRaw, contextsRaw interface{}) bool {
		dbMachineID := dbMachineIDRaw.(string)
		contexts := contextsRaw.([]string)
		
		log.Printf("[DisconnectAllClients] Processing machine ID: %s with %d contexts", dbMachineID, len(contexts))
		
		// For each context, extract userID and orgID, then disconnect
		for _, machineContextKey := range contexts {
			machineSubsRaw, ok := s.machines.Load(machineContextKey)
			if !ok {
				log.Printf("[DisconnectAllClients] Machine context %s not found in machines map", machineContextKey)
				continue
			}
			
			machineSubs := machineSubsRaw.(*machineSubscriptions)
			machineSubs.mu.Lock()
			
			// Parse machine context key: userID:machineIDFromRequest:orgID
			parts := strings.Split(machineContextKey, ":")
			if len(parts) != 3 {
				log.Printf("[DisconnectAllClients] Invalid machine context key format: %s", machineContextKey)
				machineSubs.mu.Unlock()
				continue
			}
			
			userID := parts[0]
			machineIDFromRequest := parts[1]
			orgIDStr := parts[2]
			
			// Convert "null" string back to empty string for orgID
			orgID := orgIDStr
			if orgIDStr == "null" {
				orgID = ""
			}
			
			// Use the database machine ID (which is the mach_<ksuid>)
			dbMachineIDToDisconnect := machineSubs.machineID
			if dbMachineIDToDisconnect == "" {
				// Fallback to machineID from request if database machine ID is not set
				dbMachineIDToDisconnect = machineIDFromRequest
			}
			
			machineSubs.mu.Unlock()
			
			// Mark as disconnected in database
			if err := s.supabase.DisconnectClient(ctx, userID, dbMachineIDToDisconnect, orgID); err != nil {
				log.Printf("[DisconnectAllClients] Failed to disconnect client: userID=%s, machineID=%s, orgID=%s, error=%v", 
					userID, dbMachineIDToDisconnect, orgID, err)
				errorCount++
			} else {
				log.Printf("[DisconnectAllClients] Successfully disconnected client: userID=%s, machineID=%s, orgID=%s", 
					userID, dbMachineIDToDisconnect, orgID)
				disconnectedCount++
			}
		}
		
		return true
	})
	
	log.Printf("[DisconnectAllClients] Completed: %d clients disconnected, %d errors", disconnectedCount, errorCount)
}

// extractClientIP extracts the client IP from gRPC context
// Checks x-forwarded-for metadata first (for Fly.io proxy), then falls back to peer address
func extractClientIP(ctx context.Context) string {
	// Check x-forwarded-for from gRPC metadata first (Fly.io proxy)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if xff := md.Get("x-forwarded-for"); len(xff) > 0 {
			// Take the first IP (client IP) from comma-separated list
			ips := strings.Split(xff[0], ",")
			if len(ips) > 0 {
				return strings.TrimSpace(ips[0])
			}
		}
	}
	// Fall back to peer address
	if p, ok := peer.FromContext(ctx); ok {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err == nil {
			return host
		}
		return p.Addr.String()
	}
	return ""
}

// CreateAnonymousChannel creates an anonymous ephemeral channel (no auth required)
func (s *Service) CreateAnonymousChannel(ctx context.Context, req *proto.CreateAnonymousChannelRequest) (*proto.CreateAnonymousChannelResponse, error) {
	// Reject authenticated users - they should use their existing topics
	if _, err := s.extractTokenInfo(ctx); err == nil {
		return nil, status.Error(codes.PermissionDenied, "authenticated users cannot create anonymous channels. Please use your existing topics.")
	}

	// Extract client IP
	ip := extractClientIP(ctx)
	if ip == "" {
		return nil, status.Error(codes.Internal, "failed to extract client IP")
	}

	log.Printf("[CreateAnonymousChannel] Request from IP: %s", ip)

	// Check per-IP limit (max 3 channels per IP)
	count, err := s.subscriber.CheckAnonIPCount(ctx, ip)
	if err != nil {
		log.Printf("[CreateAnonymousChannel] Failed to check IP count: %v", err)
		return nil, status.Error(codes.Internal, "failed to check IP limit")
	}
	if count >= 3 {
		return nil, status.Error(codes.ResourceExhausted, "maximum anonymous channels per IP exceeded (limit: 3)")
	}

	// Generate topic ID: "anon_" + KSUID
	topicID := fmt.Sprintf("anon_%s", ksuid.New().String())

	// Calculate expiry (2 hours from now)
	expiresAt := time.Now().Add(2 * time.Hour)

	// Create channel in Redis
	if err := s.subscriber.CreateAnonChannel(ctx, topicID, ip, expiresAt); err != nil {
		log.Printf("[CreateAnonymousChannel] Failed to create channel: %v", err)
		return nil, status.Error(codes.Internal, "failed to create anonymous channel")
	}

	// Async: Insert into Supabase for analytics (fire-and-forget)
	go func() {
		if err := s.supabase.InsertAnonymousTopic(context.Background(), topicID, ip); err != nil {
			log.Printf("[CreateAnonymousChannel] Failed to insert anonymous topic (non-fatal): %v", err)
		}
	}()

	// Get ingest base URL from environment
	ingestBaseURL := os.Getenv("INGEST_BASE_URL")
	if ingestBaseURL == "" {
		ingestBaseURL = "https://ingest.hookie.sh" // Default fallback
	}
	webhookURL := fmt.Sprintf("%s/anon/%s", ingestBaseURL, topicID)

	// Get anonymous tier limits (10/min, 100/day, 64KB payload)
	// These match the constants in backend/ingest/internal/ratelimit/tier.go
	limits := &proto.AnonymousLimits{
		RequestsPerDay:    100,
		RequestsPerMinute: 10,
		MaxPayloadBytes:   64 * 1024, // 64KB
	}

	log.Printf("[CreateAnonymousChannel] Created channel: %s for IP: %s, expires at: %s", topicID, ip, expiresAt.Format(time.RFC3339))

	return &proto.CreateAnonymousChannelResponse{
		ChannelId:  topicID,
		WebhookUrl: webhookURL,
		ExpiresAt:  expiresAt.Unix(),
		Limits:     limits,
	}, nil
}

// handleAnonymousSubscribe handles anonymous channel subscriptions (no auth required)
func (s *Service) handleAnonymousSubscribe(req *proto.SubscribeRequest, stream proto.RelayService_SubscribeServer) error {
	ctx := stream.Context()
	topicID := req.TopicId

	log.Printf("[handleAnonymousSubscribe] Anonymous subscription request: topic_id=%q", topicID)

	// Validate topic_id is provided
	if topicID == "" {
		return status.Error(codes.InvalidArgument, "topic_id is required for anonymous subscriptions")
	}

	// Validate topicId starts with "anon_"
	if !strings.HasPrefix(topicID, "anon_") {
		return status.Error(codes.InvalidArgument, "invalid anonymous topic ID format")
	}

	// Validate channel exists + not expired
	if err := s.subscriber.ValidateAnonChannel(ctx, topicID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return status.Error(codes.NotFound, "anonymous channel not found")
		}
		if strings.Contains(err.Error(), "expired") {
			return status.Error(codes.NotFound, "anonymous channel expired")
		}
		if strings.Contains(err.Error(), "disabled") {
			return status.Error(codes.PermissionDenied, "anonymous channel disabled")
		}
		return status.Error(codes.Internal, fmt.Sprintf("failed to validate anonymous channel: %v", err))
	}

	// Track connection
	if err := s.subscriber.TrackAnonConnection(ctx, topicID); err != nil {
		log.Printf("[handleAnonymousSubscribe] Warning: failed to track connection: %v", err)
		// Continue anyway - tracking is not critical
	}

	// Remove connection tracking on disconnect
	defer func() {
		if err := s.subscriber.RemoveAnonConnection(ctx, topicID); err != nil {
			log.Printf("[handleAnonymousSubscribe] Warning: failed to remove connection tracking: %v", err)
		}
	}()

	// Create event channel
	eventsCh := make(chan redis.StreamEvent, 100)

	// Subscribe to Redis stream for anonymous topic
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := s.subscriber.SubscribeToAnonymousTopic(topicID, eventsCh); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to subscribe to anonymous topic: %v", err))
	}

	log.Printf("[handleAnonymousSubscribe] Subscribed to anonymous topic: %s", topicID)

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
			protoEvent, err := s.convertToAnonymousProtoEvent(ctx, event, topicID)
			if err != nil {
				log.Printf("[handleAnonymousSubscribe] Error converting event: %v", err)
				continue // Skip this event but continue processing
			}
			if err := stream.Send(protoEvent); err != nil {
				log.Printf("[handleAnonymousSubscribe] Error sending event: %v", err)
				return err
			}
		}
	}
}

// convertToAnonymousProtoEvent converts a Redis stream event to proto Event for anonymous topics
func (s *Service) convertToAnonymousProtoEvent(ctx context.Context, event redis.StreamEvent, topicID string) (*proto.Event, error) {
	// For anonymous topics, we don't have an application_id
	// The topicID is the channel ID itself
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
		AppId:         "", // Anonymous topics don't have an application
		TopicId:       topicID,
		EventType:     "webhook",
	}, nil
}

