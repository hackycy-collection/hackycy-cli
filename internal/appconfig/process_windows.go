//go:build windows

package appconfig

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func nativeProcessAlive(pid int) (bool, error) {
	if pid <= 0 {
		return false, nil
	}
	output, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	if err != nil {
		// A denied inspection is treated as alive, matching the legacy process.kill(pid, 0) rule.
		return true, nil
	}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		foundPID, parseErr := strconv.Atoi(fields[1])
		if parseErr == nil && foundPID == pid {
			return true, nil
		}
	}
	return false, nil
}
