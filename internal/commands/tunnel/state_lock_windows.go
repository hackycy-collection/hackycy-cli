//go:build windows

package tunnel

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func nativeStateLockProcessAlive(pid int) (bool, error) {
	if pid <= 0 {
		return false, nil
	}
	output, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	if err != nil {
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
