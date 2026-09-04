package test

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

const (
	cmTestResolvePhaseID    = "resolve-cm-test-profile"
	cmTestResolvePhaseName  = "Resolve CM test profile"
	cmTestProviderPhaseID   = "test-cm-provider"
	cmTestProviderPhaseName = "Test CM provider"
)

func runTest(options *Options) error {
	if options == nil || options.Store == nil || options.HTTP == nil || options.Terminal == nil {
		return errors.New("config cm test options are incomplete")
	}
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	run, err := options.Terminal.OpenConsole(ctx, terminalCMTestConsoleDescriptor(options.Profile))
	if err != nil {
		return err
	}
	defer run.Close()
	caps := options.Terminal.Capabilities()
	var presentationErr error
	finish := func(outcome terminalexperience.FinishOutcome, document *terminalexperience.PresentationDocument, workErr error) error {
		return errors.Join(workErr, presentationErr, run.Finish(outcome, document))
	}

	if caps.Interaction == terminalexperience.RichInteractive {
		if err := run.Milestone(terminalCMTestIntroDocument()); err != nil {
			presentationErr = errors.Join(presentationErr, err)
		}
	}

	phases := newCMTestPhaseSink(run, cancel)
	phases.begin(cmTestResolvePhaseID, cmTestResolvePhaseName, "Resolving CM test profile...")
	if err := ctx.Err(); err != nil {
		presentationErr = errors.Join(presentationErr, phases.end(terminalexperience.PhaseCancelled, "Cancelled while resolving CM test profile"))
		return finish(terminalexperience.Cancelled, nil, err)
	}
	type resolutionResult struct {
		module  *TestModule
		profile appconfig.ResolvedCMProfile
		err     error
	}
	resolution := make(chan resolutionResult, 1)
	go func() {
		resolver, err := options.Store()
		if err != nil {
			resolution <- resolutionResult{err: err}
			return
		}
		module, err := NewTest(TestDependencies{Resolver: resolver, Transport: options.HTTP})
		if err != nil {
			resolution <- resolutionResult{err: err}
			return
		}
		profile, err := module.resolveProfile(TestRequest{Profile: options.Profile})
		resolution <- resolutionResult{module: module, profile: profile, err: err}
	}()
	var resolved resolutionResult
	select {
	case resolved = <-resolution:
	case <-ctx.Done():
		// Prefer a result that was already published over a simultaneous
		// cancellation so the phase state reflects completed resolution.
		select {
		case resolved = <-resolution:
		default:
			presentationErr = errors.Join(presentationErr, phases.end(terminalexperience.PhaseCancelled, "Cancelled while resolving CM test profile"))
			return finish(terminalexperience.Cancelled, nil, ctx.Err())
		}
	}
	err = resolved.err
	module := resolved.module
	profile := resolved.profile
	if err != nil {
		state := terminalexperience.PhaseFailed
		outcome := terminalexperience.Failed
		detail := "Unable to resolve CM test profile (" + cmTestResolverFailureKind(err) + ")"
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			state = terminalexperience.PhaseCancelled
			outcome = terminalexperience.Cancelled
			detail = "Cancelled while resolving CM test profile"
		}
		presentationErr = errors.Join(presentationErr, phases.end(state, detail))
		return finish(outcome, nil, err)
	}
	presentationErr = errors.Join(presentationErr, phases.end(terminalexperience.PhaseCompleted, "Profile: "+safeCMTestProfile(redactCMTestText(profile.Name, profile.APIKey))))

	providerDetail := fmt.Sprintf(
		"Provider: %s; Base URL: %s; Model: %s; waiting for response",
		safeCMTestProfile(redactCMTestText(profile.Name, profile.APIKey)),
		safeCMTestURL(redactCMTestText(profile.BaseURL, profile.APIKey)),
		safeCMTestModel(redactCMTestText(profile.Model, profile.APIKey)),
	)
	phases.begin(cmTestProviderPhaseID, cmTestProviderPhaseName, providerDetail)
	if err := ctx.Err(); err != nil {
		presentationErr = errors.Join(presentationErr, phases.end(terminalexperience.PhaseCancelled, "Cancelled while testing CM provider"))
		return finish(terminalexperience.Cancelled, nil, err)
	}
	result, runErr := module.testProvider(ctx, profile)
	if runErr != nil {
		if errors.Is(runErr, context.Canceled) && errors.Is(ctx.Err(), context.Canceled) {
			presentationErr = errors.Join(presentationErr, phases.end(terminalexperience.PhaseCancelled, "Cancelled while testing CM provider"))
			return finish(terminalexperience.Cancelled, nil, runErr)
		}
		category := cmTestProviderFailureKind(runErr)
		presentationErr = errors.Join(presentationErr, phases.end(terminalexperience.PhaseFailed, "Provider request failed ("+string(category)+")"))
		if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
			document := terminalCMTestRichFailureDocument(result, string(category))
			return finish(terminalexperience.Failed, &document, runErr)
		}
		document := terminalCMTestDocument(result)
		return finish(terminalexperience.Failed, &document, runErr)
	}
	presentationErr = errors.Join(presentationErr, phases.end(terminalexperience.PhaseCompleted, "Response received"))
	if caps.Interaction == terminalexperience.RichInteractive {
		if err := run.Milestone(terminalCMTestResponseSummaryDocument(result)); err != nil {
			presentationErr = errors.Join(presentationErr, err)
		}
	}
	var document terminalexperience.PresentationDocument
	if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
		document = terminalCMTestRichDocument(result)
	} else {
		document = terminalCMTestDocument(result)
	}
	return finish(terminalexperience.Succeeded, &document, nil)
}

func terminalCMTestConsoleDescriptor(profile string) terminalexperience.ConsoleDescriptor {
	metadata := []terminalexperience.ConsoleMetadata{{
		Label: "scope",
		Value: "non-mutating provider check",
	}}
	if strings.TrimSpace(profile) != "" {
		metadata = append(metadata, terminalexperience.ConsoleMetadata{
			Label: "profile",
			Value: safeCMTestProfile(profile),
		})
	}
	return terminalexperience.ConsoleDescriptor{
		Command:  "YCY / config cm test",
		Target:   "provider connection",
		Status:   "READY",
		Metadata: metadata,
	}
}

type cmTestPhaseSink struct {
	run           terminalexperience.ExperienceRun
	requestCancel context.CancelFunc
	current       *cmTestPhaseTrack
}

type cmTestPhaseTrack struct {
	id      string
	updates chan terminalexperience.OperationPhase
	done    chan error
}

func newCMTestPhaseSink(run terminalexperience.ExperienceRun, requestCancel context.CancelFunc) *cmTestPhaseSink {
	return &cmTestPhaseSink{run: run, requestCancel: requestCancel}
}

func (sink *cmTestPhaseSink) begin(id, name, detail string) {
	if sink.current != nil {
		_ = sink.end(terminalexperience.PhaseFailed, "Phase interrupted")
	}
	updates := make(chan terminalexperience.OperationPhase, 4)
	done := make(chan error, 1)
	sink.current = &cmTestPhaseTrack{id: id, updates: updates, done: done}
	go func() {
		done <- sink.run.Track(terminalexperience.TrackedOperation{
			ID:            id,
			Label:         "Test commit message provider",
			Phases:        []terminalexperience.PhaseDefinition{{ID: id, Name: name}},
			Updates:       updates,
			RequestCancel: sink.requestCancel,
		})
	}()
	updates <- terminalexperience.OperationPhase{ID: id, State: terminalexperience.PhaseActive, Detail: detail}
}

func (sink *cmTestPhaseSink) end(state terminalexperience.PhaseState, detail string) error {
	if sink.current == nil {
		return nil
	}
	track := sink.current
	sink.current = nil
	track.updates <- terminalexperience.OperationPhase{ID: track.id, State: state, Detail: detail}
	close(track.updates)
	return <-track.done
}

func terminalCMTestIntroDocument() terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / config cm test"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Test commit message provider"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Verify the resolved profile can answer a connection check"},
	}}
}

func terminalCMTestDocument(result TestResult) terminalexperience.PresentationDocument {
	if result.Diagnostic != nil {
		document := terminalCMTestFailureDocument(*result.Diagnostic, terminalexperience.VisualRoleMuted)
		document.Blocks = append([]terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "Commit message provider test"}, {Role: terminalexperience.VisualRoleWarning, Text: "Provider request failed"}}, document.Blocks...)
		return document
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "Commit message provider test"}, {Role: terminalexperience.VisualRolePlain, Text: "Response:\n" + result.Content}, {Role: terminalexperience.VisualRoleSuccess, Text: "Done"}}}
}

func terminalCMTestRichDocument(result TestResult) terminalexperience.PresentationDocument {
	intro := terminalCMTestIntroDocument().Blocks
	intro = append(intro, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRolePlain, Text: "Response:\n" + safeCMTestResponse(result.Content)})
	if usage := terminalCMTestUsageSummary(result.usage); usage != "" {
		intro = append(intro, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: usage})
	}
	intro = append(intro, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleSuccess, Text: "Done"})
	return terminalexperience.PresentationDocument{Blocks: intro}
}

func terminalCMTestResponseSummaryDocument(result TestResult) terminalexperience.PresentationDocument {
	blocks := []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleSuccess, Text: "Response received"}}
	if usage := terminalCMTestUsageSummary(result.usage); usage != "" {
		blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: usage})
	}
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func terminalCMTestUsageSummary(usage *cmTestTokenUsage) string {
	if usage == nil {
		return ""
	}
	parts := make([]string, 0, 3)
	promptTokens, hasPromptTokens := cmTestUsageValue(usage.PromptTokens)
	completionTokens, hasCompletionTokens := cmTestUsageValue(usage.CompletionTokens)
	totalTokens, hasTotalTokens := cmTestUsageValue(usage.TotalTokens)
	if !hasTotalTokens && hasPromptTokens && hasCompletionTokens {
		totalTokens = promptTokens + completionTokens
		hasTotalTokens = true
	}
	if hasPromptTokens {
		parts = append(parts, fmt.Sprintf("Prompt tokens: %g", promptTokens))
	}
	if hasCompletionTokens {
		parts = append(parts, fmt.Sprintf("Completion tokens: %g", completionTokens))
	}
	if hasTotalTokens {
		parts = append(parts, fmt.Sprintf("Total tokens: %g", totalTokens))
	}
	return strings.Join(parts, "  ")
}

func cmTestUsageValue(value *float64) (float64, bool) {
	if value == nil {
		return 0, false
	}
	return cmTestUsageNumber(*value)
}

func terminalCMTestRichFailureDocument(result TestResult, category string) terminalexperience.PresentationDocument {
	blocks := terminalCMTestIntroDocument().Blocks
	blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: "Provider request failed"})
	blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Category: " + safeCMTestField(category, "provider")})
	if result.Diagnostic != nil {
		blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("Provider: %s\nBase URL: %s\nModel: %s", safeCMTestProfile(result.Diagnostic.Provider), safeCMTestURL(result.Diagnostic.BaseURL), safeCMTestModel(result.Diagnostic.Model))})
	}
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func terminalCMTestFailureDocument(diagnostic TestDiagnostic, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: fmt.Sprintf("Provider: %s\nBase URL: %s\nModel: %s", safeCMTestProfile(diagnostic.Provider), safeCMTestURL(diagnostic.BaseURL), safeCMTestModel(diagnostic.Model))}}}
}
