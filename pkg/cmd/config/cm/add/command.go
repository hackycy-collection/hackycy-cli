package add

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type StoreProvider func() (AddWriter, error)

type Options struct {
	Context  context.Context
	Store    StoreProvider
	Terminal *terminal.Runtime
}

func NewCmdAdd(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runAdd
	}
	return &cobra.Command{
		Use:   "add",
		Short: "Add a CM profile",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
				return errors.New("config cm add Factory is incomplete")
			}
			return runF(&Options{Context: command.Context(), Store: func() (AddWriter, error) {
				store, err := factory.ConfigStore()
				if err != nil {
					return nil, err
				}
				return store, nil
			}, Terminal: factory.Terminal})
		},
	}
}

func runAdd(options *Options) error {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return errors.New("config cm add options are incomplete")
	}
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	run, err := options.Terminal.OpenConsole(ctx, terminalCMAddConsoleDescriptor())
	if err != nil {
		return err
	}
	defer run.Close()
	caps := options.Terminal.Capabilities()
	if err := ctx.Err(); err != nil {
		return errors.Join(err, run.Finish(terminal.Cancelled, nil))
	}
	if caps.Interaction == terminal.Automation {
		return errors.Join(errConfigCMAddRequiresInteractive, run.Finish(terminal.Failed, nil))
	}
	if caps.Interaction == terminal.RichInteractive {
		if err := run.Notice(terminalCMAddIntroDocument()); err != nil {
			return errors.Join(err, run.Finish(terminal.Failed, nil))
		}
	}
	adapter := newTerminalCMAddAdapter(run)
	phases := newCMAddPhaseSink(run, caps)
	phases.beginCollect()
	writer, workErr := options.Store()
	if workErr == nil && writer == nil {
		workErr = errors.New("config cm add store is nil")
	}
	if workErr == nil {
		var input AddInput
		input, cancelled, promptErr := PromptAdd(adapter)
		if promptErr != nil {
			workErr = promptErr
			phases.endCollect(terminal.PhaseFailed, "Unable to collect CM profile details")
		} else if cancelled {
			phases.endCollect(terminal.PhaseCancelled, "Profile setup cancelled")
			document := terminalCMAddDocument("Cancelled", true)
			return finishCMAdd(run, phases, terminal.Cancelled, &document, nil)
		} else {
			if err := ctx.Err(); err != nil {
				phases.endCollect(terminal.PhaseCancelled, "Profile setup cancelled")
				document := terminalCMAddDocument("Cancelled", true)
				return finishCMAdd(run, phases, terminal.Cancelled, &document, err)
			}
			phases.endCollect(terminal.PhaseCompleted, cmAddCollectDetail(input))
			phases.beginSave()
			workErr = SaveAdd(writer, input)
			if workErr != nil {
				phases.endSave(terminal.PhaseFailed, "Unable to save CM profile")
			} else {
				phases.endSave(terminal.PhaseCompleted, "Profile saved")
				document := terminalCMAddDocument(fmt.Sprintf("Profile %s added", safeCMAddName(input.Name)), false)
				if caps.Interaction == terminal.RichInteractive && caps.Stdout.Terminal {
					document = terminalCMAddSuccessDocument(input)
				}
				return finishCMAdd(run, phases, terminal.Succeeded, &document, nil)
			}
		}
	} else {
		phases.endCollect(terminal.PhaseFailed, "Unable to collect CM profile details")
	}
	return finishCMAdd(run, phases, terminal.Failed, nil, workErr)
}

var errConfigCMAddRequiresInteractive = errors.New("config cm add requires an interactive terminal")

func terminalCMAddConsoleDescriptor() terminal.ConsoleDescriptor {
	return terminal.ConsoleDescriptor{
		Command: "YCY / config cm add",
		Target:  "commit message profile setup",
		Status:  "READY",
		Metadata: []terminal.ConsoleMetadata{{
			Label: "scope",
			Value: "commit message configuration",
		}},
	}
}

var _ AddWriter = (*appconfig.Store)(nil)

type cmAddPhaseSink struct {
	run           terminal.ExperienceRun
	caps          terminal.Capabilities
	updates       chan terminal.OperationPhase
	done          chan error
	closed        bool
	tracking      bool
	trackErr      error
	collectState  terminal.PhaseState
	collectDetail string
}

func newCMAddPhaseSink(run terminal.ExperienceRun, caps terminal.Capabilities) *cmAddPhaseSink {
	return &cmAddPhaseSink{run: run, caps: caps}
}

func (sink *cmAddPhaseSink) beginCollect() {
	if sink.caps.Interaction == terminal.PlainInteractive {
		_ = sink.run.Notice(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Role: terminal.VisualRoleActive, Text: "Collecting CM profile details..."}}})
		return
	}
	if sink.caps.Interaction != terminal.RichInteractive {
		return
	}
	// Ask and Track share the Experience operation lock. Keep the collect
	// state in memory while the form is active, then replay it once prompting
	// has finished and the save phase is ready to run.
	sink.collectState = terminal.PhaseActive
	sink.collectDetail = "Answer the four profile fields"
}

func (sink *cmAddPhaseSink) endCollect(state terminal.PhaseState, detail string) {
	if sink.caps.Interaction == terminal.RichInteractive {
		sink.collectState = state
		sink.collectDetail = detail
		return
	}
}

func (sink *cmAddPhaseSink) beginSave() {
	if sink.caps.Interaction == terminal.PlainInteractive {
		_ = sink.run.Notice(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Role: terminal.VisualRoleActive, Text: "Saving CM profile..."}}})
		return
	}
	if sink.caps.Interaction == terminal.RichInteractive {
		sink.start()
		sink.updates <- terminal.OperationPhase{ID: cmAddSavePhaseID, State: terminal.PhaseActive, Detail: "Writing encrypted profile"}
	}
}

func (sink *cmAddPhaseSink) endSave(state terminal.PhaseState, detail string) {
	if sink.caps.Interaction != terminal.RichInteractive {
		return
	}
	sink.updates <- terminal.OperationPhase{ID: cmAddSavePhaseID, State: state, Detail: detail}
	close(sink.updates)
	sink.closed = true
	sink.trackErr = sink.wait()
}

func (sink *cmAddPhaseSink) finishWithoutSave() {
	if sink.caps.Interaction != terminal.RichInteractive {
		return
	}
	if sink.closed {
		return
	}
	sink.start()
	close(sink.updates)
	sink.closed = true
	sink.trackErr = sink.wait()
}

func (sink *cmAddPhaseSink) start() {
	if sink.tracking {
		return
	}
	sink.updates = make(chan terminal.OperationPhase, 8)
	sink.done = make(chan error, 1)
	sink.tracking = true
	go func() {
		sink.done <- sink.run.Track(terminal.TrackedOperation{
			ID:    "config-cm-add",
			Label: "Add commit message profile",
			Phases: []terminal.PhaseDefinition{
				{ID: cmAddCollectPhaseID, Name: cmAddCollectPhaseName},
				{ID: cmAddSavePhaseID, Name: cmAddSavePhaseName},
			},
			Updates: sink.updates,
		})
	}()
	if sink.collectState == terminal.PhasePending {
		sink.collectState = terminal.PhaseActive
		sink.collectDetail = "Answer the four profile fields"
	}
	sink.updates <- terminal.OperationPhase{ID: cmAddCollectPhaseID, State: terminal.PhaseActive, Detail: "Answer the four profile fields"}
	if sink.collectState != terminal.PhaseActive {
		sink.updates <- terminal.OperationPhase{ID: cmAddCollectPhaseID, State: sink.collectState, Detail: sink.collectDetail}
	}
}

func (sink *cmAddPhaseSink) wait() error {
	if sink.done == nil {
		return nil
	}
	return <-sink.done
}

func finishCMAdd(run terminal.ExperienceRun, sink *cmAddPhaseSink, outcome terminal.FinishOutcome, document *terminal.PresentationDocument, workErr error) error {
	if outcome == terminal.Cancelled || outcome == terminal.Failed {
		sink.finishWithoutSave()
	}
	return errors.Join(workErr, sink.trackErr, run.Finish(outcome, document))
}

const (
	cmAddCollectPhaseID   = "collect-cm-profile-details"
	cmAddCollectPhaseName = "Collect CM profile details"
	cmAddSavePhaseID      = "save-cm-profile"
	cmAddSavePhaseName    = "Save CM profile"
)

func terminalCMAddIntroDocument() terminal.PresentationDocument {
	return terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{
		{Role: terminal.VisualRoleMuted, Text: "YCY / config cm add"},
		{Role: terminal.VisualRoleTitle, Text: "Add commit message profile"},
		{Role: terminal.VisualRoleMuted, Text: "Configure an OpenAI-compatible provider"},
	}}
}

func terminalCMAddSuccessDocument(input AddInput) terminal.PresentationDocument {
	return terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{
		{Role: terminal.VisualRoleMuted, Text: "YCY / config cm add"},
		{Role: terminal.VisualRoleTitle, Text: "Add commit message profile"},
		{Role: terminal.VisualRoleSuccess, Text: "Profile " + safeCMAddName(input.Name) + " added"},
	}}
}

func cmAddCollectDetail(input AddInput) string {
	return fmt.Sprintf("Profile: %s; Base URL: %s; Model: %s; API key: [redacted]", safeCMAddName(input.Name), safeCMAddURL(input.BaseURL), safeCMAddModel(input.Model))
}

func safeCMAddName(value string) string { return safeCMAddField(value, "Profile configured") }

func safeCMAddModel(value string) string { return safeCMAddField(value, "Model configured") }

func safeCMAddField(value, fallback string) string {
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

func safeCMAddURL(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !utf8.ValidString(trimmed) || strings.IndexFunc(trimmed, unicode.IsControl) >= 0 {
		return "Base URL configured"
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "Base URL configured"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return safeCMAddField(parsed.Scheme+"://"+parsed.Host+parsed.EscapedPath(), "Base URL configured")
}
