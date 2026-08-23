//go:build !windows

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func configureHeatGitChild(child *exec.Cmd) {
	child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func defaultHeatGitTerminationSignal() os.Signal {
	return syscall.SIGTERM
}

func stopHeatGitChild(process *os.Process, signal os.Signal) error {
	signalValue, ok := signal.(syscall.Signal)
	if !ok {
		signalValue = syscall.SIGTERM
	}
	return signalHeatGitGroup(process, signalValue)
}

func killHeatGitChild(process *os.Process) error {
	return signalHeatGitGroup(process, syscall.SIGKILL)
}

func reapHeatGitChild(process *os.Process) error {
	return signalHeatGitGroup(process, syscall.SIGKILL)
}

func signalHeatGitGroup(process *os.Process, signal syscall.Signal) error {
	err := syscall.Kill(-process.Pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func heatGitSignalExitCode(ctx context.Context) (int, bool) {
	cause, ok := context.Cause(ctx).(signalCause)
	if !ok {
		return 0, false
	}
	signal, ok := cause.Signal().(syscall.Signal)
	if !ok {
		return 0, false
	}
	return 128 + int(signal), true
}
