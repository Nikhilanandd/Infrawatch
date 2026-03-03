package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMetricPayloadRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	payload := MetricPayload{
		NodeID:    "node-001",
		Hostname:  "web-server-1",
		Timestamp: now,
		CPU: CPUStats{
			UsagePercent: 45.5,
			CoreCount:    4,
			PerCore:      []float64{40.0, 50.0, 45.0, 47.0},
			LoadAvg1:     1.2,
			LoadAvg5:     1.0,
			LoadAvg15:    0.8,
		},
		Memory: MemStats{
			TotalBytes:     8589934592,
			UsedBytes:      4294967296,
			AvailableBytes: 4294967296,
			UsagePercent:   50.0,
			SwapTotalBytes: 2147483648,
			SwapUsedBytes:  0,
		},
		Disk: DiskStats{
			Partitions: []PartitionStats{
				{Device: "/dev/sda1", Mountpoint: "/", Fstype: "ext4",
					TotalBytes: 107374182400, UsedBytes: 53687091200, FreeBytes: 53687091200, UsagePercent: 50.0},
			},
		},
		Network: NetStats{
			Interfaces: []InterfaceStats{
				{Name: "eth0", BytesSent: 1000000, BytesRecv: 2000000, PktsSent: 1000, PktsRecv: 2000},
			},
		},
		Uptime: 86400.5,
		Containers: []ContainerStats{
			{ID: "abc123", Name: "nginx", Image: "nginx:latest", State: "running", CPUPercent: 2.5, MemoryUsage: 52428800, MemoryLimit: 268435456},
		},
		Services: []ServiceStatus{
			{Name: "nginx", ActiveState: "active", SubState: "running", Description: "Web server"},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded MetricPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.NodeID != "node-001" {
		t.Errorf("NodeID = %q, want %q", decoded.NodeID, "node-001")
	}
	if decoded.CPU.UsagePercent != 45.5 {
		t.Errorf("CPU.UsagePercent = %f, want 45.5", decoded.CPU.UsagePercent)
	}
	if decoded.CPU.CoreCount != 4 {
		t.Errorf("CPU.CoreCount = %d, want 4", decoded.CPU.CoreCount)
	}
	if len(decoded.CPU.PerCore) != 4 {
		t.Errorf("CPU.PerCore len = %d, want 4", len(decoded.CPU.PerCore))
	}
	if decoded.Memory.TotalBytes != 8589934592 {
		t.Errorf("Memory.TotalBytes = %d, want 8589934592", decoded.Memory.TotalBytes)
	}
	if len(decoded.Disk.Partitions) != 1 || decoded.Disk.Partitions[0].UsagePercent != 50.0 {
		t.Error("Disk partition mismatch")
	}
	if len(decoded.Network.Interfaces) != 1 || decoded.Network.Interfaces[0].BytesRecv != 2000000 {
		t.Error("Network interface mismatch")
	}
	if decoded.Uptime != 86400.5 {
		t.Errorf("Uptime = %f, want 86400.5", decoded.Uptime)
	}
	if len(decoded.Containers) != 1 || decoded.Containers[0].State != "running" {
		t.Error("Container mismatch")
	}
	if len(decoded.Services) != 1 || decoded.Services[0].ActiveState != "active" {
		t.Error("Service mismatch")
	}
	if !decoded.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", decoded.Timestamp, now)
	}
}

func TestMetricPayloadOmitEmpty(t *testing.T) {
	payload := MetricPayload{
		NodeID:    "node-002",
		Hostname:  "test",
		Timestamp: time.Now().UTC(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	if _, ok := raw["containers"]; ok {
		t.Error("containers should be omitted when nil")
	}
	if _, ok := raw["services"]; ok {
		t.Error("services should be omitted when nil")
	}
}

func TestMetricPayloadFromJSON(t *testing.T) {
	jsonStr := `{
		"node_id": "test-node",
		"hostname": "host",
		"timestamp": "2026-03-03T10:00:00Z",
		"cpu": {"usage_percent": 75.5, "core_count": 8, "load_avg_1": 2.0, "load_avg_5": 1.5, "load_avg_15": 1.0},
		"memory": {"total_bytes": 17179869184, "used_bytes": 8589934592, "available_bytes": 8589934592, "usage_percent": 50.0},
		"disk": {"partitions": []},
		"network": {"interfaces": []},
		"uptime_seconds": 3600.0
	}`

	var p MetricPayload
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.NodeID != "test-node" {
		t.Errorf("NodeID = %q, want %q", p.NodeID, "test-node")
	}
	if p.CPU.UsagePercent != 75.5 {
		t.Errorf("CPU = %f, want 75.5", p.CPU.UsagePercent)
	}
	if p.Uptime != 3600.0 {
		t.Errorf("Uptime = %f, want 3600.0", p.Uptime)
	}
}

func TestTypesRoundTrip(t *testing.T) {
	alert := Alert{
		ID: 1, RuleID: 2, NodeID: "n1", Message: "test",
		Severity: "critical", Value: 92.5, FiredAt: time.Now().UTC().Truncate(time.Second), Resolved: false,
	}
	data, _ := json.Marshal(alert)
	var decoded Alert
	json.Unmarshal(data, &decoded)
	if decoded.NodeID != "n1" || decoded.Value != 92.5 {
		t.Error("Alert roundtrip failed")
	}

	node := Node{ID: "n1", Hostname: "h1", LastSeenAt: time.Now().UTC(), Status: "online"}
	data, _ = json.Marshal(node)
	var dn Node
	json.Unmarshal(data, &dn)
	if dn.ID != "n1" || dn.Status != "online" {
		t.Error("Node roundtrip failed")
	}

	wsMsg := WSMessage{Type: "metric", Payload: "test"}
	data, _ = json.Marshal(wsMsg)
	var dw WSMessage
	json.Unmarshal(data, &dw)
	if dw.Type != "metric" {
		t.Error("WSMessage roundtrip failed")
	}
}
