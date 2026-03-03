package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewAndClose(t *testing.T) {
	db := newTestDB(t)
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestUpsertAndGetNodes(t *testing.T) {
	db := newTestDB(t)

	if err := db.UpsertNode("node-1", "web-server-01"); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}
	if err := db.UpsertNode("node-2", "db-server-01"); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	nodes, err := db.GetNodes()
	if err != nil {
		t.Fatalf("GetNodes: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("len(nodes) = %d, want 2", len(nodes))
	}

	// Upsert same node with different hostname
	if err := db.UpsertNode("node-1", "web-server-01-updated"); err != nil {
		t.Fatalf("UpsertNode update: %v", err)
	}
	nodes, err = db.GetNodes()
	if err != nil {
		t.Fatalf("GetNodes after upsert: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("Upsert should not create duplicate, got %d nodes", len(nodes))
	}
}

func TestInsertAndGetMetrics(t *testing.T) {
	db := newTestDB(t)

	_ = db.UpsertNode("node-1", "host-1")

	now := time.Now()
	payload := &models.MetricPayload{
		NodeID:    "node-1",
		Timestamp: now,
		CPU: models.CPUStats{
			UsagePercent: 45.5,
			LoadAvg1:     1.2,
			LoadAvg5:     0.8,
			LoadAvg15:    0.5,
			CoreCount:    4,
		},
	}

	if err := db.InsertMetric("node-1", now, payload); err != nil {
		t.Fatalf("InsertMetric: %v", err)
	}

	metrics, err := db.GetMetrics("node-1", now.Add(-time.Hour), now.Add(time.Hour), 10)
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("len(metrics) = %d, want 1", len(metrics))
	}
	if metrics[0].CPU.UsagePercent != 45.5 {
		t.Errorf("CPU.UsagePercent = %f, want 45.5", metrics[0].CPU.UsagePercent)
	}
}

func TestGetLatestMetric(t *testing.T) {
	db := newTestDB(t)
	_ = db.UpsertNode("node-1", "host-1")

	now := time.Now()
	p1 := &models.MetricPayload{
		NodeID:    "node-1",
		Timestamp: now.Add(-time.Minute),
		CPU:       models.CPUStats{UsagePercent: 30},
	}
	p2 := &models.MetricPayload{
		NodeID:    "node-1",
		Timestamp: now,
		CPU:       models.CPUStats{UsagePercent: 60},
	}

	_ = db.InsertMetric("node-1", now.Add(-time.Minute), p1)
	_ = db.InsertMetric("node-1", now, p2)

	latest, err := db.GetLatestMetric("node-1")
	if err != nil {
		t.Fatalf("GetLatestMetric: %v", err)
	}
	if latest == nil {
		t.Fatal("latest should not be nil")
	}
	if latest.CPU.UsagePercent != 60 {
		t.Errorf("Expected latest CPU usage 60, got %f", latest.CPU.UsagePercent)
	}
}

func TestGetLatestMetricNoData(t *testing.T) {
	db := newTestDB(t)
	_ = db.UpsertNode("node-1", "host-1")

	latest, err := db.GetLatestMetric("node-1")
	// Implementation returns sql.ErrNoRows when no metrics exist
	if err == nil && latest == nil {
		// Acceptable: nil result, no error
	} else if err != nil && latest == nil {
		// Also acceptable: error on no rows
	} else if latest != nil {
		t.Errorf("Expected nil for node with no metrics, got %v", latest)
	}
}

func TestCreateAndGetUser(t *testing.T) {
	db := newTestDB(t)

	if err := db.CreateUser("admin", "$2a$10$fakehash", "admin"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	user, err := db.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("Username = %q, want %q", user.Username, "admin")
	}
	if user.Role != "admin" {
		t.Errorf("Role = %q, want %q", user.Role, "admin")
	}
	if user.PasswordHash != "$2a$10$fakehash" {
		t.Errorf("PasswordHash mismatch")
	}
}

func TestGetUserByUsernameNotFound(t *testing.T) {
	db := newTestDB(t)

	user, err := db.GetUserByUsername("nonexistent")
	if err == nil && user != nil {
		t.Error("Expected nil user or error for nonexistent user")
	}
}

func TestAlertRules(t *testing.T) {
	db := newTestDB(t)

	rule := models.AlertRule{
		Name:      "High CPU",
		Metric:    "cpu",
		Operator:  ">",
		Threshold: 80.0,
		Severity:  "warning",
		Enabled:   true,
	}

	id, err := db.UpsertAlertRule(rule)
	if err != nil {
		t.Fatalf("UpsertAlertRule: %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero rule ID")
	}

	rules, err := db.GetAlertRules()
	if err != nil {
		t.Fatalf("GetAlertRules: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(rules))
	}
	if rules[0].Name != "High CPU" {
		t.Errorf("Rule Name = %q, want %q", rules[0].Name, "High CPU")
	}
	if rules[0].Threshold != 80.0 {
		t.Errorf("Rule Threshold = %f, want 80.0", rules[0].Threshold)
	}
}

func TestInsertAndGetAlerts(t *testing.T) {
	db := newTestDB(t)
	_ = db.UpsertNode("node-1", "host-1")

	alert := models.Alert{
		RuleID:   1,
		NodeID:   "node-1",
		Message:  "CPU usage exceeded 80%",
		Severity: "warning",
		Value:    85.3,
		FiredAt:  time.Now(),
	}

	id, err := db.InsertAlert(alert)
	if err != nil {
		t.Fatalf("InsertAlert: %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero alert ID")
	}

	alerts, err := db.GetAlerts("node-1", 10)
	if err != nil {
		t.Fatalf("GetAlerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("len(alerts) = %d, want 1", len(alerts))
	}
	if alerts[0].Message != "CPU usage exceeded 80%" {
		t.Errorf("Alert Message = %q", alerts[0].Message)
	}
	if alerts[0].Value != 85.3 {
		t.Errorf("Alert Value = %f, want 85.3", alerts[0].Value)
	}
}

func TestPurgeOldMetrics(t *testing.T) {
	db := newTestDB(t)
	_ = db.UpsertNode("node-1", "host-1")

	old := time.Now().Add(-48 * time.Hour)
	recent := time.Now()

	pOld := &models.MetricPayload{
		NodeID:    "node-1",
		Timestamp: old,
		CPU:       models.CPUStats{UsagePercent: 10},
	}
	pRecent := &models.MetricPayload{
		NodeID:    "node-1",
		Timestamp: recent,
		CPU:       models.CPUStats{UsagePercent: 20},
	}

	_ = db.InsertMetric("node-1", old, pOld)
	_ = db.InsertMetric("node-1", recent, pRecent)

	purged, err := db.PurgeOldMetrics(24 * time.Hour)
	if err != nil {
		t.Fatalf("PurgeOldMetrics: %v", err)
	}
	if purged != 1 {
		t.Errorf("purged = %d, want 1", purged)
	}

	metrics, err := db.GetMetrics("node-1", old.Add(-time.Hour), recent.Add(time.Hour), 100)
	if err != nil {
		t.Fatalf("GetMetrics after purge: %v", err)
	}
	if len(metrics) != 1 {
		t.Errorf("Remaining metrics = %d, want 1", len(metrics))
	}
}

func TestMetricsLimit(t *testing.T) {
	db := newTestDB(t)
	_ = db.UpsertNode("node-1", "host-1")

	now := time.Now()
	for i := 0; i < 20; i++ {
		ts := now.Add(time.Duration(i) * time.Second)
		p := &models.MetricPayload{
			NodeID:    "node-1",
			Timestamp: ts,
			CPU:       models.CPUStats{UsagePercent: float64(i)},
		}
		_ = db.InsertMetric("node-1", ts, p)
	}

	metrics, err := db.GetMetrics("node-1", now.Add(-time.Hour), now.Add(time.Hour), 5)
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}
	if len(metrics) != 5 {
		t.Errorf("len(metrics) = %d, want 5 (limit)", len(metrics))
	}
}
