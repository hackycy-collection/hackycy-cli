package main

import (
	"fmt"
	"io"
	"strings"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
)

type terminalForkRemovePrompter struct {
	*terminalForkAddPrompter
}

func newTerminalForkRemovePrompter(input io.Reader, output io.Writer) *terminalForkRemovePrompter {
	return &terminalForkRemovePrompter{terminalForkAddPrompter: newTerminalForkAddPrompter(input, output)}
}

func (prompter *terminalForkRemovePrompter) Confirm(question configfork.ConfirmPrompt) (bool, bool) {
	for {
		_, _ = fmt.Fprintf(prompter.output, "%s [y/N]: ", question.Message)
		line, err := prompter.input.ReadString('\n')
		value := strings.TrimSpace(line)
		if err != nil && value == "" {
			return false, true
		}
		switch strings.ToLower(value) {
		case "y", "yes":
			return true, false
		case "", "n", "no":
			return false, false
		default:
			_, _ = fmt.Fprintln(prompter.output, "Invalid confirmation")
			if err != nil {
				return false, true
			}
		}
	}
}

type terminalForkRemovePresenter struct {
	output io.Writer
}

func (presenter terminalForkRemovePresenter) Info(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalForkRemovePresenter) Outcome(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}
