//go:build !windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"syscall"

	runcommand "github.com/hackycy/hackycy-cli/internal/commands/run"
)

func configureRunChild(child *exec.Cmd) {
	child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func defaultRunTerminationSignal() os.Signal {
	return syscall.SIGTERM
}

func stopRunChild(process *os.Process, signal os.Signal) error {
	signalValue, ok := signal.(syscall.Signal)
	if !ok {
		signalValue = syscall.SIGTERM
	}
	return signalRunChildGroup(process, signalValue)
}

func killRunChild(process *os.Process) error {
	return signalRunChildGroup(process, syscall.SIGKILL)
}

func reapRunChild(process *os.Process) error {
	return signalRunChildGroup(process, syscall.SIGKILL)
}

func signalRunChildGroup(process *os.Process, signal syscall.Signal) error {
	err := syscall.Kill(-process.Pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func runChildExitResult(exited *exec.ExitError) runcommand.Result {
	if status, ok := exited.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return runcommand.Result{ExitCode: 128 + int(status.Signal())}
	}
	return runcommand.Result{ExitCode: exited.ExitCode()}
}
