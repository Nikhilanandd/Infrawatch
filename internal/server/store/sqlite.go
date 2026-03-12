package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// DB wraps the SQLite database connection.
type DB struct {
	conn *sql.DB
}

// New opens or creates a SQLite database and runs migrations.
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			hostname TEXT NOT NULL,
			last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			status TEXT DEFAULT 'online'
		)`,
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			node_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			payload TEXT NOT NULL,
			FOREIGN KEY (node_id) REFERENCES nodes(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_node_ts ON metrics(node_id, timestamp DESC)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'viewer',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			metric TEXT NOT NULL,
			operator TEXT NOT NULL,
			threshold REAL NOT NULL,
			severity TEXT NOT NULL DEFAULT 'warning',
			enabled BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rule_id INTEGER,
			node_id TEXT NOT NULL,
			message TEXT NOT NULL,
			severity TEXT NOT NULL,
			value REAL,
			fired_at DATETIME NOT NULL,
			resolved BOOLEAN DEFAULT 0,
			FOREIGN KEY (rule_id) REFERENCES alert_rules(id),
			FOREIGN KEY (node_id) REFERENCES nodes(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_node ON alerts(node_id, fired_at DESC)`,
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", q[:40], err)
		}
	}
	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// --- Nodes ---

// UpsertNode inserts or updates a node.
func (db *DB) UpsertNode(id, hostname string) error {
	_, err := db.conn.Exec(
		`INSERT INTO nodes (id, hostname, last_seen_at, status) VALUES (?, ?, CURRENT_TIMESTAMP, 'online')
		 ON CONFLICT(id) DO UPDATE SET hostname = excluded.hostname, last_seen_at = CURRENT_TIMESTAMP, status = 'online'`,
		id, hostname,
	)
	return err
}

// GetNodes returns all nodes.
func (db *DB) GetNodes() ([]models.Node, error) {
	rows, err := db.conn.Query(`SELECT id, hostname, last_seen_at, status FROM nodes ORDER BY hostname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []models.Node
	for rows.Next() {
		var n models.Node
		if err := rows.Scan(&n.ID, &n.Hostname, &n.LastSeenAt, &n.Status); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

// --- Metrics ---

// InsertMetric stores a metric payload as JSON.
func (db *DB) InsertMetric(nodeID string, ts time.Time, payload *models.MetricPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}
	_, err = db.conn.Exec(
		`INSERT INTO metrics (node_id, timestamp, payload) VALUES (?, ?, ?)`,
		nodeID, ts.UTC(), string(data),
	)
	return err
}

// GetMetrics retrieves metrics for a node within a time window.
func (db *DB) GetMetrics(nodeID string, from, to time.Time, limit int) ([]models.MetricPayload, error) {
	rows, err := db.conn.Query(
		`SELECT payload FROM metrics WHERE node_id = ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp DESC LIMIT ?`,
		nodeID, from.UTC(), to.UTC(), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.MetricPayload
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var m models.MetricPayload
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// GetLatestMetric returns the most recent metric for a node.
func (db *DB) GetLatestMetric(nodeID string) (*models.MetricPayload, error) {
	var raw string
	err := db.conn.QueryRow(
		`SELECT payload FROM metrics WHERE node_id = ? ORDER BY timestamp DESC LIMIT 1`, nodeID,
	).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var m models.MetricPayload
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// --- Users ---

// CreateUser inserts a new user.
func (db *DB) CreateUser(username, passwordHash, role string) error {
	_, err := db.conn.Exec(
		`INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`,
		username, passwordHash, role,
	)
	return err
}

// GetUserByUsername fetches a user by username.
func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	var u models.User
	err := db.conn.QueryRow(
		`SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// --- Alert Rules ---

// UpsertAlertRule inserts or updates an alert rule by name.
func (db *DB) UpsertAlertRule(rule models.AlertRule) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO alert_rules (name, metric, operator, threshold, severity, enabled) VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(name) DO UPDATE SET metric = excluded.metric, operator = excluded.operator, threshold = excluded.threshold, severity = excluded.severity, enabled = excluded.enabled`,
		rule.Name, rule.Metric, rule.Operator, rule.Threshold, rule.Severity, rule.Enabled,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAlertRules returns all alert rules.
func (db *DB) GetAlertRules() ([]models.AlertRule, error) {
	rows, err := db.conn.Query(`SELECT id, name, metric, operator, threshold, severity, enabled FROM alert_rules`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.AlertRule
	for rows.Next() {
		var r models.AlertRule
		if err := rows.Scan(&r.ID, &r.Name, &r.Metric, &r.Operator, &r.Threshold, &r.Severity, &r.Enabled); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// --- Alerts ---

// InsertAlert inserts a fired alert.
func (db *DB) InsertAlert(a models.Alert) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO alerts (rule_id, node_id, message, severity, value, fired_at, resolved) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.RuleID, a.NodeID, a.Message, a.Severity, a.Value, a.FiredAt.UTC(), a.Resolved,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAlerts returns alerts, optionally filtered by node.
func (db *DB) GetAlerts(nodeID string, limit int) ([]models.Alert, error) {
	var rows *sql.Rows
	var err error
	if nodeID != "" {
		rows, err = db.conn.Query(
			`SELECT id, rule_id, node_id, message, severity, value, fired_at, resolved FROM alerts WHERE node_id = ? ORDER BY fired_at DESC LIMIT ?`,
			nodeID, limit,
		)
	} else {
		rows, err = db.conn.Query(
			`SELECT id, rule_id, node_id, message, severity, value, fired_at, resolved FROM alerts ORDER BY fired_at DESC LIMIT ?`,
			limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(&a.ID, &a.RuleID, &a.NodeID, &a.Message, &a.Severity, &a.Value, &a.FiredAt, &a.Resolved); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// PurgeOldMetrics deletes metrics older than the given retention period.
func (db *DB) PurgeOldMetrics(retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention).UTC()
	res, err := db.conn.Exec(`DELETE FROM metrics WHERE timestamp < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
