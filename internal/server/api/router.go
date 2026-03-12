package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nikhilanandd/infrawatch/internal/server/auth"
)

// CORS middleware for frontend access.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter creates the HTTP router with all routes.
func (h *Handler) NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)

	// Agent metric ingestion (mTLS-protected, no JWT needed)
	mux.HandleFunc("POST /api/v1/metrics", h.IngestMetrics)

	// WebSocket (auth via query param)
	mux.HandleFunc("GET /api/v1/ws", h.Hub.HandleWebSocket)

	// Authenticated routes
	jwtMw := auth.Middleware(h.Config.Auth.JWTSecret)

	// Viewer + Admin routes
	mux.Handle("GET /api/v1/nodes", jwtMw(http.HandlerFunc(h.GetNodes)))
	mux.Handle("GET /api/v1/metrics", jwtMw(http.HandlerFunc(h.GetMetrics)))
	mux.Handle("GET /api/v1/alerts", jwtMw(http.HandlerFunc(h.GetAlerts)))
	mux.Handle("GET /api/v1/alert-rules", jwtMw(http.HandlerFunc(h.GetAlertRules)))

	// Admin-only routes
	adminMw := func(next http.Handler) http.Handler {
		return jwtMw(auth.RequireRole("admin")(next))
	}
	_ = adminMw // reserved for future admin-only endpoints

	// Serve React SPA from web/build if it exists
	if webDir := findWebBuildDir(); webDir != "" {
		spa := spaHandler{staticDir: webDir}
		mux.Handle("GET /", spa)
	}

	return corsMiddleware(mux)
}

// spaHandler serves static files and falls back to index.html for client-side routing.
type spaHandler struct {
	staticDir string
}

func (s spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only serve GET requests
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	// Clean the path
	p := filepath.Clean(r.URL.Path)
	if p == "/" {
		p = "/index.html"
	}

	// Don't serve API or WebSocket paths
	if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/health") {
		http.NotFound(w, r)
		return
	}

	// Check if the file exists on disk
	fullPath := filepath.Join(s.staticDir, p)
	if _, err := os.Stat(fullPath); err == nil {
		http.FileServer(http.Dir(s.staticDir)).ServeHTTP(w, r)
		return
	}

	// Fall back to index.html for SPA client-side routing
	http.ServeFile(w, r, filepath.Join(s.staticDir, "index.html"))
}

// findWebBuildDir looks for the React build directory in known locations.
func findWebBuildDir() string {
	candidates := []string{
		"web/build",     // development (run from project root)
		"./web/build",   // explicit relative
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	// Check if running from a different working directory
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidate := filepath.Join(dir, "web", "build")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}
