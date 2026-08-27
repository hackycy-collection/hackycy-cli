package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	cmcommand "github.com/hackycy/hackycy-cli/internal/commands/git/cm"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

var errGitCMRequiresInteractive = errors.New("git cm requires an interactive terminal")

func newGitCMHandler(experience *terminalexperience.Runtime) cliapp.GitCMHandler {
	return func(ctx context.Context, request cmcommand.Input) (cmcommand.Result, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		run := experience.Open(ctx)
		defer run.Close()
		adapter := newTerminalGitCMAdapter(run, experience.Session(), cancel)

		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return cmcommand.Result{}, err
		}
		module, err := cmcommand.New(cmcommand.Dependencies{
			Git:       newOSCMGitRunner(),
			Files:     osCMSnapshotFileSystem{},
			Prompter:  adapter,
			Committer: adapter,
			Resolver:  store,
			Transport: http.DefaultClient,
			Tracker:   adapter,
		})
		if err != nil {
			return cmcommand.Result{}, err
		}
		if experience.Session().Kind == terminalexperience.Automation {
			requiresInteraction, err := cmcommand.RequiresInteraction(request)
			if err != nil {
				return cmcommand.Result{}, err
			}
			if requiresInteraction {
				return cmcommand.Result{}, errGitCMRequiresInteractive
			}
		}

		result, err := module.Run(ctx, request)
		if err != nil {
			if result.Generated == nil && result.Profile != (cmcommand.ProfileDiagnostic{}) {
				if presentErr := adapter.PresentFailure(result); presentErr != nil {
					return result, presentErr
				}
			}
			if result.Committed && !result.Pushed {
				if presentErr := adapter.PresentOutcome(result); presentErr != nil {
					return result, presentErr
				}
			}
			return result, err
		}
		if result.Generated != nil && !result.PromptedCommit {
			if err := adapter.PresentGenerated(result); err != nil {
				return result, err
			}
		}
		if err := adapter.PresentOutcome(result); err != nil {
			return result, err
		}
		if err := adapter.Flush(); err != nil {
			return result, err
		}
		return result, nil
	}
}

type terminalGitCMAdapter struct {
	run           terminalexperience.ExperienceRun
	session       terminalexperience.Session
	requestCancel context.CancelFunc
	pending       []terminalexperience.PresentationDocument
}

func newTerminalGitCMAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session, requestCancel context.CancelFunc) *terminalGitCMAdapter {
	return &terminalGitCMAdapter{run: run, session: session, requestCancel: requestCancel}
}

func (adapter *terminalGitCMAdapter) SelectFiles(prompt cmcommand.StagePrompt) ([]string, bool, error) {
	answer, cancelled, err := adapter.ask(gitCMStageRequest(prompt))
	if err != nil || cancelled {
		return nil, cancelled, err
	}
	return append([]string(nil), answer.Values...), false, nil
}

func (adapter *terminalGitCMAdapter) ConfirmCommit(prompt cmcommand.CommitPrompt) (bool, bool, error) {
	if adapter.session.Kind == terminalexperience.PlainInteractive {
		if err := adapter.present(gitCMGeneratedDocument(adapter.session, prompt.Generated, prompt.Profile)); err != nil {
			return false, false, err
		}
	}
	answer, cancelled, err := adapter.ask(gitCMCommitRequest(prompt, adapter.session.Kind == terminalexperience.RichInteractive))
	if err != nil || cancelled {
		return false, cancelled, err
	}
	return answer.Confirmed, false, nil
}

func (adapter *terminalGitCMAdapter) Start(_ context.Context) (cmcommand.PhaseReporter, error) {
	reporter := &terminalGitCMPhaseReporter{
		updates:  make(chan terminalexperience.OperationPhase, 1),
		finished: make(chan struct{}),
	}
	go func() {
		err := adapter.run.Track(terminalexperience.TrackedOperation{
			Label:         "Git CM",
			Updates:       reporter.updates,
			RequestCancel: adapter.requestCancel,
		})
		reporter.complete(err)
	}()
	return reporter, nil
}

func (adapter *terminalGitCMAdapter) PresentGenerated(result cmcommand.Result) error {
	if result.Generated == nil {
		return nil
	}
	return adapter.present(gitCMGeneratedDocument(adapter.session, *result.Generated, result.Profile))
}

func (adapter *terminalGitCMAdapter) PresentFailure(result cmcommand.Result) error {
	return adapter.run.Present(gitCMFailureDocument(adapter.session, result.Profile))
}

func (adapter *terminalGitCMAdapter) PresentOutcome(result cmcommand.Result) error {
	document := gitCMOutcomeDocument(adapter.session, result)
	if len(document.Blocks) == 0 {
		return nil
	}
	return adapter.present(document)
}

func (adapter *terminalGitCMAdapter) Flush() error {
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

func (adapter *terminalGitCMAdapter) ask(request terminalexperience.InteractionRequest) (terminalexperience.InteractionAnswer, bool, error) {
	answer, err := adapter.run.Ask(request)
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return terminalexperience.InteractionAnswer{}, true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return terminalexperience.InteractionAnswer{}, false, errGitCMRequiresInteractive
	}
	if err != nil {
		return terminalexperience.InteractionAnswer{}, false, err
	}
	return answer, false, nil
}

func (adapter *terminalGitCMAdapter) present(document terminalexperience.PresentationDocument) error {
	if adapter.session.Kind == terminalexperience.Automation {
		adapter.pending = append(adapter.pending, document)
		return nil
	}
	return adapter.run.Present(document)
}

func gitCMStageRequest(prompt cmcommand.StagePrompt) terminalexperience.InteractionRequest {
	options := make([]terminalexperience.InteractionOption, 0, len(prompt.Options))
	for _, option := range prompt.Options {
		options = append(options, terminalexperience.InteractionOption{Label: option.Label, Value: option.Value})
	}
	return terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionMultiSelect,
		Message:      prompt.Message,
		PlainLead:    prompt.Message,
		PlainPrompt:  "> ",
		Options:      options,
		HasDefault:   true,
		Default:      terminalexperience.InteractionAnswer{Values: append([]string(nil), prompt.InitialValues...)},
		CancelValues: []string{"q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			value = strings.TrimSpace(value)
			if strings.EqualFold(value, "all") {
				return terminalexperience.InteractionAnswer{Values: append([]string(nil), prompt.InitialValues...)}, nil
			}
			if strings.EqualFold(value, "none") {
				return terminalexperience.InteractionAnswer{Values: []string{}}, nil
			}
			selected, valid := selectGitCMOptions(value, prompt.Options)
			if !valid {
				return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
			}
			return terminalexperience.InteractionAnswer{Values: selected}, nil
		},
	}
}

func gitCMCommitRequest(prompt cmcommand.CommitPrompt, richDescription bool) terminalexperience.InteractionRequest {
	description := ""
	if richDescription {
		description = strings.TrimSuffix(gitCMGeneratedText(prompt.Generated, prompt.Profile), "\n")
	}
	return terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionConfirm,
		Message:      prompt.Message,
		Description:  description,
		HasDefault:   true,
		Default:      terminalexperience.InteractionAnswer{Confirmed: true},
		CancelValues: []string{"q", "quit", "cancel"},
		PlainPrompt:  prompt.Message + " [Y/n]: ",
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			switch strings.ToLower(strings.TrimSpace(value)) {
			case "y", "yes":
				return terminalexperience.InteractionAnswer{Confirmed: true}, nil
			case "n", "no":
				return terminalexperience.InteractionAnswer{Confirmed: false}, nil
			default:
				return terminalexperience.InteractionAnswer{}, errors.New("Invalid confirmation")
			}
		},
	}
}

func selectGitCMOptions(value string, options []cmcommand.StageOption) ([]string, bool) {
	parts := strings.FieldsFunc(value, func(character rune) bool {
		return character == ',' || character == ' ' || character == '\t'
	})
	if len(parts) == 0 {
		return nil, false
	}
	selected := make([]string, 0, len(parts))
	seen := make(map[int]bool, len(parts))
	for _, part := range parts {
		index, err := strconv.Atoi(part)
		if err != nil || index < 1 || index > len(options) || seen[index] {
			return nil, false
		}
		seen[index] = true
		selected = append(selected, options[index-1].Value)
	}
	return selected, true
}

type terminalGitCMPhaseReporter struct {
	updates  chan terminalexperience.OperationPhase
	finished chan struct{}

	mu     sync.Mutex
	closed bool
	err    error
}

func (reporter *terminalGitCMPhaseReporter) Report(phase cmcommand.Phase) {
	update := terminalGitCMPhase(phase)
	select {
	case reporter.updates <- update:
	case <-reporter.finished:
	}
}

func (reporter *terminalGitCMPhaseReporter) Close() error {
	reporter.mu.Lock()
	if !reporter.closed {
		close(reporter.updates)
		reporter.closed = true
	}
	reporter.mu.Unlock()
	<-reporter.finished
	reporter.mu.Lock()
	defer reporter.mu.Unlock()
	return reporter.err
}

func (reporter *terminalGitCMPhaseReporter) complete(err error) {
	reporter.mu.Lock()
	reporter.err = err
	reporter.mu.Unlock()
	close(reporter.finished)
}

func terminalGitCMPhase(phase cmcommand.Phase) terminalexperience.OperationPhase {
	name := "Staging selected files"
	detail := gitCMPhaseFileDetail(phase.FileCount)
	switch phase.Kind {
	case cmcommand.PhaseCollect:
		name = "Collecting changes"
	case cmcommand.PhaseGenerate:
		name = "Generating commit message"
	case cmcommand.PhaseCommit:
		name = "Creating commit"
		detail = ""
	case cmcommand.PhasePush:
		name = "Pushing commit"
		detail = phase.Remote
	}
	return terminalexperience.OperationPhase{Name: name, Detail: detail, State: terminalGitCMPhaseState(phase.State)}
}

func gitCMPhaseFileDetail(count int) string {
	if count == 0 {
		return ""
	}
	if count == 1 {
		return "1 file"
	}
	return fmt.Sprintf("%d files", count)
}

func terminalGitCMPhaseState(state cmcommand.PhaseState) terminalexperience.PhaseState {
	switch state {
	case cmcommand.PhaseCompleted:
		return terminalexperience.PhaseCompleted
	case cmcommand.PhaseCancelled:
		return terminalexperience.PhaseCancelled
	case cmcommand.PhaseFailed:
		return terminalexperience.PhaseFailed
	default:
		return terminalexperience.PhaseActive
	}
}

func gitCMGeneratedDocument(session terminalexperience.Session, generated cmcommand.GeneratedMessage, profile cmcommand.ProfileDiagnostic) terminalexperience.PresentationDocument {
	messageRole := terminalexperience.VisualRolePlain
	metadataRole := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		messageRole = terminalexperience.VisualRoleSuccess
		metadataRole = terminalexperience.VisualRoleMuted
	}
	coverage := generated.Evidence
	blocks := []terminalexperience.PresentationBlock{
		{Role: messageRole, Text: generated.Message + "\n\n"},
		{Role: metadataRole, Text: fmt.Sprintf("Profile: %s (%s)", profile.Name, profile.Model)},
		{Role: metadataRole, Text: formatGitCMTokenUsage(generated.Usage)},
		{Role: metadataRole, Text: fmt.Sprintf("Local evidence estimate: ~%s serialized prompt tokens / %d of %d clusters / %d of %d facts", formatGitCMCount(float64(coverage.EstimatedLocalPromptTokens)), coverage.RepresentedClusters, coverage.TotalClusters, coverage.IncludedFacts, coverage.IncludedFacts+coverage.OmittedFacts)},
	}
	if coverage.ContentCompacted {
		suffix := "s"
		if coverage.TotalClusters == 1 {
			suffix = ""
		}
		blocks = append(blocks, terminalexperience.PresentationBlock{Role: metadataRole, Text: fmt.Sprintf("Commit scope: %d cluster%s represented with compacted semantic evidence. This does not affect which files are committed.", coverage.TotalClusters, suffix)})
	}
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func gitCMGeneratedText(generated cmcommand.GeneratedMessage, profile cmcommand.ProfileDiagnostic) string {
	return terminalexperience.RenderPlain(gitCMGeneratedDocument(terminalexperience.Session{Kind: terminalexperience.PlainInteractive}, generated, profile))
}

func gitCMFailureDocument(session terminalexperience.Session, profile cmcommand.ProfileDiagnostic) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleMuted
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
		Role: role,
		Text: fmt.Sprintf("Provider: %s\nBase URL: %s\nModel: %s", profile.Name, profile.BaseURL, profile.Model),
	}}}
}

func gitCMOutcomeDocument(session terminalexperience.Session, result cmcommand.Result) terminalexperience.PresentationDocument {
	text := ""
	role := terminalexperience.VisualRolePlain
	switch {
	case result.NoChanges && result.NoChangeScope == cmcommand.ScopeStaged:
		text = "No staged changes."
		role = terminalexperience.VisualRoleWarning
	case result.NoChanges:
		text = "No uncommitted changes."
		role = terminalexperience.VisualRoleWarning
	case result.NothingSelected:
		text = "Nothing selected."
		role = terminalexperience.VisualRoleWarning
	case result.Cancelled:
		text = "Cancelled"
		role = terminalexperience.VisualRoleWarning
	case result.Pushed:
		text = "Commit created and pushed"
		role = terminalexperience.VisualRoleSuccess
	case result.Committed:
		text = "Commit created"
		role = terminalexperience.VisualRoleSuccess
	}
	if text == "" {
		return terminalexperience.PresentationDocument{}
	}
	if session.Kind != terminalexperience.RichInteractive {
		role = terminalexperience.VisualRolePlain
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func formatGitCMTokenUsage(usage *cmcommand.TokenUsage) string {
	if usage == nil {
		return "Provider tokens: unavailable"
	}
	return "Provider tokens: " + formatGitCMTokenValue(usage.PromptTokens) + " prompt / " + formatGitCMTokenValue(usage.CompletionTokens) + " completion / " + formatGitCMTokenValue(usage.TotalTokens) + " total"
}

func formatGitCMTokenValue(value *float64) string {
	if value == nil {
		return "unknown"
	}
	return formatGitCMCount(*value)
}

func formatGitCMCount(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', -1, 64)
	decimal := strings.IndexByte(formatted, '.')
	integer := formatted
	fraction := ""
	if decimal >= 0 {
		integer = formatted[:decimal]
		fraction = formatted[decimal:]
	}
	start := 0
	if strings.HasPrefix(integer, "-") {
		start = 1
	}
	for index := len(integer) - 3; index > start; index -= 3 {
		integer = integer[:index] + "," + integer[index:]
	}
	return integer + fraction
}
