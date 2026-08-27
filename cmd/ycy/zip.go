package main

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	zipcommand "github.com/hackycy/hackycy-cli/internal/commands/zip"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errZipRequiresInteractive = errors.New("zip requires an interactive terminal")

func newZipHandler(experience *terminalexperience.Runtime) cliapp.ZipHandler {
	return func(ctx context.Context, input zipcommand.Input) (zipcommand.Result, error) {
		run := experience.Open(ctx)
		defer run.Close()
		adapter := newTerminalZipAdapter(run, experience.Session())
		module, err := zipcommand.New(zipcommand.Dependencies{
			Prompter:           adapter,
			Presenter:          adapter,
			RemoteNameResolver: newZipRemoteNameResolver(osZipRemoteOutputRunner{}),
			Revealer:           newHostZipRevealer(osZipHostCommandRunner{}),
		})
		if err != nil {
			return zipcommand.Result{}, err
		}
		result, err := module.Run(input)
		if err != nil {
			return zipcommand.Result{}, err
		}
		if err := adapter.Flush(); err != nil {
			return zipcommand.Result{}, err
		}
		return result, nil
	}
}

type terminalZipAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
	pending []terminalexperience.PresentationDocument
}

func newTerminalZipAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalZipAdapter {
	return &terminalZipAdapter{run: run, session: session}
}

func (adapter *terminalZipAdapter) SelectPackage(step zipcommand.SelectPackageStep) (string, bool, error) {
	answer, cancelled, err := adapter.ask(zipChoiceRequest(step.Message, step.Options))
	return answer.Value, cancelled, err
}

func (adapter *terminalZipAdapter) SelectSource(step zipcommand.SelectSourceStep) (string, bool, error) {
	answer, cancelled, err := adapter.ask(zipChoiceRequest(step.Message, step.Options))
	return answer.Value, cancelled, err
}

func (adapter *terminalZipAdapter) SelectGlob(step zipcommand.SelectGlobStep) ([]string, bool, error) {
	options := zipInteractionOptions(step.Options)
	answer, cancelled, err := adapter.ask(terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionMultiSelect,
		Message:      step.Message,
		PlainLead:    step.Message,
		PlainPrompt:  "> ",
		Options:      options,
		HasDefault:   true,
		Default:      terminalexperience.InteractionAnswer{Values: append([]string(nil), step.InitialValues...)},
		CancelValues: []string{"q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			return parseZipGlob(value, options, step.InitialValues)
		},
	})
	return append([]string(nil), answer.Values...), cancelled, err
}

func (adapter *terminalZipAdapter) EditOutputFile(step zipcommand.EditOutputFileStep) (string, bool, error) {
	answer, cancelled, err := adapter.ask(terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionText,
		Message:      step.Message,
		Placeholder:  step.InitialValue,
		PlainPrompt:  step.Message + " [" + step.InitialValue + "]: ",
		HasDefault:   true,
		Default:      terminalexperience.InteractionAnswer{Value: step.InitialValue},
		CancelValues: []string{"q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			return parseZipOutput(value, step.InitialValue), nil
		},
	})
	return answer.Value, cancelled, err
}

func (adapter *terminalZipAdapter) Intro() {
	if adapter.session.Kind == terminalexperience.RichInteractive {
		adapter.present(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
			{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"},
			{Role: terminalexperience.VisualRoleActive, Text: "Zip Directory"},
		}})
		return
	}
	adapter.present(terminalZipDocument(adapter.session, "HACKYCY CLI\n\nZip Directory", terminalexperience.VisualRolePlain))
}

func (adapter *terminalZipAdapter) Note(note zipcommand.PlanningNote) {
	if adapter.session.Kind == terminalexperience.RichInteractive {
		adapter.present(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
			{Role: terminalexperience.VisualRoleActive, Text: note.Title},
			{Role: terminalexperience.VisualRoleMuted, Text: strings.Join(note.Lines, "\n")},
		}})
		return
	}
	lines := append([]string{note.Title}, note.Lines...)
	adapter.present(terminalZipDocument(adapter.session, strings.Join(lines, "\n"), terminalexperience.VisualRolePlain))
}

func (adapter *terminalZipAdapter) Progress(message string) {
	adapter.present(terminalZipDocument(adapter.session, message, terminalexperience.VisualRoleActive))
}

func (adapter *terminalZipAdapter) Cancel(message string) {
	adapter.present(terminalZipDocument(adapter.session, message, terminalexperience.VisualRoleWarning))
}

func (adapter *terminalZipAdapter) Outro(message string) {
	adapter.present(terminalZipDocument(adapter.session, message, terminalexperience.VisualRoleSuccess))
}

func (adapter *terminalZipAdapter) Flush() error {
	if adapter.session.Kind != terminalexperience.Automation {
		return nil
	}
	for _, document := range adapter.pending {
		if err := adapter.run.Present(document); err != nil {
			return err
		}
	}
	adapter.pending = nil
	return nil
}

func (adapter *terminalZipAdapter) ask(request terminalexperience.InteractionRequest) (terminalexperience.InteractionAnswer, bool, error) {
	answer, err := adapter.run.Ask(request)
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return terminalexperience.InteractionAnswer{}, true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return terminalexperience.InteractionAnswer{}, false, errZipRequiresInteractive
	}
	if err != nil {
		return terminalexperience.InteractionAnswer{}, false, err
	}
	return answer, false, nil
}

func (adapter *terminalZipAdapter) present(document terminalexperience.PresentationDocument) {
	if adapter.session.Kind == terminalexperience.Automation {
		adapter.pending = append(adapter.pending, document)
		return
	}
	_ = adapter.run.Present(document)
}

func zipChoiceRequest(message string, choices []zipcommand.PlanningChoice) terminalexperience.InteractionRequest {
	options := zipInteractionOptions(choices)
	request := terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionSelect,
		Message:      message,
		PlainLead:    message,
		PlainPrompt:  "> ",
		Options:      options,
		CancelValues: []string{"q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			return parseZipSelection(value, options)
		},
	}
	if len(options) > 0 {
		request.HasDefault = true
		request.Default = terminalexperience.InteractionAnswer{Value: options[0].Value}
	}
	return request
}

func zipInteractionOptions(choices []zipcommand.PlanningChoice) []terminalexperience.InteractionOption {
	options := make([]terminalexperience.InteractionOption, 0, len(choices))
	for _, choice := range choices {
		options = append(options, terminalexperience.InteractionOption{
			Label:       choice.Label,
			Value:       choice.Value,
			Description: choice.Hint,
		})
	}
	return options
}

func terminalZipDocument(session terminalexperience.Session, text string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		role = terminalexperience.VisualRolePlain
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func parseZipSelection(value string, options []terminalexperience.InteractionOption) (terminalexperience.InteractionAnswer, error) {
	index, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || index < 1 || index > len(options) {
		return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
	}
	return terminalexperience.InteractionAnswer{Value: options[index-1].Value}, nil
}

func parseZipGlob(value string, options []terminalexperience.InteractionOption, initialValues []string) (terminalexperience.InteractionAnswer, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "all") || strings.EqualFold(value, "none") {
		return terminalexperience.InteractionAnswer{Values: append([]string(nil), initialValues...)}, nil
	}
	indices, valid := parseZipIndices(value, len(options))
	if !valid {
		return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
	}
	values := make([]string, 0, len(indices))
	for _, index := range indices {
		values = append(values, options[index].Value)
	}
	return terminalexperience.InteractionAnswer{Values: values}, nil
}

func parseZipOutput(value, initialValue string) terminalexperience.InteractionAnswer {
	value = strings.TrimSpace(value)
	if value == "" {
		value = initialValue
	}
	return terminalexperience.InteractionAnswer{Value: value}
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
