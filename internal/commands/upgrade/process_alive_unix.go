//go:build !windows

package upgrade

import (
	"errors"
	"os"
	"syscall"
)

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, os.ErrPermission)
}
