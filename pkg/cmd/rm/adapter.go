package rm

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errRMRequiresInteractive = errors.New("rm requires an interactive terminal")

func runRM(options *Options) error {
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	adapter := newTerminalRMAdapter(run, options.Terminal.Session())
	module, err := New(Dependencies{
		WorkingDirectory: options.WorkingDirectory,
		Prompter:         adapter,
		Remover:          options.Remover,
		Presenter:        adapter,
	})
	if err != nil {
		return err
	}
	if _, err := module.Run(options.Context, Input{
		Paths: options.Paths,
		Force: options.Force,
		Depth: options.Depth,
	}); err != nil {
		return err
	}
	return adapter.Flush()
}

type terminalRMAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
	pending []terminalexperience.PresentationDocument
}

func newTerminalRMAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalRMAdapter {
	return &terminalRMAdapter{run: run, session: session}
}

func (adapter *terminalRMAdapter) ConfirmExplicit(prompt ExplicitConfirmationPrompt) (bool, bool, error) {
	answer, cancelled, err := adapter.ask(terminalexperience.InteractionRequest{
		Kind:        terminalexperience.InteractionConfirm,
		Message:     prompt.Message,
		HasDefault:  true,
		Default:     terminalexperience.InteractionAnswer{Confirmed: prompt.Initial},
		PlainPrompt: prompt.Message + " [y/N]: ",
		ParsePlain:  parseRMConfirmation,
	})
	return answer.Confirmed, cancelled, err
}

func (adapter *terminalRMAdapter) SelectSmartAction(prompt SmartActionPrompt) (SmartAction, bool, error) {
	request := terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionSelect,
		Message:      prompt.Message,
		PlainLead:    prompt.Message,
		PlainPrompt:  "> ",
		Options:      rmSmartActionOptions(prompt.Options),
		CancelValues: []string{"q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			return parseRMSmartAction(value, prompt.Options)
		},
	}
	if len(prompt.Options) > 0 {
		request.HasDefault = true
		request.Default = terminalexperience.InteractionAnswer{Value: prompt.Options[0].ID}
	}
	answer, cancelled, err := adapter.ask(request)
	if err != nil || cancelled {
		return SmartAction{}, cancelled, err
	}
	for _, option := range prompt.Options {
		if option.ID == answer.Value {
			return option, false, nil
		}
	}
	return SmartAction{}, false, errors.New("invalid selection")
}

func (adapter *terminalRMAdapter) SelectSmartTargets(prompt SmartTargetPrompt) ([]string, bool, error) {
	answer, cancelled, err := adapter.ask(terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionMultiSelect,
		Message:      prompt.Message,
		PlainLead:    prompt.Message,
		PlainPrompt:  "> ",
		Options:      rmSmartTargetOptions(prompt.Options),
		HasDefault:   true,
		Default:      terminalexperience.InteractionAnswer{Values: append([]string(nil), prompt.InitialValues...)},
		CancelValues: []string{"q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			return parseRMSmartTargets(value, prompt)
		},
	})
	if err != nil || cancelled {
		return nil, cancelled, err
	}
	return append([]string{}, answer.Values...), false, nil
}

func (adapter *terminalRMAdapter) Intro(message string) {
	if adapter.session.Kind == terminalexperience.RichInteractive {
		adapter.present(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
			{Role: terminalexperience.VisualRoleTitle, Text: "HACKYCY CLI"},
			{Role: terminalexperience.VisualRoleActive, Text: message},
		}})
		return
	}
	adapter.present(terminalRMDocument(adapter.session, "HACKYCY CLI\n\n"+message, terminalexperience.VisualRolePlain))
}

func (adapter *terminalRMAdapter) Paths(paths []string) {
	var text strings.Builder
	text.WriteByte('\n')
	for _, path := range paths {
		text.WriteString("  ")
		text.WriteString(path)
		text.WriteByte('\n')
	}
	text.WriteByte('\n')
	adapter.present(terminalRMDocument(adapter.session, text.String(), terminalexperience.VisualRoleMuted))
}

func (adapter *terminalRMAdapter) Notice(message string) {
	adapter.present(terminalRMDocument(adapter.session, message, terminalexperience.VisualRoleWarning))
}

func (adapter *terminalRMAdapter) ProgressStart(message string) {
	adapter.present(terminalRMDocument(adapter.session, message, terminalexperience.VisualRoleActive))
}

func (adapter *terminalRMAdapter) ProgressStop(message string) {
	adapter.present(terminalRMDocument(adapter.session, message, terminalexperience.VisualRoleMuted))
}

func (adapter *terminalRMAdapter) Cancel(message string) {
	adapter.present(terminalRMDocument(adapter.session, message, terminalexperience.VisualRoleWarning))
}

func (adapter *terminalRMAdapter) Outro(message string) {
	adapter.present(terminalRMDocument(adapter.session, message, terminalexperience.VisualRoleSuccess))
}

func (adapter *terminalRMAdapter) Flush() error {
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

func (adapter *terminalRMAdapter) ask(request terminalexperience.InteractionRequest) (terminalexperience.InteractionAnswer, bool, error) {
	answer, err := adapter.run.Ask(request)
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return terminalexperience.InteractionAnswer{}, true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return terminalexperience.InteractionAnswer{}, false, errRMRequiresInteractive
	}
	if err != nil {
		return terminalexperience.InteractionAnswer{}, false, err
	}
	return answer, false, nil
}

func (adapter *terminalRMAdapter) present(document terminalexperience.PresentationDocument) {
	if adapter.session.Kind == terminalexperience.Automation {
		adapter.pending = append(adapter.pending, document)
		return
	}
	_ = adapter.run.Present(document)
}

func terminalRMDocument(session terminalexperience.Session, text string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		role = terminalexperience.VisualRolePlain
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func rmSmartActionOptions(actions []SmartAction) []terminalexperience.InteractionOption {
	options := make([]terminalexperience.InteractionOption, 0, len(actions))
	for _, action := range actions {
		options = append(options, terminalexperience.InteractionOption{Label: action.Label, Value: action.ID})
	}
	return options
}

func rmSmartTargetOptions(targets []SmartTargetChoice) []terminalexperience.InteractionOption {
	options := make([]terminalexperience.InteractionOption, 0, len(targets))
	for _, target := range targets {
		options = append(options, terminalexperience.InteractionOption{Label: target.Label, Value: target.Value})
	}
	return options
}

func parseRMConfirmation(value string) (terminalexperience.InteractionAnswer, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes":
		return terminalexperience.InteractionAnswer{Confirmed: true}, nil
	case "n", "no":
		return terminalexperience.InteractionAnswer{Confirmed: false}, nil
	default:
		return terminalexperience.InteractionAnswer{}, errors.New("Invalid confirmation")
	}
}

func parseRMSmartAction(value string, actions []SmartAction) (terminalexperience.InteractionAnswer, error) {
	index, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || index < 1 || index > len(actions) {
		return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
	}
	return terminalexperience.InteractionAnswer{Value: actions[index-1].ID}, nil
}

func parseRMSmartTargets(value string, prompt SmartTargetPrompt) (terminalexperience.InteractionAnswer, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "all") {
		return terminalexperience.InteractionAnswer{Values: append([]string(nil), prompt.InitialValues...)}, nil
	}
	if strings.EqualFold(value, "none") {
		return terminalexperience.InteractionAnswer{Values: []string{}}, nil
	}
	selected := make([]string, 0, len(prompt.Options))
	seen := make(map[int]bool, len(prompt.Options))
	for _, part := range strings.Split(value, ",") {
		index, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || index < 1 || index > len(prompt.Options) || seen[index] {
			return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
		}
		seen[index] = true
		selected = append(selected, prompt.Options[index-1].Value)
	}
	return terminalexperience.InteractionAnswer{Values: selected}, nil
}

type osRMRemover struct{}

func (osRMRemover) RemovePath(path string) error {
	return os.RemoveAll(path)
}
