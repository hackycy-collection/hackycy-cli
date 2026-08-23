package main

import (
	"fmt"
	"io"
	"strings"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
)

type terminalCMRemovePrompter struct {
	*terminalCMAddPrompter
}

func newTerminalCMRemovePrompter(input io.Reader, output io.Writer) *terminalCMRemovePrompter {
	return &terminalCMRemovePrompter{terminalCMAddPrompter: newTerminalCMAddPrompter(input, output)}
}

func (prompter *terminalCMRemovePrompter) Confirm(question configcm.RemoveConfirmPrompt) (bool, bool) {
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
