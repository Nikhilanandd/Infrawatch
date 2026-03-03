package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// CollectNetwork gathers network interface statistics from /proc/net/dev.
func CollectNetwork() (models.NetStats, error) {
	stats := models.NetStats{}

	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return stats, fmt.Errorf("open net/dev: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue // skip header lines
		}
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		if name == "lo" {
			continue // skip loopback
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}

		iface := models.InterfaceStats{Name: name}
		iface.BytesRecv, _ = strconv.ParseUint(fields[0], 10, 64)
		iface.PktsRecv, _ = strconv.ParseUint(fields[1], 10, 64)
		iface.ErrIn, _ = strconv.ParseUint(fields[2], 10, 64)
		iface.BytesSent, _ = strconv.ParseUint(fields[8], 10, 64)
		iface.PktsSent, _ = strconv.ParseUint(fields[9], 10, 64)
		iface.ErrOut, _ = strconv.ParseUint(fields[10], 10, 64)

		stats.Interfaces = append(stats.Interfaces, iface)
	}

	return stats, scanner.Err()
}
