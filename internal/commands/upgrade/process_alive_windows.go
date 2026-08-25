//go:build windows

package upgrade

import (
	"fmt"
	"os/exec"
	"strings"
)

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	output, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	if err != nil {
		return false
	}
	return !strings.Contains(strings.ToLower(string(output)), "no tasks")
}
