package collector

import (
	"fmt"
	"syscall"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// CollectDisk gathers disk usage stats for common mount points.
func CollectDisk() (models.DiskStats, error) {
	stats := models.DiskStats{}

	mountPoints := []string{"/", "/home", "/var", "/tmp", "/opt"}

	for _, mp := range mountPoints {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mp, &stat); err != nil {
			continue // mount point may not exist
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := total - (stat.Bfree * uint64(stat.Bsize))

		if total == 0 {
			continue
		}

		p := models.PartitionStats{
			Device:       fmt.Sprintf("fs@%s", mp),
			Mountpoint:   mp,
			Fstype:       "unknown",
			TotalBytes:   total,
			UsedBytes:    used,
			FreeBytes:    free,
			UsagePercent: float64(used) / float64(total) * 100.0,
		}
		stats.Partitions = append(stats.Partitions, p)
	}

	if len(stats.Partitions) == 0 {
		return stats, fmt.Errorf("no partitions found")
	}
	return stats, nil
}
