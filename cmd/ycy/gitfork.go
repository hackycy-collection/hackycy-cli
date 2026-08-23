package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	forkcommand "github.com/hackycy/hackycy-cli/internal/commands/git/fork"
)

func newGitForkHandler(input io.Reader, output io.Writer) cliapp.GitForkHandler {
	return func(ctx context.Context, request forkcommand.Input) (forkcommand.Result, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return forkcommand.Result{}, err
		}
		provider, err := forkcommand.NewProviderClient(&http.Client{})
		if err != nil {
			return forkcommand.Result{}, err
		}
		module, err := forkcommand.New(forkcommand.Dependencies{
			Config:           store,
			WorkingDirectory: os.Getwd,
			Directories:      osForkDirectoryReader{},
			Prompter:         newTerminalGitForkPrompter(input, output),
			Provider:         provider,
			Extractor:        forkcommand.OSArchiveExtractor{},
			CloneRunner:      newOSForkGitRunner(),
			Remover:          osForkDirectoryRemover{},
			Presenter:        terminalGitForkPresenter{output: output},
		})
		if err != nil {
			return forkcommand.Result{}, err
		}
		return module.Run(ctx, request)
	}
}

type osForkDirectoryReader struct{}

func (osForkDirectoryReader) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

type osForkDirectoryRemover struct{}

func (osForkDirectoryRemover) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

type terminalGitForkPrompter struct {
	input  *bufio.Reader
	output io.Writer
}

func newTerminalGitForkPrompter(input io.Reader, output io.Writer) *terminalGitForkPrompter {
	return &terminalGitForkPrompter{input: bufio.NewReader(input), output: output}
}

func (prompter *terminalGitForkPrompter) ConfirmOverwrite(prompt forkcommand.OverwritePrompt) (bool, bool) {
	for {
		_, _ = fmt.Fprintf(prompter.output, "%s [Y/n]: ", prompt.Message)
		value, eof := prompter.readLine()
		switch strings.ToLower(value) {
		case "":
			return true, false
		case "y", "yes":
			return true, false
		case "n", "no":
			return false, false
		case "q", "quit", "cancel":
			return false, true
		}
		if eof {
			return false, true
		}
		_, _ = fmt.Fprintln(prompter.output, "Invalid confirmation")
	}
}

func (prompter *terminalGitForkPrompter) readLine() (string, bool) {
	line, err := prompter.input.ReadString('\n')
	return strings.TrimSpace(line), err != nil
}

type terminalGitForkPresenter struct {
	output io.Writer
}

func (presenter terminalGitForkPresenter) Introduction() {
	_, _ = fmt.Fprintln(presenter.output, "HACKYCY CLI")
	_, _ = fmt.Fprintln(presenter.output)
	_, _ = fmt.Fprintln(presenter.output, "Git Fork")
}

func (presenter terminalGitForkPresenter) Resolved(repository forkcommand.Repository) {
	_, _ = fmt.Fprintf(presenter.output, "Resolved: %s/%s/%s (%s)\n", repository.Host, repository.Owner, repository.Name, repository.ProviderType)
}

func (presenter terminalGitForkPresenter) DefaultBranchStarted() {
	_, _ = fmt.Fprintln(presenter.output, "Fetching default branch...")
}

func (presenter terminalGitForkPresenter) DefaultBranchResolved(ref string) {
	_, _ = fmt.Fprintf(presenter.output, "Branch: %s\n", ref)
}

func (presenter terminalGitForkPresenter) DefaultBranchFailed(err error) {
	_, _ = fmt.Fprintf(presenter.output, "Failed to get default branch: %s\n", err)
	_, _ = fmt.Fprintln(presenter.output, "Falling back to git clone with remote default branch.")
}

func (presenter terminalGitForkPresenter) ArchiveStarted() {
	_, _ = fmt.Fprintln(presenter.output, "Downloading archive...")
}

func (presenter terminalGitForkPresenter) ArchiveSucceeded() {
	_, _ = fmt.Fprintln(presenter.output, "Archive downloaded and extracted")
}

func (presenter terminalGitForkPresenter) ArchiveFailed(err error) {
	_, _ = fmt.Fprintf(presenter.output, "Archive download failed: %s\n", err)
}

func (presenter terminalGitForkPresenter) CloneStarted() {
	_, _ = fmt.Fprintln(presenter.output, "Falling back to git clone...")
}

func (presenter terminalGitForkPresenter) CloneSucceeded() {
	_, _ = fmt.Fprintln(presenter.output, "Cloned and cleaned up")
}

func (presenter terminalGitForkPresenter) Cancelled() {
	_, _ = fmt.Fprintln(presenter.output, "Cancelled")
}

func (presenter terminalGitForkPresenter) Completed(destination string) {
	_, _ = fmt.Fprintf(presenter.output, "Done! Project created at %s\n", destination)
}
