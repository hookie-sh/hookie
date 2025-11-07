package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hookie/ingest/internal/redis"
)

type WebhookHandler struct {
	redisClient *redis.Client
}

func NewWebhookHandler(redisClient *redis.Client) *WebhookHandler {
	return &WebhookHandler{
		redisClient: redisClient,
	}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	appId := r.PathValue("appId")
	topicId := r.PathValue("topicId")

	if appId == "" || topicId == "" {
		h.respondError(w, http.StatusBadRequest, "Invalid path parameters")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		h.respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

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

	streamKey := fmt.Sprintf("webhook:events:%s:%s", appId, topicId)

	fields := map[string]interface{}{
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
		"app_id":         appId,
		"topic_id":       topicId,
	}

	if err := h.redisClient.PublishWebhook(r.Context(), streamKey, fields); err != nil {
		log.Printf("Error publishing webhook to redis: %v", err)
		h.respondError(w, http.StatusInternalServerError, "Failed to process webhook")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *WebhookHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
