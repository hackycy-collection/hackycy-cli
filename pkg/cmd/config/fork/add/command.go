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

// StoreProvider resolves the appconfig writer at command execution time.
type StoreProvider func() (AddWriter, error)

// Options contains the parsed add request and leaf-owned capabilities.
type Options struct {
	Context  context.Context
	Store    StoreProvider
	Terminal *terminal.Runtime
}

// NewCmdAdd creates the config fork add command with an optional test runner.
func NewCmdAdd(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runAdd
	}
	return &cobra.Command{
		Use:   "add",
		Short: "Add a provider instance",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
				return errors.New("config fork add Factory is incomplete")
			}
			return runF(&Options{
				Context: command.Context(),
				Store: func() (AddWriter, error) {
					store, err := factory.ConfigStore()
					if err != nil {
						return nil, err
					}
					return store, nil
				},
				Terminal: factory.Terminal,
			})
		},
	}
}

func runAdd(options *Options) error {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return errors.New("config fork add options are incomplete")
	}
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	run, err := options.Terminal.OpenConsole(ctx, terminalForkAddConsoleDescriptor())
	if err != nil {
		return err
	}
	defer run.Close()
	caps := options.Terminal.Capabilities()
	if err := ctx.Err(); err != nil {
		return errors.Join(err, run.Finish(terminal.Cancelled, nil))
	}
	if caps.Interaction == terminal.Automation {
		return errors.Join(errConfigForkAddRequiresInteractive, run.Finish(terminal.Failed, nil))
	}
	if caps.Interaction == terminal.RichInteractive {
		if err := run.Notice(terminalForkAddIntroDocument()); err != nil {
			return errors.Join(err, run.Finish(terminal.Failed, nil))
		}
	}
	adapter := newTerminalForkAddAdapter(run)
	phases := newForkAddPhaseSink(run, caps)
	phases.beginCollect()
	writer, workErr := options.Store()
	if workErr == nil && writer == nil {
		workErr = errors.New("config fork add store is nil")
	}
	if workErr == nil {
		input, cancelled, promptErr := PromptAdd(adapter)
		if promptErr != nil {
			workErr = promptErr
			phases.endCollect(terminal.PhaseFailed, "Unable to collect provider details")
		} else if cancelled {
			phases.endCollect(terminal.PhaseCancelled, "Provider setup cancelled")
			document := terminalForkAddDocument("Cancelled", true)
			return finishForkAdd(run, phases, terminal.Cancelled, &document, nil)
		} else {
			if err := ctx.Err(); err != nil {
				phases.endCollect(terminal.PhaseCancelled, "Provider setup cancelled")
				document := terminalForkAddDocument("Cancelled", true)
				return finishForkAdd(run, phases, terminal.Cancelled, &document, err)
			}
			phases.endCollect(terminal.PhaseCompleted, forkAddCollectDetail(input))
			phases.beginSave()
			workErr = SaveAdd(writer, input)
			if workErr != nil {
				phases.endSave(terminal.PhaseFailed, "Unable to save provider instance")
			} else {
				phases.endSave(terminal.PhaseCompleted, "Provider instance saved")
				document := terminalForkAddDocument(fmt.Sprintf("Instance %s (%s) added successfully", safeForkAddField(input.Alias, "Instance configured"), safeForkAddHost(input.Host)), false)
				if caps.Interaction == terminal.RichInteractive && caps.Stdout.Terminal {
					document = terminalForkAddSuccessDocument(input)
				}
				return finishForkAdd(run, phases, terminal.Succeeded, &document, nil)
			}
		}
	} else {
		phases.endCollect(terminal.PhaseFailed, "Unable to collect provider details")
	}
	return finishForkAdd(run, phases, terminal.Failed, nil, workErr)
}

var errConfigForkAddRequiresInteractive = errors.New("config fork add requires an interactive terminal")

func terminalForkAddConsoleDescriptor() terminal.ConsoleDescriptor {
	return terminal.ConsoleDescriptor{
		Command: "YCY / config fork add",
		Target:  "provider connection setup",
		Status:  "READY",
		Metadata: []terminal.ConsoleMetadata{{
			Label: "scope",
			Value: "git fork configuration",
		}},
	}
}

var _ AddWriter = (*appconfig.Store)(nil)

type forkAddPhaseSink struct {
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

func newForkAddPhaseSink(run terminal.ExperienceRun, caps terminal.Capabilities) *forkAddPhaseSink {
	return &forkAddPhaseSink{run: run, caps: caps}
}

func (sink *forkAddPhaseSink) beginCollect() {
	if sink.caps.Interaction == terminal.PlainInteractive {
		_ = sink.run.Notice(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Role: terminal.VisualRoleActive, Text: "Collecting provider details..."}}})
		return
	}
	if sink.caps.Interaction != terminal.RichInteractive {
		return
	}
	// Ask and Track serialize on the same Experience operation lock. Record
	// the form phase until all five fields are complete, then replay it when the
	// save phase starts.
	sink.collectState = terminal.PhaseActive
	sink.collectDetail = "Answer the five provider fields"
}

func (sink *forkAddPhaseSink) endCollect(state terminal.PhaseState, detail string) {
	if sink.caps.Interaction == terminal.RichInteractive {
		sink.collectState = state
		sink.collectDetail = detail
	}
}

func (sink *forkAddPhaseSink) beginSave() {
	if sink.caps.Interaction == terminal.PlainInteractive {
		_ = sink.run.Notice(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Role: terminal.VisualRoleActive, Text: "Saving provider instance..."}}})
		return
	}
	if sink.caps.Interaction == terminal.RichInteractive {
		sink.start()
		sink.updates <- terminal.OperationPhase{ID: forkAddSavePhaseID, State: terminal.PhaseActive, Detail: "Writing encrypted provider configuration"}
	}
}

func (sink *forkAddPhaseSink) endSave(state terminal.PhaseState, detail string) {
	if sink.caps.Interaction == terminal.PlainInteractive {
		if state == terminal.PhaseCompleted {
			_ = sink.run.Notice(terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{{Role: terminal.VisualRoleSuccess, Text: "Saved provider instance"}}})
		}
		return
	}
	if sink.caps.Interaction != terminal.RichInteractive {
		return
	}
	sink.updates <- terminal.OperationPhase{ID: forkAddSavePhaseID, State: state, Detail: detail}
	close(sink.updates)
	sink.closed = true
	sink.trackErr = sink.wait()
}

func (sink *forkAddPhaseSink) finishWithoutSave() {
	if sink.caps.Interaction != terminal.RichInteractive || sink.closed {
		return
	}
	sink.start()
	close(sink.updates)
	sink.closed = true
	sink.trackErr = sink.wait()
}

func (sink *forkAddPhaseSink) start() {
	if sink.tracking {
		return
	}
	sink.updates = make(chan terminal.OperationPhase, 8)
	sink.done = make(chan error, 1)
	sink.tracking = true
	go func() {
		sink.done <- sink.run.Track(terminal.TrackedOperation{
			ID:    "config-fork-add",
			Label: "Add fork provider instance",
			Phases: []terminal.PhaseDefinition{
				{ID: forkAddCollectPhaseID, Name: forkAddCollectPhaseName},
				{ID: forkAddSavePhaseID, Name: forkAddSavePhaseName},
			},
			Updates: sink.updates,
		})
	}()
	if sink.collectState == terminal.PhasePending {
		sink.collectState = terminal.PhaseActive
		sink.collectDetail = "Answer the five provider fields"
	}
	sink.updates <- terminal.OperationPhase{ID: forkAddCollectPhaseID, State: terminal.PhaseActive, Detail: "Answer the five provider fields"}
	if sink.collectState != terminal.PhaseActive {
		sink.updates <- terminal.OperationPhase{ID: forkAddCollectPhaseID, State: sink.collectState, Detail: sink.collectDetail}
	}
}

func (sink *forkAddPhaseSink) wait() error {
	if sink.done == nil {
		return nil
	}
	return <-sink.done
}

func finishForkAdd(run terminal.ExperienceRun, sink *forkAddPhaseSink, outcome terminal.FinishOutcome, document *terminal.PresentationDocument, workErr error) error {
	if outcome == terminal.Cancelled || outcome == terminal.Failed {
		sink.finishWithoutSave()
	}
	return errors.Join(workErr, sink.trackErr, run.Finish(outcome, document))
}

const (
	forkAddCollectPhaseID   = "collect-provider-details"
	forkAddCollectPhaseName = "Collect provider details"
	forkAddSavePhaseID      = "save-provider-instance"
	forkAddSavePhaseName    = "Save provider instance"
)

func terminalForkAddIntroDocument() terminal.PresentationDocument {
	return terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{
		{Role: terminal.VisualRoleMuted, Text: "YCY / config fork add"},
		{Role: terminal.VisualRoleTitle, Text: "Add fork provider instance"},
		{Role: terminal.VisualRoleMuted, Text: "Store a provider connection for git fork operations"},
	}}
}

func terminalForkAddSuccessDocument(input AddInput) terminal.PresentationDocument {
	return terminal.PresentationDocument{Blocks: []terminal.PresentationBlock{
		{Role: terminal.VisualRoleMuted, Text: "YCY / config fork add"},
		{Role: terminal.VisualRoleTitle, Text: "Add fork provider instance"},
		{Role: terminal.VisualRoleSuccess, Text: "Instance " + safeForkAddField(input.Alias, "Instance configured") + " (" + safeForkAddHost(input.Host) + ") added successfully"},
	}}
}

func forkAddCollectDetail(input AddInput) string {
	return fmt.Sprintf("Instance: %s; Host: %s; Provider: %s; Protocol: %s; Access token: [redacted]", safeForkAddField(input.Alias, "Instance configured"), safeForkAddHost(input.Host), forkAddProviderLabel(input.Type), forkAddProtocolLabel(input.Scheme))
}

func forkAddProviderLabel(value string) string {
	for _, choice := range providerChoices {
		if choice.Value == value {
			return choice.Label
		}
	}
	return "Provider configured"
}

func forkAddProtocolLabel(value string) string {
	for _, choice := range protocolChoices {
		if choice.Value == value {
			return choice.Label
		}
	}
	return "Protocol configured"
}

func safeForkAddHost(value string) string {
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
	result := parsed.Host + parsed.EscapedPath()
	return safeForkAddField(result, "Host configured")
}

func safeForkAddField(value, fallback string) string {
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
