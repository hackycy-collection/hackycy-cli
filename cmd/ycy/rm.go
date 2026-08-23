package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	rmcommand "github.com/hackycy/hackycy-cli/internal/commands/rm"
)

func newRMModule(input io.Reader, output io.Writer) (*rmcommand.Module, error) {
	return rmcommand.New(rmcommand.Dependencies{
		WorkingDirectory: os.Getwd,
		Prompter:         newTerminalRMPrompter(input, output),
		Remover:          osRMRemover{},
		Presenter:        terminalRMPresenter{output: output},
	})
}

type terminalRMPrompter struct {
	input  *bufio.Reader
	output io.Writer
}

func newTerminalRMPrompter(input io.Reader, output io.Writer) *terminalRMPrompter {
	return &terminalRMPrompter{input: bufio.NewReader(input), output: output}
}

func (prompter *terminalRMPrompter) ConfirmExplicit(prompt rmcommand.ExplicitConfirmationPrompt) (bool, bool) {
	for {
		_, _ = fmt.Fprintf(prompter.output, "%s [y/N]: ", prompt.Message)
		value, eof := prompter.readLine()
		if eof && value == "" {
			return false, true
		}
		switch strings.ToLower(value) {
		case "y", "yes":
			return true, false
		case "", "n", "no":
			return false, false
		default:
			_, _ = fmt.Fprintln(prompter.output, "Invalid confirmation")
			if eof {
				return false, true
			}
		}
	}
}

func (prompter *terminalRMPrompter) SelectSmartAction(prompt rmcommand.SmartActionPrompt) (rmcommand.SmartAction, bool) {
	_, _ = fmt.Fprintln(prompter.output, prompt.Message)
	for index, option := range prompt.Options {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, option.Label)
	}
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		value, eof := prompter.readLine()
		if eof && value == "" {
			return rmcommand.SmartAction{}, true
		}
		if isRMCancellation(value) {
			return rmcommand.SmartAction{}, true
		}
		if value == "" && len(prompt.Options) > 0 {
			return prompt.Options[0], false
		}
		index, err := strconv.Atoi(value)
		if err == nil && index >= 1 && index <= len(prompt.Options) {
			return prompt.Options[index-1], false
		}
		if eof {
			return rmcommand.SmartAction{}, true
		}
		_, _ = fmt.Fprintln(prompter.output, "Invalid selection")
	}
}

func (prompter *terminalRMPrompter) SelectSmartTargets(prompt rmcommand.SmartTargetPrompt) ([]string, bool) {
	_, _ = fmt.Fprintln(prompter.output, prompt.Message)
	for index, option := range prompt.Options {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, option.Label)
	}
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		value, eof := prompter.readLine()
		if eof && value == "" {
			return nil, true
		}
		if isRMCancellation(value) {
			return nil, true
		}
		if value == "" || strings.EqualFold(value, "all") {
			return append([]string(nil), prompt.InitialValues...), false
		}
		if strings.EqualFold(value, "none") {
			return []string{}, false
		}
		selected, valid := selectRMTargetIndexes(value, prompt.Options)
		if valid {
			return selected, false
		}
		if eof {
			return nil, true
		}
		_, _ = fmt.Fprintln(prompter.output, "Invalid selection")
	}
}

func (prompter *terminalRMPrompter) readLine() (string, bool) {
	line, err := prompter.input.ReadString('\n')
	return strings.TrimSpace(line), err != nil
}

func selectRMTargetIndexes(value string, options []rmcommand.SmartTargetChoice) ([]string, bool) {
	selected := make([]string, 0, len(options))
	seen := make(map[int]bool, len(options))
	for _, part := range strings.Split(value, ",") {
		index, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || index < 1 || index > len(options) || seen[index] {
			return nil, false
		}
		seen[index] = true
		selected = append(selected, options[index-1].Value)
	}
	return selected, true
}

func isRMCancellation(value string) bool {
	return strings.EqualFold(value, "q") || strings.EqualFold(value, "quit") || strings.EqualFold(value, "cancel")
}

type osRMRemover struct{}

func (osRMRemover) RemovePath(path string) error {
	return os.RemoveAll(path)
}

type terminalRMPresenter struct {
	output io.Writer
}

func (presenter terminalRMPresenter) Intro(message string) {
	_, _ = fmt.Fprintln(presenter.output, "HACKYCY CLI")
	_, _ = fmt.Fprintln(presenter.output)
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalRMPresenter) Paths(paths []string) {
	_, _ = fmt.Fprintln(presenter.output)
	for _, path := range paths {
		_, _ = fmt.Fprintf(presenter.output, "  %s\n", path)
	}
	_, _ = fmt.Fprintln(presenter.output)
}

func (presenter terminalRMPresenter) Notice(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalRMPresenter) ProgressStart(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalRMPresenter) ProgressStop(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalRMPresenter) Cancel(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalRMPresenter) Outro(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}
