package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

const gitTerminationGrace = 250 * time.Millisecond

type gitProcessOutput struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

type signalCause interface {
	Signal() os.Signal
}

// runGitProcess contains the external-Git invariants shared by the heat and pulse adapters.
func runGitProcess(ctx context.Context, executable string, arguments []string) (gitProcessOutput, error) {
	return runGitProcessInput(ctx, executable, arguments, nil)
}

// runGitProcessInput preserves the shared Git lifecycle while allowing the
// byte-oriented plumbing commands owned by Git CM to receive standard input.
func runGitProcessInput(ctx context.Context, executable string, arguments []string, input []byte) (gitProcessOutput, error) {
	if err := ctx.Err(); err != nil {
		return gitProcessOutput{}, err
	}

	child := exec.Command(executable, arguments...)
	if input != nil {
		child.Stdin = bytes.NewReader(input)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	child.Stdout = &stdout
	child.Stderr = &stderr
	configureGitChild(child)
	if err := child.Start(); err != nil {
		return gitProcessOutput{}, normalizeProcessStartError(err)
	}

	exited := make(chan error, 1)
	go func() {
		exited <- child.Wait()
	}()

	select {
	case err := <-exited:
		return completedGitProcess(stdout.Bytes(), stderr.Bytes(), err)
	case <-ctx.Done():
		if err := stopGitChild(child.Process, gitTerminationSignal(ctx)); err != nil {
			return gitProcessOutput{}, err
		}
		var waitErr error
		select {
		case waitErr = <-exited:
		case <-time.After(gitTerminationGrace):
			if err := killGitChild(child.Process); err != nil {
				return gitProcessOutput{}, err
			}
			waitErr = <-exited
		}
		if err := reapGitChild(child.Process); err != nil {
			return gitProcessOutput{}, err
		}
		_, _ = completedGitProcess(stdout.Bytes(), stderr.Bytes(), waitErr)
		if code, ok := gitSignalExitCode(ctx); ok {
			return gitProcessOutput{}, &gitSignalOutcome{code: code, cause: ctx.Err()}
		}
		return gitProcessOutput{}, ctx.Err()
	}
}

type gitSignalOutcome struct {
	code  int
	cause error
}

func (outcome *gitSignalOutcome) Error() string {
	return outcome.cause.Error()
}

func (outcome *gitSignalOutcome) Unwrap() error {
	return outcome.cause
}

func (outcome *gitSignalOutcome) ExitCode() int {
	return outcome.code
}

func completedGitProcess(stdout, stderr []byte, err error) (gitProcessOutput, error) {
	output := gitProcessOutput{stdout: stdout, stderr: stderr}
	if err == nil {
		return output, nil
	}
	var exited *exec.ExitError
	if errors.As(err, &exited) {
		output.exitCode = exited.ExitCode()
		return output, nil
	}
	return output, err
}

func gitTerminationSignal(ctx context.Context) os.Signal {
	if cause, ok := context.Cause(ctx).(signalCause); ok && cause.Signal() != nil {
		return cause.Signal()
	}
	return defaultGitTerminationSignal()
}
