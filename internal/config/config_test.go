package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAgentConfig(t *testing.T) {
	yaml := `
agent:
  node_id: "test-node-001"
  collect_interval: 10s
  log_level: "debug"

server:
  url: "https://monitor.example.com:8443/api/v1/metrics"
  ca_cert: "/etc/certs/ca.crt"
  client_cert: "/etc/certs/agent.crt"
  client_key: "/etc/certs/agent.key"

collectors:
  cpu: true
  memory: true
  disk: true
  network: true
  docker: false
  systemd: true
  uptime: true

services:
  - nginx
  - docker
  - sshd
`
	path := writeTempFile(t, "agent.yaml", yaml)

	cfg, err := LoadAgentConfig(path)
	if err != nil {
		t.Fatalf("LoadAgentConfig: %v", err)
	}
	if cfg.Agent.NodeID != "test-node-001" {
		t.Errorf("NodeID = %q, want %q", cfg.Agent.NodeID, "test-node-001")
	}
	if cfg.Agent.CollectInterval != 10*time.Second {
		t.Errorf("CollectInterval = %v, want 10s", cfg.Agent.CollectInterval)
	}
	if cfg.Agent.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.Agent.LogLevel, "debug")
	}
	if cfg.Server.URL != "https://monitor.example.com:8443/api/v1/metrics" {
		t.Errorf("URL = %q", cfg.Server.URL)
	}
	if !cfg.Collectors.CPU || !cfg.Collectors.Memory || !cfg.Collectors.Disk {
		t.Error("Expected CPU/Memory/Disk collectors enabled")
	}
	if cfg.Collectors.Docker {
		t.Error("Docker should be disabled")
	}
	if len(cfg.Services) != 3 {
		t.Errorf("Services len = %d, want 3", len(cfg.Services))
	}
}

func TestLoadAgentConfigDefaults(t *testing.T) {
	yaml := `
agent:
  node_id: "minimal"
`
	path := writeTempFile(t, "agent-min.yaml", yaml)

	cfg, err := LoadAgentConfig(path)
	if err != nil {
		t.Fatalf("LoadAgentConfig: %v", err)
	}
	if cfg.Agent.CollectInterval != 5*time.Second {
		t.Errorf("Default CollectInterval = %v, want 5s", cfg.Agent.CollectInterval)
	}
	if cfg.Agent.LogLevel != "info" {
		t.Errorf("Default LogLevel = %q, want %q", cfg.Agent.LogLevel, "info")
	}
}

func TestLoadServerConfig(t *testing.T) {
	yaml := `
server:
  host: "0.0.0.0"
  port: 9443
  log_level: "warn"
  tls:
    enabled: true
    ca_cert: "/certs/ca.crt"
    server_cert: "/certs/server.crt"
    server_key: "/certs/server.key"

database:
  path: "/tmp/test.db"

auth:
  jwt_secret: "super-secret"
  token_expiry: 12h

alerts:
  evaluation_interval: 30s
  rules:
    - name: "High CPU"
      metric: "cpu"
      operator: ">"
      threshold: 80.0
      severity: "warning"
      enabled: true

channels:
  email:
    enabled: true
    smtp_host: "smtp.example.com"
    smtp_port: 587
    username: "user"
    password: "pass"
    from: "alerts@example.com"
    to:
      - "ops@example.com"
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/XXX"
    channel: "#alerts"
`
	path := writeTempFile(t, "server.yaml", yaml)

	cfg, err := LoadServerConfig(path)
	if err != nil {
		t.Fatalf("LoadServerConfig: %v", err)
	}
	if cfg.Server.Port != 9443 {
		t.Errorf("Port = %d, want 9443", cfg.Server.Port)
	}
	if cfg.Server.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want %q", cfg.Server.LogLevel, "warn")
	}
	if !cfg.Server.TLS.Enabled {
		t.Error("TLS should be enabled")
	}
	if cfg.Database.Path != "/tmp/test.db" {
		t.Errorf("DB Path = %q", cfg.Database.Path)
	}
	if cfg.Auth.JWTSecret != "super-secret" {
		t.Errorf("JWTSecret = %q", cfg.Auth.JWTSecret)
	}
	if cfg.Auth.TokenExpiry != 12*time.Hour {
		t.Errorf("TokenExpiry = %v, want 12h", cfg.Auth.TokenExpiry)
	}
	if cfg.Alerts.EvaluationInterval != 30*time.Second {
		t.Errorf("EvaluationInterval = %v, want 30s", cfg.Alerts.EvaluationInterval)
	}
	if len(cfg.Alerts.Rules) != 1 {
		t.Fatalf("Rules len = %d, want 1", len(cfg.Alerts.Rules))
	}
	if cfg.Alerts.Rules[0].Threshold != 80.0 {
		t.Errorf("Rule threshold = %f, want 80.0", cfg.Alerts.Rules[0].Threshold)
	}
	if !cfg.Channels.Email.Enabled {
		t.Error("Email should be enabled")
	}
	if len(cfg.Channels.Email.To) != 1 {
		t.Errorf("Email To len = %d, want 1", len(cfg.Channels.Email.To))
	}
	if !cfg.Channels.Slack.Enabled {
		t.Error("Slack should be enabled")
	}
}

func TestLoadServerConfigDefaults(t *testing.T) {
	yaml := `
server:
  host: "0.0.0.0"
`
	path := writeTempFile(t, "server-min.yaml", yaml)

	cfg, err := LoadServerConfig(path)
	if err != nil {
		t.Fatalf("LoadServerConfig: %v", err)
	}
	if cfg.Server.Port != 8443 {
		t.Errorf("Default Port = %d, want 8443", cfg.Server.Port)
	}
	if cfg.Auth.TokenExpiry != 24*time.Hour {
		t.Errorf("Default TokenExpiry = %v, want 24h", cfg.Auth.TokenExpiry)
	}
	if cfg.Alerts.EvaluationInterval != 10*time.Second {
		t.Errorf("Default EvaluationInterval = %v, want 10s", cfg.Alerts.EvaluationInterval)
	}
	if cfg.Database.Path != "./infrawatch.db" {
		t.Errorf("Default DB Path = %q, want %q", cfg.Database.Path, "./infrawatch.db")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadAgentConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for missing agent config")
	}
	_, err = LoadServerConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for missing server config")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	path := writeTempFile(t, "bad.yaml", "{{invalid:yaml:::")
	_, err := LoadAgentConfig(path)
	if err == nil {
		t.Error("Expected error for invalid agent YAML")
	}
	_, err = LoadServerConfig(path)
	if err == nil {
		t.Error("Expected error for invalid server YAML")
	}
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return p
}
