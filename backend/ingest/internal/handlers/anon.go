package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hookie-sh/hookie/backend/ingest/internal/config"
	"github.com/hookie-sh/hookie/backend/ingest/internal/middleware"
	"github.com/hookie-sh/hookie/backend/ingest/internal/ratelimit"
	"github.com/hookie-sh/hookie/backend/ingest/internal/redis"
	"github.com/hookie-sh/hookie/backend/ingest/internal/supabase"
)

type AnonHandler struct {
	redisClient    *redis.Client
	supabaseClient *supabase.Client
}

func NewAnonHandler(redisClient *redis.Client, supabaseClient *supabase.Client) *AnonHandler {
	return &AnonHandler{
		redisClient:    redisClient,
		supabaseClient: supabaseClient,
	}
}

func (h *AnonHandler) HandleAnonWebhook(w http.ResponseWriter, r *http.Request) {
	anonTopicID := r.PathValue("anonTopicID")

	if anonTopicID == "" {
		h.respondError(w, http.StatusBadRequest, config.ErrInvalidPathParams)
		return
	}

	// Tier is already set in context by middleware (not used here but available)
	_ = middleware.TierFromContext(r.Context())

	// Body is already wrapped with MaxBytesReader by middleware
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		h.respondError(w, http.StatusBadRequest, config.ErrReadBody)
		return
	}
	defer r.Body.Close()

	// Check if any clients are connected to this topic
	hasClients, err := h.redisClient.HasConnectedClients(r.Context(), anonTopicID)
	if err != nil {
		log.Printf("Error checking connected clients: %v", err)
		// Continue anyway - don't block webhooks if check fails
	} else if !hasClients {
		// No clients connected, drop event
		log.Printf("[AnonHandler] Dropping event for anonymous topic %s - no connected clients", anonTopicID)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "dropped": "no clients connected"})
		return
	}

	// Build fields and publish
	fields := h.buildWebhookFields(r, body, anonTopicID)
	streamKey := config.BuildStreamKey(true, anonTopicID)

	if err := h.redisClient.PublishWebhook(r.Context(), streamKey, fields); err != nil {
		log.Printf("Error publishing webhook to redis: %v", err)
		h.respondError(w, http.StatusInternalServerError, config.ErrPublishWebhook)
		return
	}

	// Async tracking — fire and forget
	go h.supabaseClient.IncrementAnonTopicCount(context.Background(), anonTopicID)

	// Respond with rate limit headers (already set by middleware, but ensure they're there)
	result := middleware.ResultFromContext(r.Context())
	if result != nil {
		h.setRateLimitHeaders(w, result)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AnonHandler) buildWebhookFields(r *http.Request, body []byte, topicID string) map[string]interface{} {
	queryParams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}
	queryJSON, _ := json.Marshal(queryParams)

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = strings.Join(v, ", ")
		}
	}
	headersJSON, _ := json.Marshal(headers)

	remoteAddr := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		remoteAddr = strings.Split(xff, ",")[0]
		remoteAddr = strings.TrimSpace(remoteAddr)
	}

	contentType := r.Header.Get("Content-Type")
	contentLength := r.Header.Get("Content-Length")
	if contentLength == "" {
		contentLength = fmt.Sprintf("%d", len(body))
	}

	return map[string]interface{}{
		"method":         r.Method,
		"url":            r.URL.String(),
		"path":           r.URL.Path,
		"query":          string(queryJSON),
		"headers":        string(headersJSON),
		"body":           base64.StdEncoding.EncodeToString(body),
		"content_type":   contentType,
		"content_length": contentLength,
		"remote_addr":    remoteAddr,
		"timestamp":      time.Now().UnixNano(),
		"topic_id":       topicID,
	}
}

func (h *AnonHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *AnonHandler) setRateLimitHeaders(w http.ResponseWriter, result *ratelimit.RateLimitResult) {
	// Headers are already set by middleware, but this ensures they're present
	// This is a no-op in practice since middleware sets them
}
