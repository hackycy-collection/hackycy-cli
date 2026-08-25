//go:build !windows

package upgrade

import (
	"os/exec"
	"syscall"
)

func configureDetachedCommand(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
