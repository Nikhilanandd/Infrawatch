package collector

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CollectUptime returns system uptime in seconds from /proc/uptime.
func CollectUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, fmt.Errorf("read uptime: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected uptime format")
	}
	uptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse uptime: %w", err)
	}
	return uptime, nil
}
