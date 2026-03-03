package models

import "time"

// MetricPayload is the JSON payload sent by agents to the central server.
type MetricPayload struct {
	NodeID    string    `json:"node_id"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"`
	CPU       CPUStats  `json:"cpu"`
	Memory    MemStats  `json:"memory"`
	Disk      DiskStats `json:"disk"`
	Network   NetStats  `json:"network"`
	Uptime    float64   `json:"uptime_seconds"`
	Containers []ContainerStats `json:"containers,omitempty"`
	Services   []ServiceStatus  `json:"services,omitempty"`
}

// CPUStats contains CPU usage information.
type CPUStats struct {
	UsagePercent float64   `json:"usage_percent"`
	CoreCount    int       `json:"core_count"`
	PerCore      []float64 `json:"per_core,omitempty"`
	LoadAvg1     float64   `json:"load_avg_1"`
	LoadAvg5     float64   `json:"load_avg_5"`
	LoadAvg15    float64   `json:"load_avg_15"`
}

// MemStats contains memory usage information.
type MemStats struct {
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
	SwapTotalBytes uint64  `json:"swap_total_bytes"`
	SwapUsedBytes  uint64  `json:"swap_used_bytes"`
}

// DiskStats contains disk usage information.
type DiskStats struct {
	Partitions []PartitionStats `json:"partitions"`
}

// PartitionStats contains per-partition disk usage.
type PartitionStats struct {
	Device       string  `json:"device"`
	Mountpoint   string  `json:"mountpoint"`
	Fstype       string  `json:"fstype"`
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	FreeBytes    uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetStats contains network statistics.
type NetStats struct {
	Interfaces []InterfaceStats `json:"interfaces"`
}

// InterfaceStats contains per-interface network stats.
type InterfaceStats struct {
	Name      string `json:"name"`
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
	PktsSent  uint64 `json:"pkts_sent"`
	PktsRecv  uint64 `json:"pkts_recv"`
	ErrIn     uint64 `json:"err_in"`
	ErrOut    uint64 `json:"err_out"`
}

// ContainerStats contains Docker container information.
type ContainerStats struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	State       string  `json:"state"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryUsage uint64  `json:"memory_usage"`
	MemoryLimit uint64  `json:"memory_limit"`
}

// ServiceStatus contains systemd service status.
type ServiceStatus struct {
	Name        string `json:"name"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
	Description string `json:"description"`
}
