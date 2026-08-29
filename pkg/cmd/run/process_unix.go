//go:build !windows

package run

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	runGroupSignalAttempts   = 3
	runGroupSignalRetryDelay = 10 * time.Millisecond
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
	var err error
	for attempt := 0; attempt < runGroupSignalAttempts; attempt++ {
		err = syscall.Kill(-process.Pid, signal)
		if err == nil || errors.Is(err, syscall.ESRCH) {
			return nil
		}
		if !errors.Is(err, syscall.EPERM) || attempt == runGroupSignalAttempts-1 {
			return err
		}
		time.Sleep(runGroupSignalRetryDelay)
	}
	return err
}

func runChildExitResult(exited *exec.ExitError) Result {
	if status, ok := exited.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return Result{ExitCode: 128 + int(status.Signal())}
	}
	return Result{ExitCode: exited.ExitCode()}
}
