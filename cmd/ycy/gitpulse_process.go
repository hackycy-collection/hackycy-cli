package main

import (
	"context"

	pulsecommand "github.com/hackycy/hackycy-cli/internal/commands/git/pulse"
)

type osPulseGitRunner struct {
	executable string
}

func newOSPulseGitRunner() *osPulseGitRunner {
	return &osPulseGitRunner{executable: "git"}
}

func (runner *osPulseGitRunner) Run(ctx context.Context, arguments []string) (pulsecommand.GitOutput, error) {
	output, err := runGitProcess(ctx, runner.executable, arguments)
	return pulsecommand.GitOutput{
		Stdout:   output.stdout,
		Stderr:   output.stderr,
		ExitCode: output.exitCode,
	}, err
}

type pulseGitSignalOutcome = gitSignalOutcome
