package remove

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// StoreProvider resolves the shared appconfig store only when removal runs.
type StoreProvider func() (RemoveReader, RemoveWriter, error)

// Options contains the parsed remove request and leaf-owned capabilities.
type Options struct {
	Context  context.Context
	Store    StoreProvider
	Terminal *terminalexperience.Runtime
}

// NewCmdRemove creates the config fork remove command with an optional test runner.
func NewCmdRemove(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runRemove
	}
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove a provider instance",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
				return errors.New("config fork remove Factory is incomplete")
			}
			return runF(&Options{
				Context: command.Context(),
				Store: func() (RemoveReader, RemoveWriter, error) {
					store, err := factory.ConfigStore()
					if err != nil {
						return nil, nil, err
					}
					return store, store, nil
				},
				Terminal: factory.Terminal,
			})
		},
	}
}

func runRemove(options *Options) error {
	_, err := executeRemove(options)
	return err
}

func executeRemove(options *Options) (RemoveResult, error) {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return RemoveResult{}, errors.New("config fork remove options are incomplete")
	}
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	run := options.Terminal.Open(ctx)
	defer run.Close()
	caps := options.Terminal.Capabilities()
	if err := ctx.Err(); err != nil {
		return RemoveResult{}, errors.Join(err, run.Finish(terminalexperience.Cancelled, nil))
	}
	if caps.Interaction == terminalexperience.RichInteractive {
		if err := run.Notice(terminalForkRemoveIntroDocument()); err != nil {
			return RemoveResult{}, errors.Join(err, run.Finish(terminalexperience.Failed, nil))
		}
	}
	adapter := newTerminalForkRemoveAdapter(run)
	phases := newForkRemovePhaseSink(run, caps)
	phases.beginLoad()
	reader, writer, workErr := options.Store()
	if workErr != nil {
		return RemoveResult{}, finishForkRemoveLoadError(run, phases, workErr)
	}
	if reader == nil {
		return RemoveResult{}, finishForkRemoveLoadError(run, phases, errors.New("config fork remove reader is nil"))
	}
	if writer == nil {
		return RemoveResult{}, finishForkRemoveLoadError(run, phases, errors.New("config fork remove writer is nil"))
	}
	if err := ctx.Err(); err != nil {
		phases.endLoad(terminalexperience.PhaseCancelled, "Loading fork provider instances cancelled")
		return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Cancelled, nil, err)
	}
	instances, workErr := reader.ListForkInstances()
	if workErr != nil {
		return RemoveResult{}, finishForkRemoveLoadError(run, phases, workErr)
	}
	if err := ctx.Err(); err != nil {
		phases.endLoad(terminalexperience.PhaseCancelled, "Loading fork provider instances cancelled")
		return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Cancelled, nil, err)
	}
	if len(instances) == 0 {
		phases.endLoad(terminalexperience.PhaseCompleted, "No instances configured")
		if caps.Interaction == terminalexperience.RichInteractive {
			if err := run.Milestone(terminalForkRemoveDocument("No instances configured", terminalexperience.VisualRoleMuted)); err != nil {
				return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, err)
			}
		} else if caps.Interaction == terminalexperience.PlainInteractive {
			if err := run.Notice(terminalForkRemoveDocument("No instances configured", terminalexperience.VisualRoleMuted)); err != nil {
				return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, err)
			}
		}
		document := terminalForkRemoveDocument("Nothing to remove", terminalexperience.VisualRoleSuccess)
		return RemoveResult{Empty: true}, finishForkRemove(run, phases, terminalexperience.Succeeded, &document, nil)
	}
	phases.endLoad(terminalexperience.PhaseCompleted, fmt.Sprintf("Loaded %d provider instance%s", len(instances), pluralForkRemove(len(instances))))
	if caps.Interaction == terminalexperience.Automation {
		return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, errConfigForkRemoveRequiresInteractive)
	}
	choices := make([]Choice, len(instances))
	for index, instance := range instances {
		choices[index] = Choice{Value: instance.Name, Label: safeForkRemoveName(instance.Name), Description: safeForkRemoveHost(instance.Host)}
	}
	selected, cancelled, workErr := adapter.Select(SelectPrompt{Message: "Select instance to remove", Choices: choices})
	if workErr != nil {
		return RemoveResult{}, finishForkRemoveInteractionError(run, phases, workErr)
	}
	if cancelled {
		if caps.Interaction == terminalexperience.RichInteractive {
			if err := run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "Selection cancelled"}}}); err != nil {
				return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, err)
			}
		}
		document := terminalForkRemoveDocument("Cancelled", terminalexperience.VisualRoleWarning)
		return RemoveResult{Cancelled: true}, finishForkRemove(run, phases, terminalexperience.Cancelled, &document, nil)
	}
	var selectedInstance appconfig.ForkInstance
	found := false
	for _, instance := range instances {
		if instance.Name == selected {
			selectedInstance = instance
			found = true
			break
		}
	}
	if !found {
		return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, errors.New("selected fork instance is no longer configured"))
	}
	if caps.Interaction == terminalexperience.RichInteractive {
		if err := run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleMuted, Text: "Host: " + safeForkRemoveHost(selectedInstance.Host)}}}); err != nil {
			return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, err)
		}
	}
	question := ConfirmPrompt{Message: fmt.Sprintf("Remove instance \"%s\"?", safeForkRemoveName(selected))}
	confirmed, cancelled, workErr := adapter.Confirm(question)
	if workErr != nil {
		return RemoveResult{}, finishForkRemoveInteractionError(run, phases, workErr)
	}
	if cancelled {
		if caps.Interaction == terminalexperience.RichInteractive {
			if err := run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "Confirmation cancelled"}}}); err != nil {
				return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, err)
			}
		}
		document := terminalForkRemoveDocument("Cancelled", terminalexperience.VisualRoleWarning)
		return RemoveResult{Cancelled: true}, finishForkRemove(run, phases, terminalexperience.Cancelled, &document, nil)
	}
	if !confirmed {
		if caps.Interaction == terminalexperience.RichInteractive {
			if err := run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "Removal declined"}}}); err != nil {
				return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, err)
			}
		}
		document := terminalForkRemoveDocument("Cancelled", terminalexperience.VisualRoleWarning)
		return RemoveResult{Declined: true}, finishForkRemove(run, phases, terminalexperience.Cancelled, &document, nil)
	}
	if err := ctx.Err(); err != nil {
		return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Cancelled, nil, err)
	}
	if caps.Interaction == terminalexperience.RichInteractive {
		if err := run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: fmt.Sprintf("Remove instance \"%s\": confirmed", safeForkRemoveName(selected))}}}); err != nil {
			return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, err)
		}
	}
	phases.beginRemoval()
	_, workErr = writer.RemoveForkInstance(selected)
	if workErr != nil {
		phases.endRemoval(terminalexperience.PhaseFailed, "Unable to remove provider instance")
		return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	phases.endRemoval(terminalexperience.PhaseCompleted, "Provider instance removed")
	document := terminalForkRemoveDocument(forkRemoveSuccessMessage(selected), terminalexperience.VisualRoleSuccess)
	if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
		document = terminalForkRemoveSuccessDocument(selected)
	}
	return RemoveResult{}, finishForkRemove(run, phases, terminalexperience.Succeeded, &document, nil)
}

var _ RemoveReader = (*appconfig.Store)(nil)
var _ RemoveWriter = (*appconfig.Store)(nil)

type forkRemovePhaseSink struct {
	run      terminalexperience.ExperienceRun
	caps     terminalexperience.Capabilities
	updates  chan terminalexperience.OperationPhase
	done     chan error
	closed   bool
	trackErr error
}

func newForkRemovePhaseSink(run terminalexperience.ExperienceRun, caps terminalexperience.Capabilities) *forkRemovePhaseSink {
	return &forkRemovePhaseSink{run: run, caps: caps}
}

func (sink *forkRemovePhaseSink) beginLoad() {
	if sink.caps.Interaction == terminalexperience.PlainInteractive {
		_ = sink.run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleActive, Text: "Loading fork provider instances..."}}})
		return
	}
	if sink.caps.Interaction != terminalexperience.RichInteractive {
		return
	}
	sink.start()
	sink.updates <- terminalexperience.OperationPhase{ID: forkRemoveLoadPhaseID, State: terminalexperience.PhaseActive, Detail: "Reading provider configuration"}
}

func (sink *forkRemovePhaseSink) endLoad(state terminalexperience.PhaseState, detail string) {
	if sink.caps.Interaction == terminalexperience.PlainInteractive {
		if state == terminalexperience.PhaseCompleted && detail != "No instances configured" {
			_ = sink.run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleSuccess, Text: "Loaded fork provider instances"}}})
		}
		return
	}
	if sink.caps.Interaction != terminalexperience.RichInteractive {
		return
	}
	sink.updates <- terminalexperience.OperationPhase{ID: forkRemoveLoadPhaseID, State: state, Detail: detail}
	close(sink.updates)
	sink.closed = true
	sink.trackErr = errors.Join(sink.trackErr, sink.wait())
}

func (sink *forkRemovePhaseSink) beginRemoval() {
	if sink.caps.Interaction == terminalexperience.PlainInteractive {
		_ = sink.run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleActive, Text: "Removing provider instance..."}}})
		return
	}
	if sink.caps.Interaction == terminalexperience.RichInteractive {
		sink.start()
		sink.updates <- terminalexperience.OperationPhase{ID: forkRemovePhaseID, State: terminalexperience.PhaseActive, Detail: "Deleting stored provider instance"}
	}
}

func (sink *forkRemovePhaseSink) endRemoval(state terminalexperience.PhaseState, detail string) {
	if sink.caps.Interaction != terminalexperience.RichInteractive {
		return
	}
	sink.updates <- terminalexperience.OperationPhase{ID: forkRemovePhaseID, State: state, Detail: detail}
	close(sink.updates)
	sink.closed = true
	sink.trackErr = errors.Join(sink.trackErr, sink.wait())
}

func (sink *forkRemovePhaseSink) closeWithoutRemoval() {
	if sink.caps.Interaction != terminalexperience.RichInteractive || sink.closed {
		return
	}
	close(sink.updates)
	sink.closed = true
	sink.trackErr = errors.Join(sink.trackErr, sink.wait())
}

func (sink *forkRemovePhaseSink) start() {
	sink.updates = make(chan terminalexperience.OperationPhase, 8)
	sink.done = make(chan error, 1)
	sink.closed = false
	go func() {
		sink.done <- sink.run.Track(terminalexperience.TrackedOperation{
			ID:    "config-fork-remove",
			Label: "Remove fork provider instance",
			Phases: []terminalexperience.PhaseDefinition{
				{ID: forkRemoveLoadPhaseID, Name: forkRemoveLoadPhaseName},
				{ID: forkRemovePhaseID, Name: forkRemovePhaseName},
			},
			Updates: sink.updates,
		})
	}()
}

func (sink *forkRemovePhaseSink) wait() error {
	if sink.done == nil {
		return nil
	}
	return <-sink.done
}

func finishForkRemove(run terminalexperience.ExperienceRun, sink *forkRemovePhaseSink, outcome terminalexperience.FinishOutcome, document *terminalexperience.PresentationDocument, workErr error) error {
	sink.closeWithoutRemoval()
	return errors.Join(workErr, sink.trackErr, run.Finish(outcome, document))
}

func finishForkRemoveLoadError(run terminalexperience.ExperienceRun, sink *forkRemovePhaseSink, workErr error) error {
	if forkRemoveContextCancelled(workErr) {
		sink.endLoad(terminalexperience.PhaseCancelled, "Loading fork provider instances cancelled")
		return finishForkRemove(run, sink, terminalexperience.Cancelled, nil, workErr)
	}
	sink.endLoad(terminalexperience.PhaseFailed, "Unable to load fork provider instances")
	return finishForkRemove(run, sink, terminalexperience.Failed, nil, workErr)
}

func finishForkRemoveInteractionError(run terminalexperience.ExperienceRun, sink *forkRemovePhaseSink, workErr error) error {
	outcome := terminalexperience.Failed
	if forkRemoveContextCancelled(workErr) {
		outcome = terminalexperience.Cancelled
	}
	return finishForkRemove(run, sink, outcome, nil, workErr)
}

func forkRemoveContextCancelled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

const (
	forkRemoveLoadPhaseID   = "load-fork-provider-instances"
	forkRemoveLoadPhaseName = "Load fork provider instances"
	forkRemovePhaseID       = "remove-provider-instance"
	forkRemovePhaseName     = "Remove provider instance"
)

func terminalForkRemoveIntroDocument() terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / config fork remove"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Remove fork provider instance"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Choose a configured provider connection to remove"},
	}}
}

func terminalForkRemoveSuccessDocument(name string) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / config fork remove"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Remove fork provider instance"},
		{Role: terminalexperience.VisualRoleSuccess, Text: forkRemoveSuccessMessage(name)},
	}}
}

func pluralForkRemove(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func safeForkRemoveName(value string) string {
	return safeForkRemoveField(value, "Selected instance")
}

func safeForkRemoveHost(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !utf8.ValidString(trimmed) || strings.IndexFunc(trimmed, unicode.IsControl) >= 0 {
		return "Host configured"
	}
	parseValue := trimmed
	if !strings.Contains(parseValue, "://") {
		parseValue = "https://" + parseValue
	}
	parsed, err := url.Parse(parseValue)
	if err != nil || parsed.Host == "" {
		return "Host configured"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return safeForkRemoveField(parsed.Host+parsed.EscapedPath(), "Host configured")
}

func forkRemoveSuccessMessage(name string) string {
	projected := safeForkRemoveName(name)
	if projected == "Selected instance" {
		return "Instance removed"
	}
	return "Instance " + projected + " removed"
}

func safeForkRemoveField(value, fallback string) string {
	if !utf8.ValidString(value) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return fallback
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	runes := []rune(value)
	if len(runes) > 256 {
		return string(runes[:256]) + "..."
	}
	return value
}
