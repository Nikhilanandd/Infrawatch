package collector

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// CollectCPU gathers CPU usage statistics from /proc/stat.
func CollectCPU() (models.CPUStats, error) {
	stats := models.CPUStats{
		CoreCount: runtime.NumCPU(),
	}

	// Read load averages
	loadData, err := os.ReadFile("/proc/loadavg")
	if err == nil {
		fields := strings.Fields(string(loadData))
		if len(fields) >= 3 {
			stats.LoadAvg1, _ = strconv.ParseFloat(fields[0], 64)
			stats.LoadAvg5, _ = strconv.ParseFloat(fields[1], 64)
			stats.LoadAvg15, _ = strconv.ParseFloat(fields[2], 64)
		}
	}

	// Compute CPU usage by sampling /proc/stat twice
	idle1, total1, perCoreIdle1, perCoreTotal1, err := readCPUStat()
	if err != nil {
		return stats, fmt.Errorf("read cpu stat: %w", err)
	}

	time.Sleep(250 * time.Millisecond)

	idle2, total2, perCoreIdle2, perCoreTotal2, err := readCPUStat()
	if err != nil {
		return stats, fmt.Errorf("read cpu stat: %w", err)
	}

	totalDelta := float64(total2 - total1)
	idleDelta := float64(idle2 - idle1)
	if totalDelta > 0 {
		stats.UsagePercent = (1.0 - idleDelta/totalDelta) * 100.0
	}

	// Per-core usage
	cores := min(len(perCoreIdle1), len(perCoreIdle2))
	stats.PerCore = make([]float64, cores)
	for i := 0; i < cores; i++ {
		td := float64(perCoreTotal2[i] - perCoreTotal1[i])
		id := float64(perCoreIdle2[i] - perCoreIdle1[i])
		if td > 0 {
			stats.PerCore[i] = (1.0 - id/td) * 100.0
		}
	}

	return stats, nil
}

func readCPUStat() (idle, total uint64, perCoreIdle, perCoreTotal []uint64, err error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, nil, nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			idle, total = parseCPULine(line)
		} else if strings.HasPrefix(line, "cpu") && len(line) > 3 && line[3] >= '0' && line[3] <= '9' {
			ci, ct := parseCPULine(line)
			perCoreIdle = append(perCoreIdle, ci)
			perCoreTotal = append(perCoreTotal, ct)
		}
	}
	return idle, total, perCoreIdle, perCoreTotal, scanner.Err()
}

func parseCPULine(line string) (idle, total uint64) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0, 0
	}
	var values []uint64
	for _, f := range fields[1:] {
		v, _ := strconv.ParseUint(f, 10, 64)
		values = append(values, v)
		total += v
	}
	if len(values) >= 4 {
		idle = values[3] // idle is the 4th field
	}
	return idle, total
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
