package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"time"

	runcommand "github.com/hackycy/hackycy-cli/internal/commands/run"
)

type osRunChildRunner struct {
	input  io.Reader
	output io.Writer
	errors io.Writer
}

func newOSRunChildRunner(input io.Reader, output, errors io.Writer) *osRunChildRunner {
	return &osRunChildRunner{input: input, output: output, errors: errors}
}

const runTerminationGrace = 250 * time.Millisecond

func (runner *osRunChildRunner) Run(ctx context.Context, request runcommand.ChildRequest) (runcommand.Result, error) {
	if err := ctx.Err(); err != nil {
		return runcommand.Result{}, err
	}
	child := exec.Command(request.Executable, request.Arguments...)
	child.Dir = request.Directory
	child.Stdin = runner.input
	child.Stdout = runner.output
	child.Stderr = runner.errors
	configureRunChild(child)
	if err := child.Start(); err != nil {
		return runcommand.Result{}, err
	}
	exited := make(chan error, 1)
	go func() {
		exited <- child.Wait()
	}()

	select {
	case err := <-exited:
		return completedRunChild(err)
	case <-ctx.Done():
		if err := stopRunChild(child.Process, runTerminationSignal(ctx)); err != nil {
			return runcommand.Result{}, err
		}
		var err error
		select {
		case err = <-exited:
		case <-time.After(runTerminationGrace):
			if killErr := killRunChild(child.Process); killErr != nil {
				return runcommand.Result{}, killErr
			}
			err = <-exited
		}
		if reapErr := reapRunChild(child.Process); reapErr != nil {
			return runcommand.Result{}, reapErr
		}
		return completedRunChild(err)
	}
}

func completedRunChild(err error) (runcommand.Result, error) {
	if err == nil {
		return runcommand.Result{}, nil
	}
	var exited *exec.ExitError
	if errors.As(err, &exited) {
		return runChildExitResult(exited), nil
	}
	return runcommand.Result{}, err
}

type signalCause interface {
	Signal() os.Signal
}

func runTerminationSignal(ctx context.Context) os.Signal {
	if cause, ok := context.Cause(ctx).(signalCause); ok && cause.Signal() != nil {
		return cause.Signal()
	}
	return defaultRunTerminationSignal()
}
