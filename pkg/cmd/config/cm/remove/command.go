package remove

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type StoreProvider func() (Reader, RemoveWriter, error)

type Options struct {
	Context  context.Context
	Profile  string
	Store    StoreProvider
	Terminal *terminalexperience.Runtime
}

func NewCmdRemove(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runRemove
	}
	return &cobra.Command{Use: "remove <profile>", Short: "Remove a commit message provider profile", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
			return errors.New("config cm remove Factory is incomplete")
		}
		return runF(&Options{Context: command.Context(), Profile: arguments[0], Store: func() (Reader, RemoveWriter, error) {
			store, err := factory.ConfigStore()
			if err != nil {
				return nil, nil, err
			}
			return store, store, nil
		}, Terminal: factory.Terminal})
	}}
}

func runRemove(options *Options) error {
	_, err := executeRemove(options)
	return err
}

func executeRemove(options *Options) (RemoveResult, error) {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return RemoveResult{}, errors.New("config cm remove options are incomplete")
	}
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	run, err := options.Terminal.OpenConsole(ctx, terminalCMRemoveConsoleDescriptor(options.Profile))
	if err != nil {
		return RemoveResult{}, err
	}
	defer run.Close()
	caps := options.Terminal.Capabilities()
	if err := ctx.Err(); err != nil {
		return RemoveResult{}, errors.Join(err, run.Finish(terminalexperience.Cancelled, nil))
	}
	if caps.Interaction == terminalexperience.RichInteractive {
		if err := run.Notice(terminalCMRemoveIntroDocument()); err != nil {
			return RemoveResult{}, errors.Join(err, run.Finish(terminalexperience.Failed, nil))
		}
	}
	adapter := newTerminalCMRemoveAdapter(run)
	phases := newCMRemovePhaseSink(run, caps)
	phases.beginValidation()
	reader, writer, workErr := options.Store()
	if workErr != nil {
		phases.endValidation(terminalexperience.PhaseFailed, "Unable to validate CM profile")
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	if reader == nil {
		workErr = errors.New("config cm remove reader is nil")
		phases.endValidation(terminalexperience.PhaseFailed, "Unable to validate CM profile")
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	if writer == nil {
		workErr = errors.New("config cm remove writer is nil")
		phases.endValidation(terminalexperience.PhaseFailed, "Unable to validate CM profile")
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	profiles, workErr := reader.ListCMProfiles()
	if workErr != nil {
		phases.endValidation(terminalexperience.PhaseFailed, "Unable to validate CM profile")
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	if err := ctx.Err(); err != nil {
		phases.endValidation(terminalexperience.PhaseCancelled, "CM profile validation cancelled")
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Cancelled, nil, err)
	}
	var target *appconfig.CMProfile
	for index := range profiles.Profiles {
		if profiles.Profiles[index].Name == options.Profile {
			target = &profiles.Profiles[index]
			break
		}
	}
	if target == nil {
		workErr = fmt.Errorf("CM profile not found: %s", options.Profile)
		phases.endValidation(terminalexperience.PhaseFailed, "Unable to validate CM profile")
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	role := "Configured profile"
	if profiles.DefaultProfile == options.Profile {
		role = "Current default"
	}
	phases.endValidation(terminalexperience.PhaseCompleted, "Profile: "+safeCMRemoveName(options.Profile)+"; Role: "+role)
	if err := ctx.Err(); err != nil {
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Cancelled, nil, err)
	}
	if caps.Interaction == terminalexperience.Automation {
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, errConfigCMRemoveRequiresInteractive)
	}
	question := RemoveConfirmPrompt{Message: fmt.Sprintf("Remove CM profile \"%s\"?", safeCMRemoveName(options.Profile))}
	if role == "Current default" {
		question.Description = "Removing the default selects the first remaining stored profile, or clears the default when none remain."
	}
	confirmed, cancelled, workErr := adapter.Confirm(question)
	if workErr != nil {
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	if cancelled {
		if caps.Interaction == terminalexperience.RichInteractive {
			_ = run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "Confirmation cancelled"}}})
		}
		document := terminalCMRemoveDocument("Cancelled", true)
		return RemoveResult{Cancelled: true}, finishCMRemove(run, phases, terminalexperience.Cancelled, &document, nil)
	}
	if !confirmed {
		if caps.Interaction == terminalexperience.RichInteractive {
			_ = run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: "Removal declined"}}})
		}
		document := terminalCMRemoveDocument("Cancelled", true)
		return RemoveResult{Declined: true}, finishCMRemove(run, phases, terminalexperience.Cancelled, &document, nil)
	}
	if err := ctx.Err(); err != nil {
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Cancelled, nil, err)
	}
	if caps.Interaction == terminalexperience.RichInteractive {
		_ = run.Milestone(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleWarning, Text: fmt.Sprintf("Remove CM profile \"%s\": confirmed", safeCMRemoveName(options.Profile))}}})
	}
	phases.beginRemoval()
	removed, workErr := writer.RemoveCMProfile(options.Profile)
	if workErr != nil {
		phases.endRemoval(terminalexperience.PhaseFailed, "Unable to remove CM profile")
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	if !removed {
		workErr = fmt.Errorf("CM profile not found: %s", options.Profile)
		phases.endRemoval(terminalexperience.PhaseFailed, "Unable to remove CM profile")
		return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Failed, nil, workErr)
	}
	phases.endRemoval(terminalexperience.PhaseCompleted, "Profile removed")
	document := terminalCMRemoveDocument(fmt.Sprintf("Profile %s removed", safeCMRemoveName(options.Profile)), false)
	if caps.Interaction == terminalexperience.RichInteractive && caps.Stdout.Terminal {
		document = terminalCMRemoveSuccessDocument(options.Profile)
	}
	return RemoveResult{}, finishCMRemove(run, phases, terminalexperience.Succeeded, &document, nil)
}

var _ Reader = (*appconfig.Store)(nil)
var _ RemoveWriter = (*appconfig.Store)(nil)

func terminalCMRemoveConsoleDescriptor(profile string) terminalexperience.ConsoleDescriptor {
	return terminalexperience.ConsoleDescriptor{
		Command: "YCY / config cm remove",
		Target:  "commit message profile removal",
		Status:  "READY",
		Metadata: []terminalexperience.ConsoleMetadata{
			{Label: "scope", Value: "commit message configuration"},
			{Label: "profile", Value: safeCMRemoveName(profile)},
		},
	}
}

type cmRemovePhaseSink struct {
	run      terminalexperience.ExperienceRun
	caps     terminalexperience.Capabilities
	updates  chan terminalexperience.OperationPhase
	done     chan error
	closed   bool
	trackErr error
}

func newCMRemovePhaseSink(run terminalexperience.ExperienceRun, caps terminalexperience.Capabilities) *cmRemovePhaseSink {
	return &cmRemovePhaseSink{run: run, caps: caps}
}

func (sink *cmRemovePhaseSink) beginValidation() {
	if sink.caps.Interaction == terminalexperience.PlainInteractive {
		_ = sink.run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleActive, Text: "Checking CM profile..."}}})
		return
	}
	if sink.caps.Interaction != terminalexperience.RichInteractive {
		return
	}
	sink.start()
	sink.updates <- terminalexperience.OperationPhase{ID: cmRemoveValidationPhaseID, State: terminalexperience.PhaseActive, Detail: "Checking profile"}
}

func (sink *cmRemovePhaseSink) endValidation(state terminalexperience.PhaseState, detail string) {
	if sink.caps.Interaction == terminalexperience.RichInteractive {
		sink.updates <- terminalexperience.OperationPhase{ID: cmRemoveValidationPhaseID, State: state, Detail: detail}
		close(sink.updates)
		sink.closed = true
		sink.trackErr = errors.Join(sink.trackErr, sink.wait())
	}
}

func (sink *cmRemovePhaseSink) beginRemoval() {
	if sink.caps.Interaction == terminalexperience.PlainInteractive {
		_ = sink.run.Notice(terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleActive, Text: "Removing CM profile..."}}})
		return
	}
	if sink.caps.Interaction == terminalexperience.RichInteractive {
		sink.start()
		sink.updates <- terminalexperience.OperationPhase{ID: cmRemovePhaseID, State: terminalexperience.PhaseActive, Detail: "Deleting stored profile"}
	}
}

func (sink *cmRemovePhaseSink) endRemoval(state terminalexperience.PhaseState, detail string) {
	if sink.caps.Interaction != terminalexperience.RichInteractive {
		return
	}
	sink.updates <- terminalexperience.OperationPhase{ID: cmRemovePhaseID, State: state, Detail: detail}
	close(sink.updates)
	sink.closed = true
	sink.trackErr = errors.Join(sink.trackErr, sink.wait())
}

func (sink *cmRemovePhaseSink) closeWithoutRemoval() {
	if sink.caps.Interaction != terminalexperience.RichInteractive || sink.closed {
		return
	}
	close(sink.updates)
	sink.closed = true
	sink.trackErr = errors.Join(sink.trackErr, sink.wait())
}

func (sink *cmRemovePhaseSink) start() {
	sink.updates = make(chan terminalexperience.OperationPhase, 8)
	sink.done = make(chan error, 1)
	sink.closed = false
	go func() {
		sink.done <- sink.run.Track(terminalexperience.TrackedOperation{
			ID:    "config-cm-remove",
			Label: "Remove CM profile",
			Phases: []terminalexperience.PhaseDefinition{
				{ID: cmRemoveValidationPhaseID, Name: cmRemoveValidationPhaseName},
				{ID: cmRemovePhaseID, Name: cmRemovePhaseName},
			},
			Updates: sink.updates,
		})
	}()
}

func (sink *cmRemovePhaseSink) wait() error {
	if sink.done == nil {
		return nil
	}
	return <-sink.done
}

func finishCMRemove(run terminalexperience.ExperienceRun, sink *cmRemovePhaseSink, outcome terminalexperience.FinishOutcome, document *terminalexperience.PresentationDocument, workErr error) error {
	sink.closeWithoutRemoval()
	return errors.Join(workErr, sink.trackErr, run.Finish(outcome, document))
}

const (
	cmRemoveValidationPhaseID   = "validate-cm-profile"
	cmRemoveValidationPhaseName = "Validate CM profile"
	cmRemovePhaseID             = "remove-cm-profile"
	cmRemovePhaseName           = "Remove CM profile"
)

func terminalCMRemoveIntroDocument() terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / config cm remove"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Remove CM profile"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Delete one stored commit message provider"},
	}}
}

func terminalCMRemoveSuccessDocument(name string) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / config cm remove"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Remove CM profile"},
		{Role: terminalexperience.VisualRoleSuccess, Text: "Profile " + safeCMRemoveName(name) + " removed"},
	}}
}

func safeCMRemoveName(value string) string {
	if !utf8.ValidString(value) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return "Profile configured"
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "Profile configured"
	}
	runes := []rune(value)
	if len(runes) > 256 {
		return string(runes[:256]) + "..."
	}
	return value
}
