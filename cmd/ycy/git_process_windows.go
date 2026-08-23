//go:build windows

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

func configureGitChild(*exec.Cmd) {}

func defaultGitTerminationSignal() os.Signal {
	return os.Interrupt
}

func stopGitChild(process *os.Process, _ os.Signal) error {
	return ignoreCompletedGitProcess(process.Kill())
}

func killGitChild(process *os.Process) error {
	return ignoreCompletedGitProcess(process.Kill())
}

func reapGitChild(*os.Process) error {
	return nil
}

func ignoreCompletedGitProcess(err error) error {
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	return err
}

func gitSignalExitCode(ctx context.Context) (int, bool) {
	if _, ok := context.Cause(ctx).(signalCause); ok {
		return 130, true
	}
	return 0, false
}
