package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// dockerClient communicates with the Docker daemon via the Unix socket.
var dockerHTTPClient = &http.Client{
	Transport: &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")
		},
	},
	Timeout: 5 * time.Second,
}

type dockerContainer struct {
	ID    string   `json:"Id"`
	Names []string `json:"Names"`
	Image string   `json:"Image"`
	State string   `json:"State"`
}

type dockerStatsResponse struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
}

// CollectDocker gathers running Docker container stats.
func CollectDocker() ([]models.ContainerStats, error) {
	// List containers
	resp, err := dockerHTTPClient.Get("http://localhost/containers/json")
	if err != nil {
		return nil, fmt.Errorf("docker list: %w", err)
	}
	defer resp.Body.Close()

	var containers []dockerContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("docker decode: %w", err)
	}

	var result []models.ContainerStats
	for _, c := range containers {
		cs := models.ContainerStats{
			ID:    c.ID[:12],
			Image: c.Image,
			State: c.State,
		}
		if len(c.Names) > 0 {
			cs.Name = c.Names[0]
		}

		// Get stats (one-shot)
		statsResp, err := dockerHTTPClient.Get(fmt.Sprintf("http://localhost/containers/%s/stats?stream=false", c.ID))
		if err == nil {
			var ds dockerStatsResponse
			if json.NewDecoder(statsResp.Body).Decode(&ds) == nil {
				cpuDelta := float64(ds.CPUStats.CPUUsage.TotalUsage - ds.PreCPUStats.CPUUsage.TotalUsage)
				sysDelta := float64(ds.CPUStats.SystemCPUUsage - ds.PreCPUStats.SystemCPUUsage)
				if sysDelta > 0 {
					cs.CPUPercent = (cpuDelta / sysDelta) * 100.0
				}
				cs.MemoryUsage = ds.MemoryStats.Usage
				cs.MemoryLimit = ds.MemoryStats.Limit
			}
			statsResp.Body.Close()
		}

		result = append(result, cs)
	}

	return result, nil
}
