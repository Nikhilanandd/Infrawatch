package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/nikhilanandd/infrawatch/internal/config"
	"github.com/nikhilanandd/infrawatch/internal/server/auth"
	"github.com/nikhilanandd/infrawatch/internal/server/store"
	"github.com/nikhilanandd/infrawatch/internal/server/websocket"
	"github.com/nikhilanandd/infrawatch/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// Handler holds dependencies for API handlers.
type Handler struct {
	DB     *store.DB
	Hub    *websocket.Hub
	Config *config.ServerConfig
	Logger *slog.Logger
}

// JSON helper
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// HealthCheck handles GET /health.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "infrawatch-server",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.DB.GetUserByUsername(req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, expiresAt, err := auth.GenerateToken(user.Username, user.Role, h.Config.Auth.JWTSecret, h.Config.Auth.TokenExpiry)
	if err != nil {
		h.Logger.Error("failed to generate token", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Role:      user.Role,
	})
}

// IngestMetrics handles POST /api/v1/metrics (called by agents).
func (h *Handler) IngestMetrics(w http.ResponseWriter, r *http.Request) {
	var payload models.MetricPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Upsert node
	if err := h.DB.UpsertNode(payload.NodeID, payload.Hostname); err != nil {
		h.Logger.Error("upsert node failed", "error", err)
	}

	// Store metric
	if err := h.DB.InsertMetric(payload.NodeID, payload.Timestamp, &payload); err != nil {
		h.Logger.Error("insert metric failed", "error", err)
		writeError(w, http.StatusInternalServerError, "storage error")
		return
	}

	// Broadcast to WebSocket clients
	h.Hub.Broadcast(models.WSMessage{
		Type:    "metric",
		Payload: payload,
	})

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

// GetNodes handles GET /api/v1/nodes.
func (h *Handler) GetNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.DB.GetNodes()
	if err != nil {
		h.Logger.Error("get nodes failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if nodes == nil {
		nodes = []models.Node{}
	}
	writeJSON(w, http.StatusOK, nodes)
}

// GetMetrics handles GET /api/v1/metrics?node_id=X&from=T&to=T&limit=N.
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		writeError(w, http.StatusBadRequest, "node_id is required")
		return
	}

	from := time.Now().Add(-1 * time.Hour)
	to := time.Now()

	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}

	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	metrics, err := h.DB.GetMetrics(nodeID, from, to, limit)
	if err != nil {
		h.Logger.Error("get metrics failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if metrics == nil {
		metrics = []models.MetricPayload{}
	}
	writeJSON(w, http.StatusOK, metrics)
}

// GetAlerts handles GET /api/v1/alerts?node_id=X&limit=N.
func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	alerts, err := h.DB.GetAlerts(nodeID, limit)
	if err != nil {
		h.Logger.Error("get alerts failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if alerts == nil {
		alerts = []models.Alert{}
	}
	writeJSON(w, http.StatusOK, alerts)
}

// GetAlertRules handles GET /api/v1/alert-rules.
func (h *Handler) GetAlertRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.DB.GetAlertRules()
	if err != nil {
		h.Logger.Error("get alert rules failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if rules == nil {
		rules = []models.AlertRule{}
	}
	writeJSON(w, http.StatusOK, rules)
}
