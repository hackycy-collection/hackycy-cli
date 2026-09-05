package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/updater"
)

type upgradePhaseSink struct {
	run          terminalexperience.ExperienceRun
	capabilities terminalexperience.Capabilities
	cancel       context.CancelFunc

	mu              sync.Mutex
	current         *upgradePhaseTrack
	pendingConsume  *updater.UpgradePhaseEvent
	previousResult  *terminalexperience.PresentationDocument
	presentationErr error
}

type upgradePhaseTrack struct {
	phase   updater.UpgradePhase
	updates chan terminalexperience.OperationPhase
	done    chan error
}

func newUpgradePhaseSink(run terminalexperience.ExperienceRun, capabilities terminalexperience.Capabilities, cancel context.CancelFunc) *upgradePhaseSink {
	return &upgradePhaseSink{run: run, capabilities: capabilities, cancel: cancel}
}

func (sink *upgradePhaseSink) observer() updater.UpgradeObserver {
	return updater.UpgradeObserver{
		Phase:         sink.phase,
		PreviousState: sink.previousState,
	}
}

func (sink *upgradePhaseSink) phase(event updater.UpgradePhaseEvent) {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	if sink.presentationErr != nil {
		return
	}
	if event.State == updater.UpgradePhaseActive {
		if event.Phase == updater.UpgradePhaseConsumeStartupTransaction {
			copy := event
			sink.pendingConsume = &copy
			return
		}
		sink.start(event)
		return
	}
	if sink.pendingConsume != nil && sink.pendingConsume.Phase == event.Phase {
		pending := *sink.pendingConsume
		sink.pendingConsume = nil
		sink.start(pending)
	}
	if sink.current == nil || sink.current.phase != event.Phase {
		sink.recordLocked(errors.New("upgrade phase stream is out of order"))
		return
	}
	sink.current.updates <- terminalexperience.OperationPhase{
		ID:     string(event.Phase),
		Detail: terminalUpgradePhaseDetail(event),
		State:  terminalUpgradePhaseState(event.State),
	}
	close(sink.current.updates)
	sink.recordLocked(<-sink.current.done)
	sink.current = nil
}

func (sink *upgradePhaseSink) start(event updater.UpgradePhaseEvent) {
	if sink.current != nil {
		sink.recordLocked(errors.New("upgrade phase stream overlaps active work"))
		return
	}
	name, ok := terminalUpgradePhaseName(event.Phase)
	if !ok {
		sink.recordLocked(errors.New("upgrade phase is not known to the terminal adapter"))
		return
	}
	track := &upgradePhaseTrack{
		phase:   event.Phase,
		updates: make(chan terminalexperience.OperationPhase, 2),
		done:    make(chan error, 1),
	}
	sink.current = track
	go func() {
		track.done <- sink.run.Track(terminalexperience.TrackedOperation{
			ID:            "upgrade-" + string(event.Phase),
			OperationID:   "upgrade-" + string(event.Phase),
			Label:         "Upgrade ycy",
			Phases:        []terminalexperience.PhaseDefinition{{ID: string(event.Phase), Name: name}},
			Updates:       track.updates,
			RequestCancel: sink.cancel,
		})
	}()
	track.updates <- terminalexperience.OperationPhase{ID: string(event.Phase), State: terminalexperience.PhaseActive}
}

func (sink *upgradePhaseSink) previousState(state updater.UpdateTransaction) {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	if sink.presentationErr != nil {
		return
	}
	document := terminalUpgradeDocument(StateMessage(state), terminalUpgradeStateRole(state))
	sink.previousResult = &document
	if sink.capabilities.Interaction == terminalexperience.RichInteractive && sink.current == nil {
		sink.recordLocked(sink.run.Milestone(document))
	}
}

func (sink *upgradePhaseSink) close() {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	if sink.current == nil {
		sink.pendingConsume = nil
		return
	}
	close(sink.current.updates)
	sink.recordLocked(<-sink.current.done)
	sink.current = nil
}

func (sink *upgradePhaseSink) previousDocument() *terminalexperience.PresentationDocument {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	if sink.previousResult == nil {
		return nil
	}
	document := *sink.previousResult
	document.Blocks = append([]terminalexperience.PresentationBlock(nil), document.Blocks...)
	return &document
}

func (sink *upgradePhaseSink) err() error {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	return sink.presentationErr
}

func (sink *upgradePhaseSink) recordLocked(err error) {
	if err == nil {
		return
	}
	sink.presentationErr = errors.Join(sink.presentationErr, err)
	if sink.cancel != nil {
		sink.cancel()
	}
}

func finishUpgradeRun(run terminalexperience.ExperienceRun, diagnostics io.Writer, previous *terminalexperience.PresentationDocument, result updater.UpgradeResult, resultErr error) error {
	if resultErr != nil {
		// Cancellation is a process lifecycle outcome, even when the lower-level
		// updater preserves its historical ExitCodeError wrapper.
		if errors.Is(resultErr, context.Canceled) || errors.Is(resultErr, context.DeadlineExceeded) {
			return errors.Join(resultErr, run.Finish(terminalexperience.Cancelled, previous))
		}
		var exit *updater.ExitCodeError
		if result.Aborted && errors.As(resultErr, &exit) {
			_, _ = fmt.Fprintln(diagnostics, "error: "+terminalUpgradeDiagnostic(resultErr))
			document := terminalUpgradeDocument("Update aborted.", terminalexperience.VisualRoleWarning)
			return errors.Join(resultErr, run.Finish(terminalexperience.Failed, terminalUpgradeCombinedDocument(previous, &document)))
		}
		outcome := terminalexperience.Failed
		if errors.Is(resultErr, context.Canceled) || errors.Is(resultErr, context.DeadlineExceeded) {
			outcome = terminalexperience.Cancelled
		}
		return errors.Join(resultErr, run.Finish(outcome, previous))
	}
	if result.AlreadyCurrent {
		document := terminalUpgradeDocument(fmt.Sprintf("Current version v%s is the latest.\nNo update needed.", result.CurrentVersion), terminalexperience.VisualRoleSuccess)
		return run.Finish(terminalexperience.Succeeded, terminalUpgradeCombinedDocument(previous, &document))
	}
	if result.Scheduled {
		document := terminalUpgradeDocument(fmt.Sprintf("Update to v%s has been scheduled and will finish after ycy exits.", result.ScheduledVersion), terminalexperience.VisualRoleSuccess)
		return run.Finish(terminalexperience.Succeeded, terminalUpgradeCombinedDocument(previous, &document))
	}
	return run.Finish(terminalexperience.Succeeded, previous)
}

func terminalUpgradeCombinedDocument(first, second *terminalexperience.PresentationDocument) *terminalexperience.PresentationDocument {
	if first == nil && second == nil {
		return nil
	}
	document := terminalexperience.PresentationDocument{}
	if first != nil {
		document.Blocks = append(document.Blocks, first.Blocks...)
	}
	if second != nil {
		document.Blocks = append(document.Blocks, second.Blocks...)
	}
	return &document
}

func terminalUpgradeIntroDocument() terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / upgrade"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Upgrade ycy"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Resolve and verify the latest release before scheduling the updater"},
	}}
}

func terminalUpgradePhaseName(phase updater.UpgradePhase) (string, bool) {
	switch phase {
	case updater.UpgradePhaseConsumeStartupTransaction:
		return "Consume startup transaction", true
	case updater.UpgradePhaseResolveRelease:
		return "Resolve release", true
	case updater.UpgradePhaseResolveArtifact:
		return "Resolve artifact", true
	case updater.UpgradePhaseDownloadCandidate:
		return "Download candidate", true
	case updater.UpgradePhaseVerifyCandidate:
		return "Verify candidate", true
	case updater.UpgradePhaseStageUpdater:
		return "Stage updater", true
	case updater.UpgradePhasePublishPending:
		return "Publish pending update", true
	case updater.UpgradePhaseScheduleUpdater:
		return "Schedule updater", true
	case updater.UpgradePhaseComplete:
		return "Complete", true
	default:
		return "", false
	}
}

func terminalUpgradePhaseState(state updater.UpgradePhaseState) terminalexperience.PhaseState {
	switch state {
	case updater.UpgradePhaseCompleted:
		return terminalexperience.PhaseCompleted
	case updater.UpgradePhaseCancelled:
		return terminalexperience.PhaseCancelled
	case updater.UpgradePhaseFailed:
		return terminalexperience.PhaseFailed
	default:
		return terminalexperience.PhaseActive
	}
}

func terminalUpgradePhaseDetail(event updater.UpgradePhaseEvent) string {
	parts := []string{event.Detail}
	switch event.Phase {
	case updater.UpgradePhaseResolveRelease:
		if event.CurrentVersion != "" && event.CandidateVersion != "" {
			parts = append(parts, "Current v"+event.CurrentVersion+"; latest v"+event.CandidateVersion)
		}
		if event.TargetOS != "" && event.TargetArchitecture != "" {
			parts = append(parts, event.TargetOS+"/"+event.TargetArchitecture)
		}
	case updater.UpgradePhaseResolveArtifact:
		if event.ArtifactName != "" {
			parts = append(parts, event.ArtifactName)
		}
		if event.ChecksumSource != "" {
			parts = append(parts, "checksum: "+event.ChecksumSource)
		}
	case updater.UpgradePhaseComplete:
		if event.CandidateVersion != "" {
			parts = append(parts, "Target v"+event.CandidateVersion)
		}
	}
	return terminalUpgradeSafeDetail(strings.Join(nonEmptyUpgradeFields(parts), " | "))
}

func nonEmptyUpgradeFields(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

func terminalUpgradeSafeDetail(value string) string {
	value = terminalexperience.RenderPlain(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Text: value}}})
	value = strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(value)
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= 240 {
		return value
	}
	return string([]rune(value)[:237]) + "..."
}

func terminalUpgradeDiagnostic(err error) string {
	return terminalUpgradeSafeDetail(logging.Redact(err.Error()))
}
