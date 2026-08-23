package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	zipcommand "github.com/hackycy/hackycy-cli/internal/commands/zip"
)

type terminalZipPrompter struct {
	input  *bufio.Reader
	output io.Writer
}

func newTerminalZipPrompter(input io.Reader, output io.Writer) *terminalZipPrompter {
	return &terminalZipPrompter{input: bufio.NewReader(input), output: output}
}

func newZipModule(input io.Reader, output io.Writer) (*zipcommand.Module, error) {
	presenter := terminalZipPresenter{output: output}
	return zipcommand.New(zipcommand.Dependencies{
		Prompter:           newTerminalZipPrompter(input, output),
		Presenter:          presenter,
		RemoteNameResolver: newZipRemoteNameResolver(osZipRemoteOutputRunner{}),
		Revealer:           newHostZipRevealer(osZipHostCommandRunner{}),
	})
}

func (prompter *terminalZipPrompter) SelectPackage(step zipcommand.SelectPackageStep) (string, bool) {
	index, cancelled := prompter.selectChoice(step.Message, step.Options)
	if cancelled {
		return "", true
	}
	return step.Options[index].Value, false
}

func (prompter *terminalZipPrompter) SelectSource(step zipcommand.SelectSourceStep) (string, bool) {
	index, cancelled := prompter.selectChoice(step.Message, step.Options)
	if cancelled {
		return "", true
	}
	return step.Options[index].Value, false
}

func (prompter *terminalZipPrompter) SelectGlob(step zipcommand.SelectGlobStep) ([]string, bool) {
	_, _ = fmt.Fprintln(prompter.output, step.Message)
	for index, option := range step.Options {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, option.Label)
	}
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		value, eof := prompter.readLine()
		if isZipCancellation(value) || (value == "" && eof) {
			return nil, true
		}
		if value == "" || strings.EqualFold(value, "all") || strings.EqualFold(value, "none") {
			return append([]string(nil), step.InitialValues...), false
		}
		indices, valid := parseZipIndices(value, len(step.Options))
		if valid {
			selected := make([]string, 0, len(indices))
			for _, index := range indices {
				selected = append(selected, step.Options[index].Value)
			}
			return selected, false
		}
		if eof {
			return nil, true
		}
		_, _ = fmt.Fprintln(prompter.output, "Invalid selection")
	}
}

func (prompter *terminalZipPrompter) EditOutputFile(step zipcommand.EditOutputFileStep) (string, bool) {
	for {
		_, _ = fmt.Fprintf(prompter.output, "%s [%s]: ", step.Message, step.InitialValue)
		value, eof := prompter.readLine()
		if isZipCancellation(value) || (value == "" && eof) {
			return "", true
		}
		if value == "" {
			return step.InitialValue, false
		}
		return value, false
	}
}

func (prompter *terminalZipPrompter) selectChoice(message string, options []zipcommand.PlanningChoice) (int, bool) {
	_, _ = fmt.Fprintln(prompter.output, message)
	for index, option := range options {
		if option.Hint == "" {
			_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, option.Label)
			continue
		}
		_, _ = fmt.Fprintf(prompter.output, "%d) %s - %s\n", index+1, option.Label, option.Hint)
	}
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		value, eof := prompter.readLine()
		if isZipCancellation(value) || (value == "" && eof) || len(options) == 0 {
			return 0, true
		}
		if value == "" {
			return 0, false
		}
		index, err := strconv.Atoi(value)
		if err == nil && index >= 1 && index <= len(options) {
			return index - 1, false
		}
		if eof {
			return 0, true
		}
		_, _ = fmt.Fprintln(prompter.output, "Invalid selection")
	}
}

func (prompter *terminalZipPrompter) readLine() (string, bool) {
	line, err := prompter.input.ReadString('\n')
	return strings.TrimSpace(line), err != nil
}

func isZipCancellation(value string) bool {
	return strings.EqualFold(value, "q") || strings.EqualFold(value, "quit") || strings.EqualFold(value, "cancel")
}

func parseZipIndices(value string, optionCount int) ([]int, bool) {
	parts := strings.FieldsFunc(value, func(character rune) bool {
		return character == ',' || character == ' ' || character == '\t'
	})
	if len(parts) == 0 {
		return nil, false
	}
	indices := make([]int, 0, len(parts))
	seen := make(map[int]bool, len(parts))
	for _, part := range parts {
		index, err := strconv.Atoi(part)
		if err != nil || index < 1 || index > optionCount || seen[index] {
			return nil, false
		}
		seen[index] = true
		indices = append(indices, index-1)
	}
	return indices, true
}

type zipRemoteOutputRunner interface {
	Output(string) ([]byte, error)
}

type osZipRemoteOutputRunner struct{}

func (osZipRemoteOutputRunner) Output(directory string) ([]byte, error) {
	command := exec.Command("git", "remote", "-v")
	command.Dir = directory
	return command.Output()
}

type zipRemoteNameResolver struct {
	runner zipRemoteOutputRunner
}

func newZipRemoteNameResolver(runner zipRemoteOutputRunner) zipcommand.RemoteNameResolver {
	return zipRemoteNameResolver{runner: runner}
}

func (resolver zipRemoteNameResolver) ResolveRemoteName(directory string) (string, error) {
	output, err := resolver.runner.Output(directory)
	if err != nil {
		return "", err
	}
	lines := make([]string, 0)
	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return "", nil
	}
	selected := lines[0]
	for _, line := range lines {
		if strings.HasPrefix(line, "origin ") || strings.HasPrefix(line, "origin\t") {
			selected = line
			break
		}
	}
	parts := strings.Fields(selected)
	if len(parts) < 2 {
		return "", nil
	}
	return zipcommand.ArchiveNameFromRemoteURL(parts[1]), nil
}

type zipHostCommandRunner interface {
	Run(string, []string) error
}

type osZipHostCommandRunner struct{}

func (osZipHostCommandRunner) Run(name string, arguments []string) error {
	return exec.Command(name, arguments...).Run()
}

type hostZipRevealer struct {
	runner zipHostCommandRunner
}

func newHostZipRevealer(runner zipHostCommandRunner) zipcommand.Revealer {
	return hostZipRevealer{runner: runner}
}

func (revealer hostZipRevealer) Reveal(path string) error {
	name, arguments, err := zipRevealCommand(runtime.GOOS, path)
	if err != nil {
		return err
	}
	return revealer.runner.Run(name, arguments)
}

func zipRevealCommand(goos, path string) (string, []string, error) {
	switch goos {
	case "darwin":
		return "open", []string{path}, nil
	case "linux":
		return "xdg-open", []string{path}, nil
	case "windows":
		return "cmd", []string{"/c", "start", "", path}, nil
	default:
		return "", nil, errors.New("archive reveal is not supported on this platform")
	}
}

type terminalZipPresenter struct {
	output io.Writer
}

func (presenter terminalZipPresenter) Intro() {
	_, _ = fmt.Fprintln(presenter.output, "HACKYCY CLI")
	_, _ = fmt.Fprintln(presenter.output)
	_, _ = fmt.Fprintln(presenter.output, "Zip Directory")
}

func (presenter terminalZipPresenter) Note(note zipcommand.PlanningNote) {
	_, _ = fmt.Fprintln(presenter.output, note.Title)
	for _, line := range note.Lines {
		_, _ = fmt.Fprintln(presenter.output, line)
	}
}

func (presenter terminalZipPresenter) Progress(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalZipPresenter) Cancel(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalZipPresenter) Outro(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}
