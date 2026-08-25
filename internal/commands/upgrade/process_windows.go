//go:build windows

package upgrade

import "os/exec"

func configureDetachedCommand(command *exec.Cmd) {}
