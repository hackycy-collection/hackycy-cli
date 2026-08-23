package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/commands/exportenv"
)

func newExportEnvModule(input io.Reader, output io.Writer) (*exportenv.Module, error) {
	return exportenv.New(exportenv.Dependencies{
		WorkingDirectory: os.Getwd,
		Selector:         newTerminalExportEnvSelector(input, output),
		Reader:           osExportEnvReader{},
		Writer:           osExportEnvWriter{},
		Presenter:        terminalExportEnvPresenter{output: output},
	})
}

type terminalExportEnvSelector struct {
	input  *bufio.Reader
	output io.Writer
}

func newTerminalExportEnvSelector(input io.Reader, output io.Writer) *terminalExportEnvSelector {
	return &terminalExportEnvSelector{input: bufio.NewReader(input), output: output}
}

func (selector *terminalExportEnvSelector) SelectEnvironment(message string, choices []exportenv.EnvironmentChoice) (string, bool) {
	_, _ = fmt.Fprintln(selector.output, message)
	for index, choice := range choices {
		_, _ = fmt.Fprintf(selector.output, "%d) %s\n", index+1, choice.Label)
	}
	for {
		_, _ = fmt.Fprint(selector.output, "> ")
		line, err := selector.input.ReadString('\n')
		value := strings.TrimSpace(line)
		if value == "" || strings.EqualFold(value, "q") || strings.EqualFold(value, "quit") || strings.EqualFold(value, "cancel") {
			return "", true
		}
		index, parseErr := strconv.Atoi(value)
		if parseErr == nil && index >= 1 && index <= len(choices) {
			return choices[index-1].Value, false
		}
		if err != nil {
			return "", true
		}
		_, _ = fmt.Fprintln(selector.output, "Invalid selection")
	}
}

type osExportEnvReader struct{}

func (osExportEnvReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

type osExportEnvWriter struct{}

func (osExportEnvWriter) WriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0o666)
}

type terminalExportEnvPresenter struct {
	output io.Writer
}

func (presenter terminalExportEnvPresenter) Outro(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalExportEnvPresenter) Print(value string) {
	_, _ = fmt.Fprintln(presenter.output, value)
}

func (presenter terminalExportEnvPresenter) Cancel(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}
