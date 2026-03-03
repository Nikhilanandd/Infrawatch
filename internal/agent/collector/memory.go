package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// CollectMemory gathers memory usage stats from /proc/meminfo.
func CollectMemory() (models.MemStats, error) {
	stats := models.MemStats{}

	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return stats, fmt.Errorf("open meminfo: %w", err)
	}
	defer f.Close()

	info := make(map[string]uint64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(parts[1])
		valStr = strings.TrimSuffix(valStr, " kB")
		valStr = strings.TrimSpace(valStr)
		val, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			continue
		}
		info[key] = val * 1024 // convert kB to bytes
	}
	if err := scanner.Err(); err != nil {
		return stats, err
	}

	stats.TotalBytes = info["MemTotal"]
	stats.AvailableBytes = info["MemAvailable"]
	stats.UsedBytes = stats.TotalBytes - info["MemFree"] - info["Buffers"] - info["Cached"]
	stats.SwapTotalBytes = info["SwapTotal"]
	stats.SwapUsedBytes = info["SwapTotal"] - info["SwapFree"]

	if stats.TotalBytes > 0 {
		stats.UsagePercent = float64(stats.UsedBytes) / float64(stats.TotalBytes) * 100.0
	}

	return stats, nil
}
