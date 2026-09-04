package zip

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

const (
	zipDiscoverWorkspacePhaseID = "discover-workspace"
	zipSelectSourcePhaseID      = "select-source"
	zipSelectPatternsPhaseID    = "select-patterns"
	zipPrepareArchivePhaseID    = "prepare-archive"
	zipCollectFilesPhaseID      = "collect-files"
	zipCompressFilesPhaseID     = "compress-files"
	zipWriteArchivePhaseID      = "write-archive"
	zipRevealArchivePhaseID     = "reveal-archive"
)

var zipPhaseDefinitions = []terminalexperience.PhaseDefinition{
	{ID: zipDiscoverWorkspacePhaseID, Name: "Discover workspace"},
	{ID: zipSelectSourcePhaseID, Name: "Select source"},
	{ID: zipSelectPatternsPhaseID, Name: "Select patterns"},
	{ID: zipPrepareArchivePhaseID, Name: "Prepare archive"},
	{ID: zipCollectFilesPhaseID, Name: "Collect files"},
	{ID: zipCompressFilesPhaseID, Name: "Compress files"},
	{ID: zipWriteArchivePhaseID, Name: "Write archive"},
	{ID: zipRevealArchivePhaseID, Name: "Reveal archive"},
}

// runZIP is the terminal-owned adapter for the archive command. Module.Run
// remains available for compatibility tests and does not know about terminal
// capabilities or output streams.
func runZIP(options *Options) error {
	if options == nil || options.Terminal == nil {
		return errors.New("zip options are incomplete")
	}
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	run, err := options.Terminal.OpenConsole(ctx, terminalZipConsoleDescriptor(options))
	if err != nil {
		return err
	}
	defer run.Close()
	caps := options.Terminal.Capabilities()
	if caps.Interaction == terminalexperience.Automation {
		return errors.Join(errZipRequiresInteractive, run.Finish(terminalexperience.Failed, nil))
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(err, run.Finish(terminalexperience.Cancelled, nil))
	}

	presenter := &terminalZipPresenter{run: run}
	phases := newZipPhaseCoordinator(run, caps)
	adapter := newTerminalZipAdapter(run)
	module, err := New(Dependencies{
		Prompter:           adapter,
		Presenter:          presenter,
		RemoteNameResolver: newZipRemoteNameResolver(osZipRemoteOutputRunner{}),
		Revealer:           newHostZipRevealer(osZipHostCommandRunner{}),
		Phases:             phases,
	})
	if err != nil {
		return errors.Join(err, run.Finish(terminalexperience.Failed, nil))
	}
	result, workErr := module.RunContext(ctx, Input{
		Directory: options.Directory,
		Open:      options.Open,
		WithDir:   options.WithDir,
	})
	return finishTerminalZIP(run, caps, phases, presenter, result, workErr)
}

type terminalZipPresenter struct {
	run terminalexperience.ExperienceRun
	err error
}

func (presenter *terminalZipPresenter) Intro() {
	presenter.capture(presenter.run.Notice(terminalZipIntroDocument()))
}

func (presenter *terminalZipPresenter) Note(note PlanningNote) {
	presenter.capture(presenter.run.Notice(terminalZipPlanningNoteDocument(note)))
}

// Progress is represented by Work Phases. Keeping this method a no-op avoids
// replaying the legacy spinner/status lines in addition to the semantic phase.
func (*terminalZipPresenter) Progress(string) {}

// Module.Run reports the final command result to its caller. The terminal
// adapter submits it after the result is classified, so these compatibility
// callbacks intentionally do not write a second result.
func (*terminalZipPresenter) Cancel(string) {}
func (*terminalZipPresenter) Outro(string)  {}

func (presenter *terminalZipPresenter) capture(err error) {
	if err != nil && presenter.err == nil {
		presenter.err = err
	}
}

func terminalZipIntroDocument() terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / zip"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Zip Directory"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Plan and publish a bounded archive"},
	}}
}

func terminalZipConsoleDescriptor(options *Options) terminalexperience.ConsoleDescriptor {
	directory := "workspace"
	withDir := "disabled"
	reveal := "disabled"
	if options != nil {
		directory = zipDescriptorDirectory(options.Directory)
		if strings.TrimSpace(options.WithDir) != "" {
			withDir = "enabled"
		}
		if options.Open {
			reveal = "enabled"
		}
	}
	return terminalexperience.ConsoleDescriptor{
		Command: "YCY / zip",
		Target:  "Plan and publish a bounded archive",
		Status:  "READY",
		Metadata: []terminalexperience.ConsoleMetadata{
			{Label: "summary", Value: "Zip Directory"},
			{Label: "directory", Value: directory},
			{Label: "with-dir", Value: withDir},
			{Label: "reveal", Value: reveal},
		},
	}
}

func zipDescriptorDirectory(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "workspace"
	}
	if filepath.IsAbs(value) {
		base := filepath.Base(filepath.Clean(value))
		if base == "" || base == "." || base == string(filepath.Separator) {
			return "workspace"
		}
		return safeZipText(base, "workspace")
	}
	return safeZipText(filepath.ToSlash(filepath.Clean(value)), "workspace")
}

func terminalZipPlanningNoteDocument(note PlanningNote) terminalexperience.PresentationDocument {
	blocks := []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleActive, Text: safeZipText(note.Title, "Planning")}}
	if len(note.Lines) > 0 {
		lines := make([]string, 0, len(note.Lines))
		for _, line := range note.Lines {
			lines = append(lines, safeZipText(line, "Planning detail"))
		}
		blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: strings.Join(lines, "\n")})
	}
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

type zipPhaseCoordinator struct {
	run  terminalexperience.ExperienceRun
	caps terminalexperience.Capabilities

	mu        sync.Mutex
	err       error
	finals    map[string]terminalexperience.OperationPhase
	active    map[string]terminalexperience.OperationPhase
	order     []string
	updates   chan terminalexperience.OperationPhase
	trackDone chan error
	tracking  bool
	closed    bool
}

func newZipPhaseCoordinator(run terminalexperience.ExperienceRun, caps terminalexperience.Capabilities) *zipPhaseCoordinator {
	return &zipPhaseCoordinator{
		run:    run,
		caps:   caps,
		finals: make(map[string]terminalexperience.OperationPhase),
		active: make(map[string]terminalexperience.OperationPhase),
	}
}

func (coordinator *zipPhaseCoordinator) Report(update terminalexperience.OperationPhase) error {
	if update.ID == "" {
		update.ID = update.PhaseID
	}
	if update.ID == "" {
		return errors.New("zip phase ID is required")
	}
	if coordinator.isArchivePhase(update.ID) {
		if coordinator.tracksRich() {
			if err := coordinator.startArchiveTrack(); err != nil {
				return err
			}
			coordinator.updates <- update
			return nil
		}
	}

	return coordinator.recordAndNotice(update)
}

func (coordinator *zipPhaseCoordinator) recordAndNotice(update terminalexperience.OperationPhase) error {
	coordinator.mu.Lock()
	if coordinator.closed {
		coordinator.mu.Unlock()
		return terminalexperience.ErrExperienceRunFinished
	}
	if update.State == terminalexperience.PhaseActive {
		coordinator.active[update.ID] = update
		if !containsString(coordinator.order, update.ID) {
			coordinator.order = append(coordinator.order, update.ID)
		}
	} else if terminalPhaseIsFinal(update.State) {
		delete(coordinator.active, update.ID)
		coordinator.finals[update.ID] = update
		if !containsString(coordinator.order, update.ID) {
			coordinator.order = append(coordinator.order, update.ID)
		}
	}
	coordinator.mu.Unlock()

	// Planning interactions cannot share Track's serialized operation lock with
	// Ask. Notices keep the current Huh form responsive; Rich replays the final
	// planning states into the tracked ledger once archive work starts.
	return coordinator.run.Notice(zipPhaseDocument(update))
}

func (coordinator *zipPhaseCoordinator) tracksRich() bool {
	return coordinator.caps.Interaction == terminalexperience.RichInteractive
}

func (coordinator *zipPhaseCoordinator) startArchiveTrack() error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	if coordinator.tracking {
		return coordinator.err
	}
	coordinator.updates = make(chan terminalexperience.OperationPhase, 64)
	coordinator.trackDone = make(chan error, 1)
	coordinator.tracking = true
	go func() {
		coordinator.trackDone <- coordinator.run.Track(terminalexperience.TrackedOperation{
			ID:      "zip-archive",
			Label:   "Create archive",
			Phases:  append([]terminalexperience.PhaseDefinition(nil), zipPhaseDefinitions...),
			Updates: coordinator.updates,
		})
	}()
	// The planning phases were observed before the archive track was needed.
	// Replay their semantic active/final pairs in order so the bounded ledger
	// still contains the complete command lifecycle.
	for _, id := range coordinator.order {
		final, ok := coordinator.finals[id]
		if !ok {
			continue
		}
		coordinator.updates <- terminalexperience.OperationPhase{ID: id, State: terminalexperience.PhaseActive, Detail: final.Detail}
		coordinator.updates <- final
	}
	return nil
}

func (coordinator *zipPhaseCoordinator) markOpen(state terminalexperience.PhaseState, detail string) {
	coordinator.mu.Lock()
	updates := make([]terminalexperience.OperationPhase, 0, len(coordinator.active))
	for id := range coordinator.active {
		updates = append(updates, terminalexperience.OperationPhase{ID: id, State: state, Detail: detail})
	}
	coordinator.mu.Unlock()
	for _, update := range updates {
		_ = coordinator.Report(update)
	}
}

func (coordinator *zipPhaseCoordinator) finish() error {
	coordinator.mu.Lock()
	if coordinator.closed {
		err := coordinator.err
		coordinator.mu.Unlock()
		return err
	}
	coordinator.closed = true
	tracking := coordinator.tracking
	updates := coordinator.updates
	done := coordinator.trackDone
	coordinator.mu.Unlock()

	if tracking {
		close(updates)
		coordinator.mu.Lock()
		trackErr := <-done
		coordinator.err = errors.Join(coordinator.err, trackErr)
		err := coordinator.err
		coordinator.mu.Unlock()
		return err
	}
	if !coordinator.tracksRich() {
		return coordinator.err
	}

	// A planning-only cancellation/failure still receives phase evidence. No
	// Ask follows this point, so a short synchronous tracked operation is safe.
	coordinator.mu.Lock()
	hasFinals := len(coordinator.finals) > 0
	order := append([]string(nil), coordinator.order...)
	finals := make(map[string]terminalexperience.OperationPhase, len(coordinator.finals))
	for id, phase := range coordinator.finals {
		finals[id] = phase
	}
	coordinator.mu.Unlock()
	if !hasFinals || coordinator.caps.Interaction == terminalexperience.Automation {
		return coordinator.err
	}
	updates = make(chan terminalexperience.OperationPhase, len(order)*2)
	done = make(chan error, 1)
	go func() {
		done <- coordinator.run.Track(terminalexperience.TrackedOperation{
			ID:      "zip-planning",
			Label:   "Plan archive",
			Phases:  append([]terminalexperience.PhaseDefinition(nil), zipPhaseDefinitions...),
			Updates: updates,
		})
	}()
	for _, id := range order {
		if phase, ok := finals[id]; ok {
			updates <- terminalexperience.OperationPhase{ID: id, State: terminalexperience.PhaseActive, Detail: phase.Detail}
			updates <- phase
		}
	}
	close(updates)
	trackErr := <-done
	coordinator.mu.Lock()
	coordinator.err = errors.Join(coordinator.err, trackErr)
	err := coordinator.err
	coordinator.mu.Unlock()
	return err
}

func (coordinator *zipPhaseCoordinator) isArchivePhase(id string) bool {
	switch id {
	case zipCollectFilesPhaseID, zipCompressFilesPhaseID, zipWriteArchivePhaseID, zipRevealArchivePhaseID:
		return true
	default:
		return false
	}
}

func terminalPhaseIsFinal(state terminalexperience.PhaseState) bool {
	return state == terminalexperience.PhaseCompleted || state == terminalexperience.PhaseCancelled || state == terminalexperience.PhaseFailed
}

func zipPhaseDocument(update terminalexperience.OperationPhase) terminalexperience.PresentationDocument {
	name := zipPhaseName(update.ID)
	text := name
	if update.Detail != "" {
		text += ": " + safeZipText(update.Detail, "Phase detail")
	}
	role := terminalexperience.VisualRoleActive
	switch update.State {
	case terminalexperience.PhaseCompleted:
		role = terminalexperience.VisualRoleSuccess
	case terminalexperience.PhaseCancelled:
		role = terminalexperience.VisualRoleWarning
	case terminalexperience.PhaseFailed:
		role = terminalexperience.VisualRoleError
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func zipPhaseName(id string) string {
	for _, definition := range zipPhaseDefinitions {
		if definition.ID == id {
			return definition.Name
		}
	}
	return "Archive phase"
}

func finishTerminalZIP(
	run terminalexperience.ExperienceRun,
	caps terminalexperience.Capabilities,
	phases *zipPhaseCoordinator,
	presenter *terminalZipPresenter,
	result Result,
	workErr error,
) error {
	if workErr != nil {
		outcome := terminalexperience.Failed
		detail := "Archive operation failed"
		state := terminalexperience.PhaseFailed
		if errors.Is(workErr, context.Canceled) || errors.Is(workErr, context.DeadlineExceeded) {
			outcome = terminalexperience.Cancelled
			detail = "Archive operation cancelled"
			state = terminalexperience.PhaseCancelled
		}
		phases.markOpen(state, detail)
		return errors.Join(workErr, presenter.err, phases.finish(), run.Finish(outcome, nil))
	}

	outcome := terminalexperience.Succeeded
	if result.Kind == ResultCancelled {
		outcome = terminalexperience.Cancelled
		phases.markOpen(terminalexperience.PhaseCancelled, "Archive planning cancelled")
	} else if result.Kind != ResultCompleted {
		outcome = terminalexperience.Failed
		phases.markOpen(terminalexperience.PhaseFailed, zipResultFailureDetail(result.Kind))
	}

	document := terminalZipResultDocument(result, caps)
	return errors.Join(presenter.err, phases.finish(), run.Finish(outcome, &document))
}

func terminalZipResultDocument(result Result, caps terminalexperience.Capabilities) terminalexperience.PresentationDocument {
	if result.Kind == ResultCompleted {
		if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
			blocks := []terminalexperience.PresentationBlock{
				{Role: terminalexperience.VisualRoleMuted, Text: "YCY / zip"},
				{Role: terminalexperience.VisualRoleTitle, Text: "Archive ready"},
				{Role: terminalexperience.VisualRoleSuccess, Text: "Done!"},
			}
			if result.Plan != nil {
				blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("%d files included; output %s.zip", result.IncludedCount, safeZipName(result.Plan.File))})
			}
			if result.RevealFailed {
				blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: "Archive created; host reveal unavailable"})
			}
			return terminalexperience.PresentationDocument{Blocks: blocks}
		}
		return terminalZipDocument("Done!", terminalexperience.VisualRoleSuccess)
	}

	message := "Operation cancelled."
	role := terminalexperience.VisualRoleWarning
	switch result.Kind {
	case ResultDirectoryNotFound:
		message = "Directory not found: " + safeZipPath(result.Plan)
		role = terminalexperience.VisualRoleError
	case ResultPathNotDirectory:
		message = "Path is not a directory: " + safeZipPath(result.Plan)
		role = terminalexperience.VisualRoleError
	case ResultNoFiles:
		message = "No files matched the selected patterns."
		role = terminalexperience.VisualRoleWarning
	case ResultNoValidFiles:
		message = "No valid files matched after filtering."
		role = terminalexperience.VisualRoleWarning
	case ResultCollectionFailed:
		message = "File collection failed (collection)."
		role = terminalexperience.VisualRoleError
	case ResultCompressionFailed:
		message = "Compression failed (compression)."
		role = terminalexperience.VisualRoleError
	case ResultWriteFailed:
		message = "Failed to write zip (write)."
		role = terminalexperience.VisualRoleError
	}
	return terminalZipDocument(message, role)
}

func zipResultFailureDetail(kind ResultKind) string {
	switch kind {
	case ResultDirectoryNotFound:
		return "Directory unavailable (directory)"
	case ResultPathNotDirectory:
		return "Selected path is not a directory (path)"
	case ResultNoFiles:
		return "No files matched (no-files)"
	case ResultNoValidFiles:
		return "No valid files remained (no-valid-files)"
	case ResultCollectionFailed:
		return "Collection failed (collection)"
	case ResultCompressionFailed:
		return "Compression failed (compression)"
	case ResultWriteFailed:
		return "Publication failed (write)"
	default:
		return "Archive operation failed"
	}
}

func safeZipPath(plan *ZipPlan) string {
	if plan == nil {
		return "path"
	}
	return safeZipText(NormalizeRelativePath(plan.PackageRoot, plan.Input), "path")
}

func safeZipRelativePath(root, value string) string {
	if root == "" {
		return safeZipText(value, "source")
	}
	return safeZipText(NormalizeRelativePath(root, value), "source")
}

func safeZipName(value string) string {
	return safeZipText(value, "archive")
}

func safeZipText(value, fallback string) string {
	if !utf8.ValidString(value) {
		return fallback
	}
	var builder strings.Builder
	for _, character := range value {
		switch character {
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			if unicode.IsControl(character) {
				return fallback
			}
			builder.WriteRune(character)
		}
	}
	value = strings.TrimSpace(builder.String())
	if value == "" {
		return fallback
	}
	runes := []rune(value)
	if len(runes) > 256 {
		return string(runes[:256]) + "..."
	}
	return value
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
