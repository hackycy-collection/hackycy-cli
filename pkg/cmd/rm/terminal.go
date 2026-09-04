package rm

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

// runRM owns only the terminal projection. The Module/plan/delete functions
// remain the command's compatibility boundary and are intentionally reused.
func runRM(options *Options) error {
	if options == nil || options.Terminal == nil || options.WorkingDirectory == nil || options.Remover == nil {
		return errors.New("rm options are incomplete")
	}
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	run := options.Terminal.Open(ctx)
	defer run.Close()
	caps := options.Terminal.Capabilities()
	if err := ctx.Err(); err != nil {
		return errors.Join(err, run.Finish(terminalexperience.Cancelled, nil))
	}
	if caps.Interaction == terminalexperience.RichInteractive {
		if err := run.Notice(terminalRMIntroDocument()); err != nil {
			return errors.Join(err, run.Finish(terminalexperience.Failed, nil))
		}
	}
	sink := newRMPhaseSink(run, caps)
	workingDirectory, err := options.WorkingDirectory()
	if err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	if err := ctx.Err(); err != nil {
		return finishRM(run, sink, terminalexperience.Cancelled, nil, err)
	}
	workingDirectory, err = filepath.Abs(workingDirectory)
	if err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	if err := ctx.Err(); err != nil {
		return finishRM(run, sink, terminalexperience.Cancelled, nil, err)
	}
	adapter := newTerminalRMAdapter(run)
	input := Input{Paths: append([]string(nil), options.Paths...), Force: options.Force, Depth: options.Depth}
	if len(input.Paths) > 0 {
		return runRMExplicitTerminal(ctx, caps, sink, adapter, options.Remover, workingDirectory, input, run)
	}
	return runRMSmartTerminal(ctx, caps, sink, adapter, options.Remover, workingDirectory, input, run)
}

func runRMExplicitTerminal(
	ctx context.Context,
	caps terminalexperience.Capabilities,
	sink *rmPhaseSink,
	adapter *terminalRMAdapter,
	remover PathRemover,
	workingDirectory string,
	input Input,
	run terminalexperience.ExperienceRun,
) error {
	if err := sink.begin("resolve-explicit-targets", "Resolve explicit targets", "Resolving explicit targets"); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	plan, err := planExplicit(workingDirectory, input.Paths)
	if err != nil {
		_ = sink.end(terminalexperience.PhaseFailed, "Unable to resolve explicit targets")
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	if err := ctx.Err(); err != nil {
		_ = sink.end(terminalexperience.PhaseCancelled, "Resolving explicit targets cancelled")
		return finishRM(run, sink, terminalexperience.Cancelled, nil, err)
	}
	resolveDetail := fmt.Sprintf("Resolved %d target%s; %d missing", len(plan.existing), rmPlural(len(plan.existing)), len(plan.missing))
	if err := sink.end(terminalexperience.PhaseCompleted, resolveDetail); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	if err := presentRMMissing(caps, run, workingDirectory, plan.missing); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	if len(plan.existing) == 0 {
		document := terminalRMNoValidPathsDocument(caps, workingDirectory, plan.missing)
		if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
			document = terminalRMRichResult("Paths removed", "No valid paths to delete.", terminalexperience.VisualRoleWarning)
		}
		return finishRM(run, sink, terminalexperience.Succeeded, &document, nil)
	}
	if !input.Force {
		if err := presentRMExplicitTargets(caps, run, workingDirectory, plan.existing); err != nil {
			return finishRM(run, sink, terminalexperience.Failed, nil, err)
		}
		description := "Recursive deletion removes all contents. Targets: " + rmPathSummary(workingDirectory, plan.existing)
		confirmed, cancelled, promptErr := adapter.ConfirmExplicit(ExplicitConfirmationPrompt{
			Message:     fmt.Sprintf("Delete %d item%s?", len(plan.existing), rmPlural(len(plan.existing))),
			Initial:     false,
			Description: description,
		})
		if promptErr != nil {
			return finishRMInteractionError(run, sink, promptErr)
		}
		if cancelled || !confirmed {
			document := terminalRMDocument("Cancelled.", terminalexperience.VisualRoleWarning)
			return finishRM(run, sink, terminalexperience.Cancelled, &document, nil)
		}
	}
	if err := ctx.Err(); err != nil {
		return finishRM(run, sink, terminalexperience.Cancelled, nil, err)
	}
	if err := sink.begin("delete-selected-paths", "Delete selected paths", fmt.Sprintf("Deleting %d target%s", len(plan.existing), rmPlural(len(plan.existing)))); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	result := deletePaths(plan.existing, remover)
	phaseDetail := rmDeletionDetail(len(plan.existing), result)
	if err := sink.end(terminalexperience.PhaseCompleted, phaseDetail); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	if err := presentRMDeletion(caps, run, workingDirectory, result); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	document := terminalRMDocument("Done!", terminalexperience.VisualRoleSuccess)
	if caps.Interaction == terminalexperience.Automation {
		document = terminalRMAutomationDeletionDocument(workingDirectory, plan.missing, result)
	} else if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
		document = terminalRMRichResult("Paths removed", "Done!", terminalexperience.VisualRoleSuccess)
	}
	return finishRM(run, sink, terminalexperience.Succeeded, &document, nil)
}

func runRMSmartTerminal(
	ctx context.Context,
	caps terminalexperience.Capabilities,
	sink *rmPhaseSink,
	adapter *terminalRMAdapter,
	remover PathRemover,
	workingDirectory string,
	input Input,
	run terminalexperience.ExperienceRun,
) error {
	action, cancelled, err := adapter.SelectSmartAction(SmartActionPrompt{Message: "Select a clean action", Options: append([]SmartAction(nil), smartActions...)})
	if err != nil {
		return finishRMInteractionError(run, sink, err)
	}
	if cancelled {
		document := terminalRMDocument("Cancelled.", terminalexperience.VisualRoleWarning)
		return finishRM(run, sink, terminalexperience.Cancelled, &document, nil)
	}
	if err := ctx.Err(); err != nil {
		return finishRM(run, sink, terminalexperience.Cancelled, nil, err)
	}
	if err := sink.begin("scan-cleanup-targets", "Scan cleanup targets", "Scanning cleanup targets"); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	targets := discoverSmart(workingDirectory, action, resolvedSmartDepth(input.Depth))
	if err := ctx.Err(); err != nil {
		_ = sink.end(terminalexperience.PhaseCancelled, "Scanning cleanup targets cancelled")
		return finishRM(run, sink, terminalexperience.Cancelled, nil, err)
	}
	detail := fmt.Sprintf("Found %d target%s", len(targets), rmPlural(len(targets)))
	if err := sink.end(terminalexperience.PhaseCompleted, detail); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	if len(targets) == 0 {
		if err := presentRMSmartEmpty(caps, run); err != nil {
			return finishRM(run, sink, terminalexperience.Failed, nil, err)
		}
		document := terminalRMDocument("Nothing to clean.", terminalexperience.VisualRoleSuccess)
		if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
			document = terminalRMRichResult("Cleanup complete", "Nothing to clean.", terminalexperience.VisualRoleSuccess)
		}
		return finishRM(run, sink, terminalexperience.Succeeded, &document, nil)
	}
	selected := append([]string(nil), targets...)
	if !input.Force {
		options := make([]SmartTargetChoice, 0, len(targets))
		for _, target := range targets {
			relative, relErr := filepath.Rel(workingDirectory, target)
			if relErr != nil {
				return finishRM(run, sink, terminalexperience.Failed, nil, relErr)
			}
			options = append(options, SmartTargetChoice{Value: target, Label: safeRMPathLabel(relative)})
		}
		selected, cancelled, err = adapter.SelectSmartTargets(SmartTargetPrompt{Message: "Select items to delete", Options: options, InitialValues: append([]string(nil), targets...)})
		if err != nil {
			return finishRMInteractionError(run, sink, err)
		}
		if cancelled {
			document := terminalRMDocument("Cancelled.", terminalexperience.VisualRoleWarning)
			return finishRM(run, sink, terminalexperience.Cancelled, &document, nil)
		}
	}
	if len(selected) == 0 {
		document := terminalRMDocument("Nothing selected.", terminalexperience.VisualRoleWarning)
		return finishRM(run, sink, terminalexperience.Cancelled, &document, nil)
	}
	if err := ctx.Err(); err != nil {
		return finishRM(run, sink, terminalexperience.Cancelled, nil, err)
	}
	if err := sink.begin("delete-selected-paths", "Delete selected paths", fmt.Sprintf("Deleting %d target%s", len(selected), rmPlural(len(selected)))); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	result := deletePaths(selected, remover)
	if err := sink.end(terminalexperience.PhaseCompleted, rmDeletionDetail(len(selected), result)); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	if err := presentRMDeletion(caps, run, workingDirectory, result); err != nil {
		return finishRM(run, sink, terminalexperience.Failed, nil, err)
	}
	document := terminalRMDocument("Done!", terminalexperience.VisualRoleSuccess)
	if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
		document = terminalRMRichResult("Cleanup complete", "Done!", terminalexperience.VisualRoleSuccess)
	}
	return finishRM(run, sink, terminalexperience.Succeeded, &document, nil)
}

type rmPhaseSink struct {
	run       terminalexperience.ExperienceRun
	caps      terminalexperience.Capabilities
	updates   chan terminalexperience.OperationPhase
	done      chan error
	active    bool
	trackErr  error
	sequence  uint64
	currentID string
}

func newRMPhaseSink(run terminalexperience.ExperienceRun, caps terminalexperience.Capabilities) *rmPhaseSink {
	return &rmPhaseSink{run: run, caps: caps}
}

func (sink *rmPhaseSink) begin(id, name, detail string) error {
	if sink.active {
		return errors.New("rm phase already active")
	}
	if sink.caps.Interaction == terminalexperience.PlainInteractive {
		return sink.run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleActive, Text: name + "..."}}})
	}
	if sink.caps.Interaction != terminalexperience.RichInteractive {
		return nil
	}
	sink.sequence++
	sink.currentID = id
	sink.updates = make(chan terminalexperience.OperationPhase, 4)
	sink.done = make(chan error, 1)
	sink.active = true
	go func() {
		sink.done <- sink.run.Track(terminalexperience.TrackedOperation{
			ID:      fmt.Sprintf("rm-%d", sink.sequence),
			Label:   name,
			Phases:  []terminalexperience.PhaseDefinition{{ID: id, Name: name}},
			Updates: sink.updates,
		})
	}()
	sink.updates <- terminalexperience.OperationPhase{ID: id, State: terminalexperience.PhaseActive, Detail: detail}
	return nil
}

func (sink *rmPhaseSink) end(state terminalexperience.PhaseState, detail string) error {
	if sink.caps.Interaction != terminalexperience.RichInteractive {
		return nil
	}
	if !sink.active {
		return errors.New("rm phase is not active")
	}
	sink.updates <- terminalexperience.OperationPhase{ID: sink.phaseID(), State: state, Detail: detail}
	close(sink.updates)
	sink.active = false
	sink.trackErr = errors.Join(sink.trackErr, <-sink.done)
	return sink.trackErr
}

func (sink *rmPhaseSink) phaseID() string {
	// The current operation's ID is carried by the first update already queued.
	// Keep a copy by reading the channel-independent field instead of relying on
	// display names, which may contain hostile input.
	return sink.currentID
}

func (sink *rmPhaseSink) closeActive() error {
	if sink.caps.Interaction != terminalexperience.RichInteractive || !sink.active {
		return sink.trackErr
	}
	close(sink.updates)
	sink.active = false
	sink.trackErr = errors.Join(sink.trackErr, <-sink.done)
	return sink.trackErr
}

func finishRM(run terminalexperience.ExperienceRun, sink *rmPhaseSink, outcome terminalexperience.FinishOutcome, document *terminalexperience.PresentationDocument, workErr error) error {
	return errors.Join(workErr, sink.closeActive(), run.Finish(outcome, document))
}

func finishRMInteractionError(run terminalexperience.ExperienceRun, sink *rmPhaseSink, workErr error) error {
	outcome := terminalexperience.Failed
	if rmContextCancelled(workErr) {
		outcome = terminalexperience.Cancelled
	}
	return finishRM(run, sink, outcome, nil, workErr)
}

func rmContextCancelled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func terminalRMIntroDocument() terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / rm"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Remove"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Remove selected files or clean project artifacts"},
	}}
}

func terminalRMRichResult(title, message string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / rm"},
		{Role: terminalexperience.VisualRoleTitle, Text: title},
		{Role: role, Text: message},
	}}
}

func terminalRMNoValidPathsDocument(caps terminalexperience.Capabilities, root string, missing []string) terminalexperience.PresentationDocument {
	if caps.Interaction != terminalexperience.Automation {
		return terminalRMDocument("No valid paths to delete.", terminalexperience.VisualRoleWarning)
	}
	blocks := rmMissingBlocks(root, missing)
	blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: "No valid paths to delete."})
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

// terminalRMAutomationDeletionDocument preserves the legacy command-result
// facts on successful noninteractive deletions without reintroducing terminal
// interactions, Work Phase diagnostics, or an unbounded path list.
func terminalRMAutomationDeletionDocument(root string, missing []string, result deletionResult) terminalexperience.PresentationDocument {
	blocks := rmMissingBlocks(root, missing)
	blocks = append(blocks, terminalexperience.PresentationBlock{
		Role: terminalexperience.VisualRoleMuted,
		Text: fmt.Sprintf("Deleted %d item%s", result.succeeded, rmPlural(result.succeeded)),
	})
	for _, failure := range result.failures {
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: terminalexperience.VisualRoleWarning,
			Text: "  skipped: " + safeRMText(failure.Error()),
		})
	}
	blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleSuccess, Text: "Done!"})
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func rmMissingBlocks(root string, paths []string) []terminalexperience.PresentationBlock {
	blocks := make([]terminalexperience.PresentationBlock, 0, len(paths))
	for _, path := range paths {
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: terminalexperience.VisualRoleWarning,
			Text: "  not found, skipping: " + safeRMPathLabel(relativeRMPath(root, path)),
		})
	}
	return blocks
}

func rmPlural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func rmDeletionDetail(requested int, result deletionResult) string {
	return fmt.Sprintf("Requested: %d; Succeeded: %d; Failed: %d", requested, result.succeeded, len(result.failures))
}

func presentRMMissing(caps terminalexperience.Capabilities, run terminalexperience.ExperienceRun, root string, paths []string) error {
	if caps.Interaction == terminalexperience.Automation {
		return nil
	}
	if caps.Interaction == terminalexperience.RichInteractive {
		if len(paths) == 0 {
			return nil
		}
		return run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
			{Role: terminalexperience.VisualRoleWarning, Text: fmt.Sprintf("Missing paths skipped: %d", len(paths))},
			{Role: terminalexperience.VisualRoleMuted, Text: "Paths: " + rmPathSummary(root, paths)},
		}})
	}
	for _, path := range paths {
		relative := safeRMPathLabel(relativeRMPath(root, path))
		if err := run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "  not found, skipping: " + relative}}}); err != nil {
			return err
		}
	}
	return nil
}

func presentRMExplicitTargets(caps terminalexperience.Capabilities, run terminalexperience.ExperienceRun, root string, paths []string) error {
	text := "Targets: " + rmPathSummary(root, paths)
	if caps.Interaction == terminalexperience.RichInteractive {
		blocks := []terminalexperience.PresentationBlock{
			{Role: terminalexperience.VisualRoleWarning, Text: text},
			{Role: terminalexperience.VisualRoleMuted, Text: "Recursive deletion removes all contents."},
		}
		for _, warning := range rmExplicitRiskWarnings(root, paths) {
			blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: warning})
		}
		return run.Milestone(terminalexperience.PresentationDocument{Blocks: blocks})
	}
	if caps.Interaction == terminalexperience.Automation {
		return nil
	}
	return run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleMuted, Text: text}}})
}

func presentRMSmartEmpty(caps terminalexperience.Capabilities, run terminalexperience.ExperienceRun) error {
	if caps.Interaction == terminalexperience.Automation {
		return nil
	}
	return run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleMuted, Text: "No targets found."}}})
}

func presentRMDeletion(caps terminalexperience.Capabilities, run terminalexperience.ExperienceRun, root string, result deletionResult) error {
	if caps.Interaction == terminalexperience.Automation {
		return nil
	}
	if caps.Interaction == terminalexperience.RichInteractive {
		blocks := []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("Deleted %d item%s", result.succeeded, rmPlural(result.succeeded))}}
		for _, category := range rmFailureCategories(result.failures) {
			blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: "  skipped (" + category + ")"})
		}
		return run.Milestone(terminalexperience.PresentationDocument{Blocks: blocks})
	}
	if err := run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("Deleted %d item%s", result.succeeded, rmPlural(result.succeeded))}}}); err != nil {
		return err
	}
	for _, failure := range result.failures {
		if err := run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "  skipped: " + safeRMText(failure.Error())}}}); err != nil {
			return err
		}
	}
	return nil
}

func rmExplicitRiskWarnings(root string, paths []string) []string {
	var containsRootOrParent, containsOutside, containsDuplicate bool
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if _, exists := seen[path]; exists {
			containsDuplicate = true
		}
		seen[path] = struct{}{}
		relative, err := filepath.Rel(path, root)
		if err == nil && (relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative))) {
			containsRootOrParent = true
		}
		relative, err = filepath.Rel(root, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
			containsOutside = true
		}
	}
	warnings := make([]string, 0, 3)
	if containsRootOrParent {
		warnings = append(warnings, "Warning: target includes the current directory or a parent scope.")
	}
	if containsOutside {
		warnings = append(warnings, "Warning: target is outside the current directory.")
	}
	if containsDuplicate {
		warnings = append(warnings, "Warning: duplicate target selected.")
	}
	return warnings
}

func rmFailureCategories(failures []error) []string {
	categories := make([]string, 0, len(failures))
	seen := make(map[string]struct{}, len(failures))
	for _, failure := range failures {
		category := rmFailureCategory(failure)
		if _, exists := seen[category]; exists {
			continue
		}
		seen[category] = struct{}{}
		categories = append(categories, category)
	}
	return categories
}

func rmPathSummary(root string, paths []string) string {
	const maxPaths = 8
	labels := make([]string, 0, minInt(len(paths), maxPaths))
	for index, path := range paths {
		if index >= maxPaths {
			break
		}
		labels = append(labels, safeRMPathLabel(relativeRMPath(root, path)))
	}
	if len(paths) > maxPaths {
		labels = append(labels, fmt.Sprintf("+%d more", len(paths)-maxPaths))
	}
	return strings.Join(labels, ", ")
}

func relativeRMPath(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "" {
		return "target"
	}
	return filepath.ToSlash(relative)
}

func safeRMPathLabel(value string) string {
	if !utf8.ValidString(value) {
		return "path"
	}
	var builder strings.Builder
	for _, r := range value {
		switch r {
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			if unicode.IsControl(r) {
				return "path"
			}
			builder.WriteRune(r)
		}
	}
	value = builder.String()
	if value == "" {
		return "path"
	}
	runes := []rune(value)
	if len(runes) > 160 {
		return string(runes[:160]) + "..."
	}
	return value
}

func safeRMText(value string) string {
	if !utf8.ValidString(value) {
		return "filesystem failure"
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "filesystem failure"
		}
	}
	if len([]rune(value)) > 256 {
		return string([]rune(value)[:256]) + "..."
	}
	return value
}

func rmFailureCategory(err error) string {
	if err == nil {
		return "filesystem"
	}
	value := strings.ToLower(err.Error())
	switch {
	case strings.Contains(value, "permission") || strings.Contains(value, "denied"):
		return "permission"
	case strings.Contains(value, "not found") || strings.Contains(value, "no such"):
		return "not-found"
	case strings.Contains(value, "path"):
		return "path"
	default:
		return "filesystem"
	}
}

func minInt(first, second int) int {
	if first < second {
		return first
	}
	return second
}
