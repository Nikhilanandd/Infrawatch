package collector

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// CollectSystemd checks the status of the given systemd services.
func CollectSystemd(services []string) ([]models.ServiceStatus, error) {
	var result []models.ServiceStatus

	for _, svc := range services {
		ss := models.ServiceStatus{Name: svc}

		// Get ActiveState
		out, err := exec.Command("systemctl", "show", svc, "--property=ActiveState,SubState,Description", "--no-pager").Output()
		if err != nil {
			ss.ActiveState = "unknown"
			ss.SubState = "unknown"
			ss.Description = fmt.Sprintf("error querying %s: %v", svc, err)
			result = append(result, ss)
			continue
		}

		for _, line := range strings.Split(string(out), "\n") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			switch key {
			case "ActiveState":
				ss.ActiveState = val
			case "SubState":
				ss.SubState = val
			case "Description":
				ss.Description = val
			}
		}

		result = append(result, ss)
	}

	return result, nil
}
