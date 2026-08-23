package main

import (
	"context"

	heatcommand "github.com/hackycy/hackycy-cli/internal/commands/git/heat"
)

type osHeatGitRunner struct {
	executable string
}

func newOSHeatGitRunner() *osHeatGitRunner {
	return &osHeatGitRunner{executable: "git"}
}

func (runner *osHeatGitRunner) Run(ctx context.Context, arguments []string) (heatcommand.GitOutput, error) {
	output, err := runGitProcess(ctx, runner.executable, arguments)
	return heatcommand.GitOutput{
		Stdout:   output.stdout,
		Stderr:   output.stderr,
		ExitCode: output.exitCode,
	}, err
}

type heatGitSignalOutcome = gitSignalOutcome
