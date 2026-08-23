package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	pulsecommand "github.com/hackycy/hackycy-cli/internal/commands/git/pulse"
)

func newGitPulseModule(input io.Reader, output io.Writer) (*pulsecommand.Module, error) {
	return pulsecommand.New(pulsecommand.Dependencies{
		WorkingDirectory: os.Getwd,
		Stater:           osPulsePathStater{},
		Reader:           osPulseDirectoryReader{},
		Yield:            runtime.Gosched,
		Git:              newOSPulseGitRunner(),
		Prompter:         newTerminalPulsePrompter(input, output),
		Presenter:        terminalPulsePresenter{output: output},
		Now:              time.Now,
	})
}

type osPulsePathStater struct{}

func (osPulsePathStater) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

type osPulseDirectoryReader struct{}

func (osPulseDirectoryReader) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

type terminalPulsePrompter struct {
	input  *bufio.Reader
	output io.Writer
}

func newTerminalPulsePrompter(input io.Reader, output io.Writer) *terminalPulsePrompter {
	return &terminalPulsePrompter{input: bufio.NewReader(input), output: output}
}

func (prompter *terminalPulsePrompter) SelectDays(prompt pulsecommand.DayPrompt) (int, bool) {
	_, _ = fmt.Fprintln(prompter.output, prompt.Message)
	for index, option := range prompt.Options {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, option.Label)
	}
	index, cancelled := prompter.selectIndex(len(prompt.Options))
	if cancelled {
		return 0, true
	}
	return prompt.Options[index].Value, false
}

func (prompter *terminalPulsePrompter) SelectAuthors(prompt pulsecommand.AuthorPrompt) ([]string, bool) {
	_, _ = fmt.Fprintln(prompter.output, prompt.Message)
	for index, option := range prompt.Options {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, option.Label)
	}
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		value, eof := prompter.readLine()
		if isPulseCancellation(value) || (value == "" && eof) {
			return nil, true
		}
		if value == "" && len(prompt.InitialValues) > 0 {
			return append([]string(nil), prompt.InitialValues...), false
		}
		indices, valid := parsePulseIndices(value, len(prompt.Options))
		if valid && (!prompt.Required || len(indices) > 0) {
			selected := make([]string, 0, len(indices))
			for _, index := range indices {
				selected = append(selected, prompt.Options[index].Value)
			}
			return selected, false
		}
		if eof {
			return nil, true
		}
		if prompt.Required && value == "" {
			_, _ = fmt.Fprintln(prompter.output, "At least one author is required.")
		} else {
			_, _ = fmt.Fprintln(prompter.output, "Invalid selection")
		}
	}
}

func (prompter *terminalPulsePrompter) selectIndex(optionCount int) (int, bool) {
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		value, eof := prompter.readLine()
		if value == "" || isPulseCancellation(value) {
			return 0, true
		}
		index, err := strconv.Atoi(value)
		if err == nil && index >= 1 && index <= optionCount {
			return index - 1, false
		}
		if eof {
			return 0, true
		}
		_, _ = fmt.Fprintln(prompter.output, "Invalid selection")
	}
}

func (prompter *terminalPulsePrompter) readLine() (string, bool) {
	line, err := prompter.input.ReadString('\n')
	return strings.TrimSpace(line), err != nil
}

func isPulseCancellation(value string) bool {
	return strings.EqualFold(value, "q") || strings.EqualFold(value, "quit") || strings.EqualFold(value, "cancel")
}

func parsePulseIndices(value string, optionCount int) ([]int, bool) {
	parts := strings.FieldsFunc(value, func(character rune) bool {
		return character == ',' || character == ' ' || character == '\t'
	})
	if len(parts) == 0 {
		return nil, false
	}
	indices := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		index, err := strconv.Atoi(part)
		if err != nil || index < 1 || index > optionCount {
			return nil, false
		}
		index--
		if _, exists := seen[index]; exists {
			continue
		}
		seen[index] = struct{}{}
		indices = append(indices, index)
	}
	return indices, true
}

type terminalPulsePresenter struct {
	output io.Writer
}

func (presenter terminalPulsePresenter) Introduction(root string) {
	_, _ = fmt.Fprintln(presenter.output, "HACKYCY CLI")
	_, _ = fmt.Fprintln(presenter.output)
	_, _ = fmt.Fprintln(presenter.output, "Git Commit Tree")
	_, _ = fmt.Fprintf(presenter.output, "Workspace: %s\n", root)
}

func (presenter terminalPulsePresenter) ScanStarted() {
	_, _ = fmt.Fprintln(presenter.output, "Scanning repositories...")
}

func (presenter terminalPulsePresenter) RepositoryFound(root, repository string, count int) {
	_, _ = fmt.Fprintf(presenter.output, "Scanning repositories... [%d] %s\n", count, pulseRelativePath(root, repository))
}

func (presenter terminalPulsePresenter) RepositoriesFound(count int) {
	_, _ = fmt.Fprintf(presenter.output, "Found %d %s\n", count, pulsePlural(count, "repository", "repositories"))
}

func (presenter terminalPulsePresenter) NoRepositories() {
	_, _ = fmt.Fprintln(presenter.output, "No Git repositories found.")
}

func (presenter terminalPulsePresenter) FetchStarted(total int) {
	_, _ = fmt.Fprintf(presenter.output, "Fetching commits... [0/%d]\n", total)
}

func (presenter terminalPulsePresenter) FetchProgress(root, repository string, done, total int) {
	_, _ = fmt.Fprintf(presenter.output, "Fetching commits... [%d/%d] %s\n", done, total, pulseRelativePath(root, repository))
}

func (presenter terminalPulsePresenter) NoCommits() {
	_, _ = fmt.Fprintln(presenter.output, "No commits found in the specified date range.")
}

func (presenter terminalPulsePresenter) Cancelled() {
	_, _ = fmt.Fprintln(presenter.output, "Operation cancelled.")
}

func (presenter terminalPulsePresenter) Present(report pulsecommand.Report) {
	_, _ = fmt.Fprintln(presenter.output)
	_, _ = fmt.Fprintf(presenter.output, "Found %d %s in %d %s\n\n", report.CommitCount, pulsePlural(report.CommitCount, "commit", "commits"), len(report.Repositories), pulsePlural(len(report.Repositories), "repository", "repositories"))
	for groupIndex, repository := range report.Repositories {
		_, _ = fmt.Fprintf(presenter.output, "%s (%d %s)\n", filepath.Base(repository.Path), len(repository.Commits), pulsePlural(len(repository.Commits), "commit", "commits"))
		_, _ = fmt.Fprintf(presenter.output, "   %s%c\n", filepath.Dir(repository.Path), filepath.Separator)
		for commitIndex, commit := range repository.Commits {
			connector := "|-"
			if commitIndex == len(repository.Commits)-1 {
				connector = "`-"
			}
			_, _ = fmt.Fprintf(presenter.output, "   %s %s | %s | %s\n", connector, commit.Date, commit.Author, commit.Subject)
		}
		if groupIndex < len(report.Repositories)-1 {
			_, _ = fmt.Fprintln(presenter.output)
		}
	}
}

func pulseRelativePath(root, repository string) string {
	relative, err := filepath.Rel(root, repository)
	if err != nil || relative == "" {
		return "."
	}
	return relative
}

func pulsePlural(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
