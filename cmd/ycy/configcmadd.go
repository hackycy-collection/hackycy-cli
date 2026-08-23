package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	"golang.org/x/term"
)

type terminalCMAddPrompter struct {
	input     *bufio.Reader
	inputFile *os.File
	output    io.Writer
}

func newTerminalCMAddPrompter(input io.Reader, output io.Writer) *terminalCMAddPrompter {
	prompter := &terminalCMAddPrompter{input: bufio.NewReader(input), output: output}
	if file, ok := input.(*os.File); ok {
		prompter.inputFile = file
	}
	return prompter
}

func (prompter *terminalCMAddPrompter) Text(question configcm.AddTextPrompt) (string, bool) {
	for {
		prompter.writeTextPrompt(question)
		value, cancelled := prompter.readLine()
		if cancelled {
			return "", true
		}
		if err := question.Validate(value); err != nil {
			_, _ = fmt.Fprintln(prompter.output, err)
			continue
		}
		return value, false
	}
}

func (prompter *terminalCMAddPrompter) Password(question configcm.AddTextPrompt) (string, bool) {
	for {
		prompter.writeTextPrompt(question)
		value, cancelled := prompter.readPassword()
		if cancelled {
			return "", true
		}
		if err := question.Validate(value); err != nil {
			_, _ = fmt.Fprintln(prompter.output, err)
			continue
		}
		return value, false
	}
}

func (prompter *terminalCMAddPrompter) writeTextPrompt(question configcm.AddTextPrompt) {
	_, _ = fmt.Fprint(prompter.output, question.Message)
	if question.Placeholder != "" {
		_, _ = fmt.Fprintf(prompter.output, " (%s)", question.Placeholder)
	}
	_, _ = fmt.Fprint(prompter.output, ": ")
}

func (prompter *terminalCMAddPrompter) readLine() (string, bool) {
	line, err := prompter.input.ReadString('\n')
	value := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if err != nil && value == "" {
		return "", true
	}
	return value, false
}

func (prompter *terminalCMAddPrompter) readPassword() (string, bool) {
	if prompter.inputFile != nil && term.IsTerminal(int(prompter.inputFile.Fd())) {
		value, err := term.ReadPassword(int(prompter.inputFile.Fd()))
		_, _ = fmt.Fprintln(prompter.output)
		if err != nil {
			return "", true
		}
		return string(value), false
	}
	return prompter.readLine()
}

type terminalCMAddPresenter struct {
	output io.Writer
}

func (presenter terminalCMAddPresenter) Cancel(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalCMAddPresenter) Success(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}
