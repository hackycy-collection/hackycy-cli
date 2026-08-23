package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	"golang.org/x/term"
)

type terminalForkAddPrompter struct {
	input     *bufio.Reader
	inputFile *os.File
	output    io.Writer
}

func newTerminalForkAddPrompter(input io.Reader, output io.Writer) *terminalForkAddPrompter {
	prompter := &terminalForkAddPrompter{input: bufio.NewReader(input), output: output}
	if file, ok := input.(*os.File); ok {
		prompter.inputFile = file
	}
	return prompter
}

func (prompter *terminalForkAddPrompter) Text(question configfork.TextPrompt) (string, bool) {
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

func (prompter *terminalForkAddPrompter) Select(question configfork.SelectPrompt) (string, bool) {
	_, _ = fmt.Fprintln(prompter.output, question.Message)
	for index, choice := range question.Choices {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, choice.Label)
	}
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		line, err := prompter.input.ReadString('\n')
		value := strings.TrimSpace(line)
		if err != nil && value == "" {
			return "", true
		}
		if value == "" && len(question.Choices) > 0 {
			return question.Choices[0].Value, false
		}
		index, parseErr := strconv.Atoi(value)
		if parseErr == nil && index >= 1 && index <= len(question.Choices) {
			return question.Choices[index-1].Value, false
		}
		_, _ = fmt.Fprintln(prompter.output, "Invalid selection")
		if err != nil {
			return "", true
		}
	}
}

func (prompter *terminalForkAddPrompter) Password(question configfork.TextPrompt) (string, bool) {
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

func (prompter *terminalForkAddPrompter) writeTextPrompt(question configfork.TextPrompt) {
	_, _ = fmt.Fprint(prompter.output, question.Message)
	if question.Placeholder != "" {
		_, _ = fmt.Fprintf(prompter.output, " (%s)", question.Placeholder)
	}
	_, _ = fmt.Fprint(prompter.output, ": ")
}

func (prompter *terminalForkAddPrompter) readLine() (string, bool) {
	line, err := prompter.input.ReadString('\n')
	value := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if err != nil && value == "" {
		return "", true
	}
	return value, false
}

func (prompter *terminalForkAddPrompter) readPassword() (string, bool) {
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

type terminalForkAddPresenter struct {
	output io.Writer
}

func (presenter terminalForkAddPresenter) Cancel(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalForkAddPresenter) Success(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}
