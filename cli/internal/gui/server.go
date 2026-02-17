package gui

import (
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultUIPort = 4840

// DefaultPort returns the default GUI port (4840), or HOOKIE_UI_PORT if set.
func DefaultPort() int {
	if s := os.Getenv("HOOKIE_UI_PORT"); s != "" {
		if p, err := strconv.Atoi(s); err == nil && p > 0 {
			return p
		}
	}
	return defaultUIPort
}

// IsServerRunning returns true if a Hookie GUI server is running on the given port.
func IsServerRunning(port int) bool {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/health", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return false
	}
	return strings.Contains(string(body), `"ok"`)
}

// AcquireOrUseServer either starts a new GUI server on the given port or returns the URL
// of an existing one. Returns (guiURL, started, err). started is true if this process
// started the server.
func AcquireOrUseServer(port int) (*url.URL, bool, error) {
	if IsServerRunning(port) {
		u, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
		return u, false, nil
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err == nil {
		guiStorage := NewStorage(1000)
		Server(ln, guiStorage)
		u, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
		return u, true, nil
	}

	if !strings.Contains(err.Error(), "address already in use") {
		return nil, false, err
	}

	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		if IsServerRunning(port) {
			u, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
			return u, false, nil
		}
	}
	return nil, false, fmt.Errorf("port %d in use but GUI health check failed", port)
}

// Server starts the GUI HTTP server
func Server(ln net.Listener, storage *Storage) *http.Server {
	staticFS, _ := fs.Sub(Dist, "dist")
	handler := Handler(staticFS, storage)

	srv := &http.Server{
		Handler: handler,
	}
	go func() {
		_ = srv.Serve(ln)
	}()
	return srv
}
