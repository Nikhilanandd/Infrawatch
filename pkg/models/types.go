package models

import "time"

// User represents a system user.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // "admin" or "viewer"
	CreatedAt    time.Time `json:"created_at"`
}

// LoginRequest is the request body for authentication.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the response body after successful authentication.
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	Role      string `json:"role"`
}

// AlertRule defines when an alert should fire.
type AlertRule struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Metric    string  `json:"metric"`    // "cpu", "disk", "service"
	Operator  string  `json:"operator"`  // ">", "<", "=="
	Threshold float64 `json:"threshold"` // e.g. 80.0
	Severity  string  `json:"severity"`  // "warning", "critical"
	Enabled   bool    `json:"enabled"`
}

// Alert is a fired alert instance.
type Alert struct {
	ID        int64     `json:"id"`
	RuleID    int64     `json:"rule_id"`
	NodeID    string    `json:"node_id"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Value     float64   `json:"value"`
	FiredAt   time.Time `json:"fired_at"`
	Resolved  bool      `json:"resolved"`
}

// Node represents a monitored node.
type Node struct {
	ID         string    `json:"id"`
	Hostname   string    `json:"hostname"`
	LastSeenAt time.Time `json:"last_seen_at"`
	Status     string    `json:"status"` // "online", "offline"
}

// WSMessage is a WebSocket message for live updates.
type WSMessage struct {
	Type    string      `json:"type"` // "metric", "alert"
	Payload interface{} `json:"payload"`
}
