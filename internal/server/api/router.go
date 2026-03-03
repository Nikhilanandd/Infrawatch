package api

import (
	"net/http"

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

	return corsMiddleware(mux)
}
