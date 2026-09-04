package env

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

var errExportEnvRequiresInteractive = errors.New("export env requires an interactive terminal")

func runEnv(options *Options) error {
	if options == nil || options.WorkingDirectory == nil || options.Terminal == nil || options.Reader == nil || options.Writer == nil {
		return errors.New("export env options are incomplete")
	}
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	run, err := options.Terminal.OpenConsole(ctx, terminalExportEnvConsoleDescriptor(options))
	if err != nil {
		return err
	}
	defer run.Close()
	caps := options.Terminal.Capabilities()
	adapter := newTerminalExportEnvAdapter(run, caps.Interaction == terminalexperience.Automation)
	if caps.Interaction == terminalexperience.RichInteractive {
		if err := run.Notice(terminalExportEnvIntroDocument(options)); err != nil {
			return errors.Join(err, run.Finish(terminalexperience.Failed, nil))
		}
	}
	sink := newExportEnvPhaseSink(run, caps, options.Output != "")
	module, err := New(Dependencies{
		WorkingDirectory: options.WorkingDirectory,
		Selector:         adapter,
		Reader:           options.Reader,
		Writer:           options.Writer,
		Presenter:        adapter,
	})
	if err != nil {
		return err
	}
	observer := &runObserver{}
	observer.phase = sink.phase
	observer.selected = sink.selected
	observer.variables = sink.variables
	result, err := module.run(ctx, Input{
		Directory:   options.Directory,
		Environment: options.Environment,
		Merge:       options.Merge,
		Output:      options.Output,
	}, observer)
	sink.close()
	if sink.err != nil {
		err = errors.Join(err, sink.err)
	}
	if result.Cancelled {
		document := terminalExportEnvResultDocument("Cancelled", terminalexperience.VisualRoleWarning)
		return errors.Join(err, run.Finish(terminalexperience.Cancelled, &document))
	}
	if err != nil {
		return errors.Join(err, run.Finish(terminalexperience.Failed, nil))
	}
	if options.Output != "" {
		if sink.err != nil {
			return errors.Join(sink.err, run.Finish(terminalexperience.Failed, nil))
		}
		document := terminalExportEnvResultDocument("Wrote output to "+safeExportTarget(options.Output), terminalexperience.VisualRoleSuccess)
		return run.Finish(terminalexperience.Succeeded, &document)
	}
	// The JSON is intentionally a separate plain block so Rich styling never
	// inserts symbols or alters its durable structure.
	contents := observer.output
	if sink.err != nil {
		return errors.Join(sink.err, run.Finish(terminalexperience.Failed, nil))
	}
	document := terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleSuccess, Text: "Exported variables:"},
		{Role: terminalexperience.VisualRolePlain, Text: contents},
	}}
	return run.Finish(terminalexperience.Succeeded, &document)
}

func terminalExportEnvConsoleDescriptor(options *Options) terminalexperience.ConsoleDescriptor {
	directory := "."
	merge := "off"
	if options != nil {
		directory = options.Directory
		if strings.TrimSpace(directory) == "" {
			directory = "."
		}
		if options.Merge {
			merge = "on"
		}
	}
	return terminalexperience.ConsoleDescriptor{
		Command: "YCY / export env",
		Target:  "environment JSON",
		Status:  "READY",
		Metadata: []terminalexperience.ConsoleMetadata{
			{Label: "directory", Value: safeExportText(directory)},
			{Label: "merge base .env", Value: merge},
		},
	}
}

type exportEnvPhaseSink struct {
	run              terminalexperience.ExperienceRun
	caps             terminalexperience.Capabilities
	withOutput       bool
	current          *exportEnvTrack
	cluster          string
	pendingVariables *terminalexperience.PresentationDocument
	err              error
}

type exportEnvTrack struct {
	updates chan terminalexperience.OperationPhase
	done    chan error
}

func newExportEnvPhaseSink(run terminalexperience.ExperienceRun, caps terminalexperience.Capabilities, withOutput bool) *exportEnvPhaseSink {
	return &exportEnvPhaseSink{run: run, caps: caps, withOutput: withOutput}
}

func (sink *exportEnvPhaseSink) phase(id, name string, state terminalPhaseState, detail string) {
	if sink.err != nil {
		return
	}
	if id == "select-environment" {
		if state == terminalPhaseActive {
			sink.closeTrack()
			sink.notice(name, detail, terminalexperience.VisualRoleActive)
			return
		}
		role := terminalexperience.VisualRoleSuccess
		if state == terminalPhaseCancelled {
			role = terminalexperience.VisualRoleWarning
		} else if state == terminalPhaseFailed {
			role = terminalexperience.VisualRoleError
		}
		sink.milestone(terminalExportEnvPhaseDocument(name, detail, state, role))
		return
	}
	cluster := "discovery"
	if id == "read-selected-files" || id == "parse-and-merge-values" || id == "encode-json" || id == "write-output-file" {
		cluster = "output"
	}
	if sink.current != nil && sink.cluster != cluster {
		sink.closeTrack()
	}
	if state == terminalPhaseActive && sink.current == nil {
		sink.startTrack(id)
	}
	if sink.current == nil {
		return
	}
	sink.cluster = cluster
	update := terminalexperience.OperationPhase{ID: id, State: exportEnvPhaseState(state), Detail: detail}
	sink.current.updates <- update
	if state != terminalPhaseActive && sink.isTrackEnd(id) {
		sink.closeTrack()
	}
}

func (sink *exportEnvPhaseSink) selected(selection Selection, source string, merge bool) {
	if sink.caps.Interaction != terminalexperience.RichInteractive || sink.err != nil {
		return
	}
	// Selection is the boundary between discovery and output. Close the
	// discovery Track before publishing its milestone so the runtime operation
	// lock is not re-entered while the Track is still consuming updates.
	sink.closeTrack()
	if sink.err != nil {
		return
	}
	sink.milestone(terminalExportEnvSelectionDocument(selection, source, merge))
}

func (sink *exportEnvPhaseSink) variables(count int) {
	if sink.caps.Interaction != terminalexperience.RichInteractive || sink.err != nil {
		return
	}
	// The observer reports the variable count while the output Track is still
	// consuming its parse phase. Defer the milestone until that Track closes so
	// the runtime operation lock is never re-entered from an active Track.
	document := terminalExportEnvVariableDocument(count)
	sink.pendingVariables = &document
}

func (sink *exportEnvPhaseSink) startTrack(firstID string) {
	definitions := []terminalexperience.PhaseDefinition{
		{ID: "resolve-directory", Name: "Resolve directory"},
		{ID: "discover-environment-files", Name: "Discover environment files"},
	}
	operationID := "export-env-discovery"
	if firstID == "read-selected-files" {
		definitions = []terminalexperience.PhaseDefinition{
			{ID: "read-selected-files", Name: "Read selected files"},
			{ID: "parse-and-merge-values", Name: "Parse and merge values"},
			{ID: "encode-json", Name: "Encode JSON"},
		}
		if sink.withOutput {
			definitions = append(definitions, terminalexperience.PhaseDefinition{ID: "write-output-file", Name: "Write output file"})
		}
		operationID = "export-env-output"
	}
	updates := make(chan terminalexperience.OperationPhase, len(definitions)+4)
	done := make(chan error, 1)
	sink.current = &exportEnvTrack{updates: updates, done: done}
	sink.cluster = map[bool]string{false: "discovery", true: "output"}[firstID == "read-selected-files"]
	go func() {
		done <- sink.run.Track(terminalexperience.TrackedOperation{
			ID:      operationID,
			Label:   "Export environment",
			Phases:  definitions,
			Updates: updates,
		})
	}()
}

func (sink *exportEnvPhaseSink) isTrackEnd(id string) bool {
	if sink.withOutput {
		return id == "write-output-file"
	}
	return id == "encode-json"
}

func (sink *exportEnvPhaseSink) closeTrack() {
	if sink.current == nil {
		return
	}
	close(sink.current.updates)
	err := <-sink.current.done
	sink.err = errors.Join(sink.err, err)
	sink.current = nil
	if sink.pendingVariables != nil && sink.err == nil {
		document := *sink.pendingVariables
		sink.pendingVariables = nil
		sink.err = errors.Join(sink.err, sink.run.Milestone(document))
	}
}

func (sink *exportEnvPhaseSink) notice(name, detail string, role terminalexperience.VisualRole) {
	if sink.caps.Interaction == terminalexperience.Automation {
		return
	}
	document := terminalExportEnvPhaseDocument(name, detail, terminalPhaseActive, role)
	sink.err = errors.Join(sink.err, sink.run.Notice(document))
}

func (sink *exportEnvPhaseSink) milestone(document terminalexperience.PresentationDocument) {
	if sink.caps.Interaction == terminalexperience.Automation {
		return
	}
	sink.err = errors.Join(sink.err, sink.run.Milestone(document))
}

func (sink *exportEnvPhaseSink) close() {
	sink.closeTrack()
}

func exportEnvPhaseState(state terminalPhaseState) terminalexperience.PhaseState {
	switch state {
	case terminalPhaseSucceeded:
		return terminalexperience.PhaseCompleted
	case terminalPhaseCancelled:
		return terminalexperience.PhaseCancelled
	case terminalPhaseFailed:
		return terminalexperience.PhaseFailed
	default:
		return terminalexperience.PhaseActive
	}
}

func terminalExportEnvIntroDocument(options *Options) terminalexperience.PresentationDocument {
	merge := "off"
	if options.Merge {
		merge = "on"
	}
	directory := strings.TrimSpace(options.Directory)
	if directory == "" {
		directory = "."
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "YCY / export env"},
		{Role: terminalexperience.VisualRoleTitle, Text: "Export environment"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Export .env file contents as JSON"},
		{Role: terminalexperience.VisualRolePlain, Text: "Directory: " + safeExportText(directory) + "  Merge base .env: " + merge},
	}}
}

func terminalExportEnvPhaseDocument(name, detail string, state terminalPhaseState, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	text := name
	if detail != "" {
		text += ": " + detail
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func terminalExportEnvResultDocument(text string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func terminalExportEnvSelectionDocument(selection Selection, source string, merge bool) terminalexperience.PresentationDocument {
	files := make([]string, 0, len(selection.Files))
	for _, file := range selection.Files {
		files = append(files, safeExportText(filepath.ToSlash(file)))
	}
	mergeValue := "off"
	if merge {
		mergeValue = "on"
	}
	selected := "environment"
	if len(selection.Files) > 0 {
		selected = environmentLabel(selection.Files[len(selection.Files)-1])
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleSuccess, Text: "Selected environment: " + safeExportText(selected) + " (" + safeExportText(source) + ")"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Files: " + strings.Join(files, ", ") + "  Merge: " + mergeValue},
	}}
}

func terminalExportEnvVariableDocument(count int) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleSuccess,
		Text: fmt.Sprintf("Exported %d variable%s", count, pluralSuffix(count)),
	}}}
}

func safeExportTarget(value string) string {
	value = safeExportText(value)
	if filepath.IsAbs(value) {
		return filepath.Base(filepath.Clean(value))
	}
	return value
}

func safeExportText(value string) string {
	if !utf8.ValidString(value) {
		return "configured path"
	}
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsControl(r) {
			return "configured path"
		}
		builder.WriteRune(r)
	}
	value = strings.TrimSpace(builder.String())
	if value == "" {
		return "."
	}
	runes := []rune(value)
	if len(runes) > 160 {
		return fmt.Sprintf("%s...", string(runes[:160]))
	}
	return value
}

type terminalExportEnvAdapter struct {
	run        terminalexperience.ExperienceRun
	automation bool
}

func newTerminalExportEnvAdapter(run terminalexperience.ExperienceRun, automation bool) *terminalExportEnvAdapter {
	return &terminalExportEnvAdapter{run: run, automation: automation}
}

func (adapter *terminalExportEnvAdapter) SelectEnvironment(message string, choices []EnvironmentChoice) (string, bool, error) {
	if adapter.automation && len(choices) == 1 {
		return choices[0].Value, false, nil
	}
	answer, err := adapter.run.Ask(terminalexperience.InteractionRequest{
		Kind:            terminalexperience.InteractionSelect,
		Message:         message,
		Options:         exportEnvInteractionOptions(choices),
		CancelValues:    []string{"", "q", "quit", "cancel"},
		TranscriptLabel: "Selected environment",
	})
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) || errors.Is(err, context.Canceled) {
		return "", true, nil
	}
	if errors.Is(err, terminalexperience.ErrAutomationInteraction) {
		return "", false, errExportEnvRequiresInteractive
	}
	if err != nil {
		return "", false, err
	}
	return answer.Value, false, nil
}

func (adapter *terminalExportEnvAdapter) Outro(message string) {
	_ = adapter.run.Result(terminalExportEnvDocument(message, terminalexperience.VisualRoleMuted))
}

func (adapter *terminalExportEnvAdapter) Print(value string) {
	_ = adapter.run.Result(terminalExportEnvDocument(value, terminalexperience.VisualRolePlain))
}

func (adapter *terminalExportEnvAdapter) Cancel(message string) {
	_ = adapter.run.Result(terminalExportEnvDocument(message, terminalexperience.VisualRoleWarning))
}

func exportEnvInteractionOptions(choices []EnvironmentChoice) []terminalexperience.InteractionOption {
	options := make([]terminalexperience.InteractionOption, 0, len(choices))
	for _, choice := range choices {
		options = append(options, terminalexperience.InteractionOption{Label: choice.Label, Value: choice.Value, Description: choice.Value})
	}
	return options
}

func terminalExportEnvDocument(text string, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}
