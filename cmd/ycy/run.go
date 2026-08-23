package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	runcommand "github.com/hackycy/hackycy-cli/internal/commands/run"
)

func newRunModule(input io.Reader, output, errorOutput io.Writer) (*runcommand.Module, error) {
	return runcommand.New(runcommand.Dependencies{
		WorkingDirectory: os.Getwd,
		Reader:           osRunFileReader{},
		Exists:           osRunPathExists,
		Prompter:         newTerminalRunPrompter(input, output),
		Runner:           newOSRunChildRunner(input, output, errorOutput),
		Presenter:        terminalRunPresenter{output: output},
	})
}

type osRunFileReader struct{}

func (osRunFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func osRunPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

type terminalRunPrompter struct {
	input  *bufio.Reader
	output io.Writer
}

func newTerminalRunPrompter(input io.Reader, output io.Writer) *terminalRunPrompter {
	return &terminalRunPrompter{input: bufio.NewReader(input), output: output}
}

func (prompter *terminalRunPrompter) SelectScript(prompt runcommand.ScriptPrompt) (string, bool) {
	_, _ = fmt.Fprintln(prompter.output, prompt.Message)
	for index, option := range prompt.Options {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s - %s\n", index+1, option.Label, option.Hint)
	}
	index, cancelled := prompter.selectIndex(len(prompt.Options))
	if cancelled {
		return "", true
	}
	return prompt.Options[index].Value, false
}

func (prompter *terminalRunPrompter) SelectPackageManager(prompt runcommand.PackageManagerPrompt) (runcommand.PackageManager, bool) {
	_, _ = fmt.Fprintln(prompter.output, prompt.Message)
	for index, option := range prompt.Options {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, option.Label)
	}
	index, cancelled := prompter.selectIndex(len(prompt.Options))
	if cancelled {
		return "", true
	}
	return prompt.Options[index].Value, false
}

func (prompter *terminalRunPrompter) selectIndex(optionCount int) (int, bool) {
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		value, eof := prompter.readLine()
		if value == "" || isRunCancellation(value) {
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

func (prompter *terminalRunPrompter) readLine() (string, bool) {
	line, err := prompter.input.ReadString('\n')
	return strings.TrimSpace(line), err != nil
}

func isRunCancellation(value string) bool {
	return strings.EqualFold(value, "q") || strings.EqualFold(value, "quit") || strings.EqualFold(value, "cancel")
}

type terminalRunPresenter struct {
	output io.Writer
}

func (presenter terminalRunPresenter) Intro(message string) {
	_, _ = fmt.Fprintln(presenter.output, "HACKYCY CLI")
	_, _ = fmt.Fprintln(presenter.output)
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalRunPresenter) Info(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalRunPresenter) Blank() {
	_, _ = fmt.Fprintln(presenter.output)
}

func (presenter terminalRunPresenter) Cancel(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}
