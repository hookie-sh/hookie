package gui

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

// Handler returns an http.Handler for the GUI server
func Handler(staticFS fs.FS, storage *Storage) http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("POST /api/ingest", func(w http.ResponseWriter, r *http.Request) {
		handleIngest(w, r, storage)
	})
	mux.HandleFunc("GET /api/events", func(w http.ResponseWriter, r *http.Request) {
		handleEvents(w, r, storage)
	})
	mux.HandleFunc("GET /api/stream", func(w http.ResponseWriter, r *http.Request) {
		handleStream(w, r, storage)
	})
	mux.HandleFunc("POST /api/events/clear", func(w http.ResponseWriter, r *http.Request) {
		storage.Clear()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	// Static files - SPA fallback: serve index.html for non-API, non-asset routes
	fileServer := http.FileServer(http.FS(staticFS))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || path == "index.html" {
			index, _ := fs.ReadFile(staticFS, "index.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(index)
			return
		}
		// Try to serve the file
		f, err := staticFS.Open(path)
		if err != nil {
			// SPA fallback: serve index.html for client-side routing
			index, _ := fs.ReadFile(staticFS, "index.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(index)
			return
		}
		_ = f.Close()
		fileServer.ServeHTTP(w, r)
	})

	return mux
}

func handleIngest(w http.ResponseWriter, r *http.Request, storage *Storage) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req IngestRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	event := storage.Add(req)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "id": event.ID})
}

func handleEvents(w http.ResponseWriter, r *http.Request, storage *Storage) {
	since := r.URL.Query().Get("since")
	events := storage.Events(since)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"events": events})
}

func handleStream(w http.ResponseWriter, r *http.Request, storage *Storage) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch, cancel := storage.Subscribe()
	defer cancel()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			_, _ = w.Write([]byte("event: event\n"))
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(data)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
