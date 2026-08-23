package main

import (
	"context"
	"io"
	"io/fs"
	"os"

	cmcommand "github.com/hackycy/hackycy-cli/internal/commands/git/cm"
)

type osCMGitRunner struct {
	executable string
}

func newOSCMGitRunner() *osCMGitRunner {
	return &osCMGitRunner{executable: "git"}
}

func (runner *osCMGitRunner) Run(ctx context.Context, arguments []string) (cmcommand.GitOutput, error) {
	output, err := runGitProcess(ctx, runner.executable, arguments)
	return cmGitOutput(output), err
}

func (runner *osCMGitRunner) RunInput(ctx context.Context, arguments []string, input []byte) (cmcommand.GitOutput, error) {
	output, err := runGitProcessInput(ctx, runner.executable, arguments, input)
	return cmGitOutput(output), err
}

func cmGitOutput(output gitProcessOutput) cmcommand.GitOutput {
	return cmcommand.GitOutput{
		Stdout:   output.stdout,
		Stderr:   output.stderr,
		ExitCode: output.exitCode,
	}
}

type osCMSnapshotFileSystem struct{}

func (osCMSnapshotFileSystem) Lstat(path string) (fs.FileInfo, error) {
	return os.Lstat(path)
}

func (osCMSnapshotFileSystem) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (osCMSnapshotFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
