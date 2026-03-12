package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikhilanandd/infrawatch/internal/config"
	"github.com/nikhilanandd/infrawatch/internal/server/auth"
	"github.com/nikhilanandd/infrawatch/internal/server/store"
	"github.com/nikhilanandd/infrawatch/internal/server/websocket"
	"github.com/nikhilanandd/infrawatch/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// testEnv sets up a full Handler with real DB and config for integration testing.
type testEnv struct {
	handler *Handler
	server  *httptest.Server
	db      *store.DB
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "integration-test.db")
	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Seed admin user
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err := db.CreateUser("admin", string(hash), "admin"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := websocket.NewHub(logger)
	go hub.Run()

	cfg := &config.ServerConfig{}
	cfg.Auth.JWTSecret = "test-integration-secret"
	cfg.Auth.TokenExpiry = time.Hour

	h := &Handler{
		DB:     db,
		Hub:    hub,
		Config: cfg,
		Logger: logger,
	}

	srv := httptest.NewServer(h.NewRouter())
	t.Cleanup(srv.Close)

	return &testEnv{handler: h, server: srv, db: db}
}

func (te *testEnv) getToken(t *testing.T) string {
	t.Helper()
	token, _, err := auth.GenerateToken("admin", "admin", te.handler.Config.Auth.JWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	return token
}

func TestHealthCheck(t *testing.T) {
	env := setupTestEnv(t)

	resp, err := http.Get(env.server.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
	if body["service"] != "infrawatch-server" {
		t.Errorf("service = %q, want %q", body["service"], "infrawatch-server")
	}
}

func TestLoginSuccess(t *testing.T) {
	env := setupTestEnv(t)

	body, _ := json.Marshal(models.LoginRequest{Username: "admin", Password: "admin"})
	resp, err := http.Post(env.server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/v1/auth/login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp models.LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)
	if loginResp.Token == "" {
		t.Error("Expected non-empty token")
	}
	if loginResp.Role != "admin" {
		t.Errorf("role = %q, want %q", loginResp.Role, "admin")
	}
	if loginResp.ExpiresAt == 0 {
		t.Error("Expected non-zero expiry")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	env := setupTestEnv(t)

	body, _ := json.Marshal(models.LoginRequest{Username: "admin", Password: "wrong"})
	resp, err := http.Post(env.server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/v1/auth/login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	env := setupTestEnv(t)

	body, _ := json.Marshal(models.LoginRequest{Username: "nobody", Password: "pass"})
	resp, err := http.Post(env.server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/v1/auth/login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestGetNodesUnauthorized(t *testing.T) {
	env := setupTestEnv(t)

	resp, err := http.Get(env.server.URL + "/api/v1/nodes")
	if err != nil {
		t.Fatalf("GET /api/v1/nodes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestGetNodesAuthorized(t *testing.T) {
	env := setupTestEnv(t)
	token := env.getToken(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/nodes", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/v1/nodes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var nodes []models.Node
	json.NewDecoder(resp.Body).Decode(&nodes)
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes initially, got %d", len(nodes))
	}
}

func TestIngestMetricsThenQuery(t *testing.T) {
	env := setupTestEnv(t)
	token := env.getToken(t)

	// Ingest a metric
	payload := models.MetricPayload{
		NodeID:    "test-node-01",
		Hostname:  "test-host",
		Timestamp: time.Now().UTC(),
		CPU:       models.CPUStats{UsagePercent: 42.5, CoreCount: 4},
		Memory:    models.MemStats{TotalBytes: 8 * 1024 * 1024 * 1024, UsedBytes: 4 * 1024 * 1024 * 1024, UsagePercent: 50.0},
		Uptime:    86400,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(env.server.URL+"/api/v1/metrics", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/v1/metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("ingest status = %d, body = %s", resp.StatusCode, string(bodyBytes))
	}

	// Query nodes — should now have 1
	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/nodes", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/v1/nodes: %v", err)
	}
	defer resp2.Body.Close()

	var nodes []models.Node
	json.NewDecoder(resp2.Body).Decode(&nodes)
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node after ingest, got %d", len(nodes))
	}
	if nodes[0].ID != "test-node-01" {
		t.Errorf("node ID = %q, want %q", nodes[0].ID, "test-node-01")
	}

	// Query metrics
	from := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	to := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)

	req3, _ := http.NewRequest("GET", env.server.URL+"/api/v1/metrics?node_id=test-node-01&from="+from+"&to="+to+"&limit=10", nil)
	req3.Header.Set("Authorization", "Bearer "+token)

	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("GET /api/v1/metrics: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		t.Errorf("metrics status = %d, want 200", resp3.StatusCode)
	}

	var metrics []models.MetricPayload
	json.NewDecoder(resp3.Body).Decode(&metrics)
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].CPU.UsagePercent != 42.5 {
		t.Errorf("CPU UsagePercent = %f, want 42.5", metrics[0].CPU.UsagePercent)
	}
}

func TestGetMetricsMissingNodeID(t *testing.T) {
	env := setupTestEnv(t)
	token := env.getToken(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/v1/metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestGetAlertsEmpty(t *testing.T) {
	env := setupTestEnv(t)
	token := env.getToken(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/alerts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/v1/alerts: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var alerts []models.Alert
	json.NewDecoder(resp.Body).Decode(&alerts)
	if len(alerts) != 0 {
		t.Errorf("Expected 0 alerts, got %d", len(alerts))
	}
}

func TestGetAlertRulesEmpty(t *testing.T) {
	env := setupTestEnv(t)
	token := env.getToken(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/alert-rules", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/v1/alert-rules: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var rules []models.AlertRule
	json.NewDecoder(resp.Body).Decode(&rules)
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules, got %d", len(rules))
	}
}

func TestInvalidTokenRejected(t *testing.T) {
	env := setupTestEnv(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/nodes", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/v1/nodes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestLoginThenUseToken(t *testing.T) {
	env := setupTestEnv(t)

	// Login to get real token
	loginBody, _ := json.Marshal(models.LoginRequest{Username: "admin", Password: "admin"})
	loginResp, err := http.Post(env.server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("POST login: %v", err)
	}
	defer loginResp.Body.Close()

	var lr models.LoginResponse
	json.NewDecoder(loginResp.Body).Decode(&lr)
	if lr.Token == "" {
		t.Fatal("Empty token from login")
	}

	// Use the token to access nodes
	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/nodes", nil)
	req.Header.Set("Authorization", "Bearer "+lr.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/v1/nodes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (using token from login)", resp.StatusCode)
	}
}
