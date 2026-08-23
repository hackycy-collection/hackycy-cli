//go:build windows

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

func configureHeatGitChild(*exec.Cmd) {}

func defaultHeatGitTerminationSignal() os.Signal {
	return os.Interrupt
}

func stopHeatGitChild(process *os.Process, _ os.Signal) error {
	return ignoreCompletedHeatGitProcess(process.Kill())
}

func killHeatGitChild(process *os.Process) error {
	return ignoreCompletedHeatGitProcess(process.Kill())
}

func reapHeatGitChild(*os.Process) error {
	return nil
}

func ignoreCompletedHeatGitProcess(err error) error {
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	return err
}

func heatGitSignalExitCode(ctx context.Context) (int, bool) {
	if _, ok := context.Cause(ctx).(signalCause); ok {
		return 130, true
	}
	return 0, false
}
