//go:build !windows

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func configureGitChild(child *exec.Cmd) {
	child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func defaultGitTerminationSignal() os.Signal {
	return syscall.SIGTERM
}

func stopGitChild(process *os.Process, signal os.Signal) error {
	signalValue, ok := signal.(syscall.Signal)
	if !ok {
		signalValue = syscall.SIGTERM
	}
	return signalGitGroup(process, signalValue)
}

func killGitChild(process *os.Process) error {
	return signalGitGroup(process, syscall.SIGKILL)
}

func reapGitChild(process *os.Process) error {
	return signalGitGroup(process, syscall.SIGKILL)
}

func signalGitGroup(process *os.Process, signal syscall.Signal) error {
	err := syscall.Kill(-process.Pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func gitSignalExitCode(ctx context.Context) (int, bool) {
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
