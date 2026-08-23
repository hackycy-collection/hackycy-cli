//go:build windows

package main

import (
	"errors"
	"os"
	"os/exec"

	runcommand "github.com/hackycy/hackycy-cli/internal/commands/run"
)

func configureRunChild(*exec.Cmd) {}

func defaultRunTerminationSignal() os.Signal {
	return os.Interrupt
}

func stopRunChild(process *os.Process, _ os.Signal) error {
	return ignoreCompletedRunProcess(process.Kill())
}

func killRunChild(process *os.Process) error {
	return ignoreCompletedRunProcess(process.Kill())
}

func reapRunChild(*os.Process) error {
	return nil
}

func ignoreCompletedRunProcess(err error) error {
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	return err
}

func runChildExitResult(exited *exec.ExitError) runcommand.Result {
	code := exited.ExitCode()
	if code < 0 {
		code = 1
	}
	return runcommand.Result{ExitCode: code}
}
