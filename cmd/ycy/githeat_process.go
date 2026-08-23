package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	heatcommand "github.com/hackycy/hackycy-cli/internal/commands/git/heat"
)

const heatGitTerminationGrace = 250 * time.Millisecond

type osHeatGitRunner struct {
	executable string
}

func newOSHeatGitRunner() *osHeatGitRunner {
	return &osHeatGitRunner{executable: "git"}
}

func (runner *osHeatGitRunner) Run(ctx context.Context, arguments []string) (heatcommand.GitOutput, error) {
	if err := ctx.Err(); err != nil {
		return heatcommand.GitOutput{}, err
	}

	child := exec.Command(runner.executable, arguments...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	child.Stdout = &stdout
	child.Stderr = &stderr
	configureHeatGitChild(child)
	if err := child.Start(); err != nil {
		return heatcommand.GitOutput{}, err
	}

	exited := make(chan error, 1)
	go func() {
		exited <- child.Wait()
	}()

	select {
	case err := <-exited:
		return completedHeatGitChild(stdout.Bytes(), stderr.Bytes(), err)
	case <-ctx.Done():
		if err := stopHeatGitChild(child.Process, heatGitTerminationSignal(ctx)); err != nil {
			return heatcommand.GitOutput{}, err
		}
		var waitErr error
		select {
		case waitErr = <-exited:
		case <-time.After(heatGitTerminationGrace):
			if err := killHeatGitChild(child.Process); err != nil {
				return heatcommand.GitOutput{}, err
			}
			waitErr = <-exited
		}
		if err := reapHeatGitChild(child.Process); err != nil {
			return heatcommand.GitOutput{}, err
		}
		_, _ = completedHeatGitChild(stdout.Bytes(), stderr.Bytes(), waitErr)
		if code, ok := heatGitSignalExitCode(ctx); ok {
			return heatcommand.GitOutput{}, &heatGitSignalOutcome{code: code, cause: ctx.Err()}
		}
		return heatcommand.GitOutput{}, ctx.Err()
	}
}

type heatGitSignalOutcome struct {
	code  int
	cause error
}

func (outcome *heatGitSignalOutcome) Error() string {
	return outcome.cause.Error()
}

func (outcome *heatGitSignalOutcome) Unwrap() error {
	return outcome.cause
}

func (outcome *heatGitSignalOutcome) ExitCode() int {
	return outcome.code
}

func completedHeatGitChild(stdout, stderr []byte, err error) (heatcommand.GitOutput, error) {
	output := heatcommand.GitOutput{Stdout: stdout, Stderr: stderr}
	if err == nil {
		return output, nil
	}
	var exited *exec.ExitError
	if errors.As(err, &exited) {
		output.ExitCode = exited.ExitCode()
		return output, nil
	}
	return output, err
}

func heatGitTerminationSignal(ctx context.Context) os.Signal {
	if cause, ok := context.Cause(ctx).(signalCause); ok && cause.Signal() != nil {
		return cause.Signal()
	}
	return defaultHeatGitTerminationSignal()
}
