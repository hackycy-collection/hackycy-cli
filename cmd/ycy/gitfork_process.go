package main

import (
	"context"

	forkcommand "github.com/hackycy/hackycy-cli/internal/commands/git/fork"
)

type osForkGitRunner struct {
	executable string
}

func newOSForkGitRunner() *osForkGitRunner {
	return &osForkGitRunner{executable: "git"}
}

func (runner *osForkGitRunner) Run(ctx context.Context, arguments []string) (forkcommand.CloneOutput, error) {
	output, err := runGitProcess(ctx, runner.executable, arguments)
	return forkcommand.CloneOutput{Stderr: output.stderr, ExitCode: output.exitCode}, err
}

type forkGitSignalOutcome = gitSignalOutcome
