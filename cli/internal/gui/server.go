package gui

import (
	"io/fs"
	"net"
	"net/http"
)

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
