package terminal

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"
)

var errRichUnavailable = errors.New("rich terminal is unavailable")

type richController struct {
	runtime *Runtime
	model   *richRootModel
	program *tea.Program
	lease   *RendererLease

	done chan struct{}
	mu   sync.Mutex
	err  error
	next uint64
}

func newRichController(runtime *Runtime) *richController {
	return &richController{runtime: runtime, done: make(chan struct{})}
}

func (controller *richController) start() error {
	if controller.runtime.inputTerminal == nil || controller.runtime.diagnosticTerminal == nil {
		return errRichUnavailable
	}
	width, height, err := term.GetSize(int(controller.runtime.diagnosticTerminal.Fd()))
	if err != nil || width <= 0 || height <= 0 {
		return errRichUnavailable
	}

	controller.lease = controller.runtime.diagnostics.AcquireRendererLease()
	controller.model = newRichRootModel(
		width,
		height,
		controller.runtime.capabilities.Stderr.Color,
	)
	controller.program = tea.NewProgram(
		controller.model,
		tea.WithInput(controller.runtime.inputTerminal),
		// Keep the root's writes inside the renderer lease.  This makes the
		// renderer and the semantic replay share one serialized terminal owner.
		tea.WithOutput(rendererTerminalWriter{
			writer:   controller.lease.Writer(),
			terminal: controller.runtime.diagnosticTerminal,
		}),
		tea.WithoutSignalHandler(),
	)
	go func() {
		_, runErr := controller.program.Run()
		controller.mu.Lock()
		controller.err = runErr
		controller.mu.Unlock()
		close(controller.done)
	}()

	ack := make(chan struct{})
	if err := controller.send(richReadyMsg{ack: ack}, ack); err != nil {
		controller.program.Quit()
		<-controller.done
		return errors.Join(err, controller.programErrorOrNil(), controller.releaseLease())
	}
	return nil
}

// rendererTerminalWriter preserves Bubble Tea's terminal capability detection
// while routing every renderer write through the active diagnostic lease.
type rendererTerminalWriter struct {
	writer   io.Writer
	terminal *os.File
}

func (writer rendererTerminalWriter) Write(value []byte) (int, error) {
	return writer.writer.Write(value)
}

func (writer rendererTerminalWriter) Read(value []byte) (int, error) {
	return writer.terminal.Read(value)
}

// Close deliberately leaves the inherited diagnostic terminal open. Bubble Tea
// only needs the terminal-file shape for capability detection and sizing.
func (rendererTerminalWriter) Close() error {
	return nil
}

func (writer rendererTerminalWriter) Fd() uintptr {
	return writer.terminal.Fd()
}

func (controller *richController) ask(ctx context.Context, handler *InteractionHandler, request InteractionRequest) (InteractionAnswer, error) {
	if err := validateInteractionRequest(request); err != nil {
		return InteractionAnswer{}, err
	}
	controller.next++
	id := controller.next
	form, answer, err := newRichForm(handler, request, id)
	if err != nil {
		return InteractionAnswer{}, err
	}

	response := make(chan richAskResult, 1)
	ack := make(chan struct{})
	if err := controller.send(richShowFormMsg{
		id:       id,
		form:     form,
		answer:   answer,
		response: response,
		ack:      ack,
	}, ack); err != nil {
		return InteractionAnswer{}, err
	}

	select {
	case result := <-response:
		return result.answer, result.err
	case <-ctx.Done():
		cancelAck := make(chan struct{})
		_ = controller.send(richCancelFormMsg{id: id, err: ctx.Err(), ack: cancelAck}, cancelAck)
		return InteractionAnswer{}, ctx.Err()
	case <-controller.done:
		return InteractionAnswer{}, controller.programError()
	}
}

func (controller *richController) notice(document PresentationDocument) error {
	ack := make(chan struct{})
	return controller.send(richNoticeMsg{document: document, ack: ack}, ack)
}

func (controller *richController) milestone(document PresentationDocument) error {
	ack := make(chan struct{})
	return controller.send(richMilestoneMsg{document: document, ack: ack}, ack)
}

func (controller *richController) startTrack(label string, phases []OperationPhase, requestCancel func() error) error {
	ack := make(chan struct{})
	return controller.send(richStartTrackMsg{label: label, phases: phases, requestCancel: requestCancel, ack: ack}, ack)
}

func (controller *richController) updateTrack(phase OperationPhase) error {
	ack := make(chan struct{})
	return controller.send(richTrackPhaseMsg{phase: phase, ack: ack}, ack)
}

func (controller *richController) cancelTrack() error {
	ack := make(chan struct{})
	return controller.send(richCancelTrackMsg{ack: ack}, ack)
}

func (controller *richController) finishTrack() error {
	ack := make(chan struct{})
	return controller.send(richFinishTrackMsg{ack: ack}, ack)
}

func (controller *richController) send(message tea.Msg, ack <-chan struct{}) error {
	select {
	case <-controller.done:
		return controller.programError()
	default:
	}
	controller.program.Send(message)
	select {
	case <-ack:
		return nil
	case <-controller.done:
		return controller.programError()
	}
}

func (controller *richController) close(ledger *TranscriptLedger) error {
	return controller.closeWith(ledger, true)
}

// closeAfterFailure performs the same terminal restoration and replay as a
// normal close, but leaves the already-returned renderer error to the caller so
// it is not reported twice.
func (controller *richController) closeAfterFailure(ledger *TranscriptLedger) error {
	return controller.closeWith(ledger, false)
}

func (controller *richController) closeWith(ledger *TranscriptLedger, includeProgramError bool) error {
	if controller.program != nil {
		controller.program.Quit()
		<-controller.done
	}

	// Bubble Tea has restored the primary screen at this point, while the
	// lease still owns stderr.  Replay only the bounded semantic ledger; raw
	// frames and command results are deliberately not copied here.
	var replayErr error
	if ledger != nil {
		if transcript := ledger.Render(); transcript != "" && controller.lease != nil {
			_, replayErr = io.WriteString(controller.lease.Writer(), transcript)
		}
	}
	var programErr error
	if includeProgramError {
		programErr = controller.programErrorOrNil()
	}
	return errors.Join(programErr, replayErr, controller.releaseLease())
}

func (controller *richController) releaseLease() error {
	if controller.lease == nil {
		return nil
	}
	lease := controller.lease
	controller.lease = nil
	return lease.Close()
}

func (controller *richController) programError() error {
	if err := controller.programErrorOrNil(); err != nil {
		return err
	}
	return io.ErrUnexpectedEOF
}

func (controller *richController) programErrorOrNil() error {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.err
}

func (controller *richController) stopped() bool {
	if controller == nil || controller.done == nil {
		return false
	}
	select {
	case <-controller.done:
		return true
	default:
		return false
	}
}

type richMode uint8

const (
	richNoticeMode richMode = iota
	richFormMode
	richTrackMode
)

type richRootModel struct {
	width  int
	height int
	color  bool
	mode   richMode

	notices  []PresentationDocument
	formID   uint64
	form     richFormModel
	answer   func() InteractionAnswer
	response chan<- richAskResult
	track    *trackedState
}

func newRichRootModel(width, height int, color bool) *richRootModel {
	return &richRootModel{
		width:  width,
		height: height,
		color:  color,
	}
}

func (*richRootModel) Init() tea.Cmd {
	return nil
}

func (model *richRootModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch value := message.(type) {
	case richReadyMsg:
		close(value.ack)
		return model, nil
	case tea.WindowSizeMsg:
		model.width = max(value.Width, 1)
		model.height = max(value.Height, 1)
		if model.form != nil {
			model.configureForm()
			updated, cmd := model.form.Update(tea.WindowSizeMsg{Width: model.width, Height: model.formHeight()})
			model.form = updated.(richFormModel)
			return model, cmd
		}
		return model, nil
	case richNoticeMsg:
		model.preserveTrack()
		if len(value.document.Blocks) > 0 {
			model.notices = append(model.notices, value.document)
		}
		model.mode = richNoticeMode
		close(value.ack)
		return model, nil
	case richMilestoneMsg:
		model.preserveTrack()
		if len(value.document.Blocks) > 0 {
			model.notices = append(model.notices, value.document)
		}
		model.mode = richNoticeMode
		close(value.ack)
		return model, nil
	case richShowFormMsg:
		model.preserveTrack()
		model.mode = richFormMode
		model.formID = value.id
		model.form = value.form
		model.answer = value.answer
		model.response = value.response
		model.configureForm()
		close(value.ack)
		return model, model.form.Init()
	case richFormSubmittedMsg:
		if value.id == model.formID && model.form != nil {
			model.response <- richAskResult{answer: model.answer()}
			model.clearForm()
		}
		return model, nil
	case richFormCancelledMsg:
		if value.id == model.formID && model.form != nil {
			model.response <- richAskResult{err: ErrInteractionCancelled}
			model.clearForm()
		}
		return model, nil
	case richCancelFormMsg:
		if value.id == model.formID && model.form != nil {
			model.response <- richAskResult{err: value.err}
			model.clearForm()
		}
		close(value.ack)
		return model, nil
	case richStartTrackMsg:
		model.preserveTrack()
		model.mode = richTrackMode
		model.track = &trackedState{
			label:  value.label,
			phases: append([]OperationPhase(nil), value.phases...),
			requestStop: func() {
				_ = value.requestCancel()
			},
		}
		close(value.ack)
		return model, nil
	case richTrackPhaseMsg:
		if model.track != nil {
			model.track.applyPhase(value.phase)
		}
		close(value.ack)
		return model, nil
	case richCancelTrackMsg:
		if model.track != nil {
			model.track.requestCancellation()
		}
		close(value.ack)
		return model, nil
	case richFinishTrackMsg:
		if model.track != nil {
			model.track.cancelArmed = false
		}
		close(value.ack)
		return model, nil
	case tea.KeyPressMsg:
		if model.mode == richTrackMode && model.track != nil {
			switch value.String() {
			case "ctrl+c":
				model.track.requestCancellation()
			case "esc":
				if model.track.cancelArmed {
					model.track.requestCancellation()
				} else {
					model.track.cancelArmed = true
				}
			default:
				model.track.cancelArmed = false
			}
			return model, nil
		}
		if model.mode == richFormMode && model.form != nil && value.String() == "esc" {
			if model.form.handlesEscape() {
				updated, command := model.form.Update(message)
				model.form = updated.(richFormModel)
				return model, command
			}
			model.response <- richAskResult{err: ErrInteractionCancelled}
			model.clearForm()
			return model, nil
		}
	}

	if model.mode == richFormMode && model.form != nil {
		updated, cmd := model.form.Update(message)
		model.form = updated.(richFormModel)
		return model, cmd
	}
	return model, nil
}

func (model *richRootModel) View() tea.View {
	if model.width <= 0 || model.height <= 0 {
		return tea.View{AltScreen: true, DisableBracketedPasteMode: true}
	}
	parts := make([]string, 0, 2)
	if notices := model.noticeView(); notices != "" {
		parts = append(parts, notices)
	}

	activeHeight := model.formHeight()
	var active string
	switch model.mode {
	case richFormMode:
		if model.form != nil {
			active = model.form.View().Content
		}
	case richTrackMode:
		if model.track != nil {
			active = model.track.view(model.width, richStyles(model.color))
		}
	}
	if active != "" {
		parts = append(parts, takeFirstLines(active, activeHeight))
	}
	return tea.View{
		Content:                   takeFirstLines(strings.Join(parts, "\n"), model.height),
		AltScreen:                 true,
		DisableBracketedPasteMode: true,
	}
}

func (model *richRootModel) configureForm() {
	if model.form == nil {
		return
	}
	model.form.configure(max(model.width, 1), model.formHeight(), !model.compact())
}

func (model *richRootModel) compact() bool {
	return model.width < 32 || model.height < 8
}

func (model *richRootModel) formHeight() int {
	return max(model.height-model.noticeHeight(), 1)
}

func (model *richRootModel) noticeHeight() int {
	if model.compact() || len(model.notices) == 0 {
		return 0
	}
	return min(lineCount(model.renderNotices()), max(model.height/3, 1))
}

func (model *richRootModel) noticeView() string {
	height := model.noticeHeight()
	if height == 0 {
		return ""
	}
	return takeLastLines(model.renderNotices(), height)
}

func (model *richRootModel) renderNotices() string {
	parts := make([]string, 0, len(model.notices))
	for _, document := range model.notices {
		rendered := strings.TrimSuffix(renderRich(document, RichOptions{
			Width: model.width,
			Color: model.color,
		}), "\n")
		if rendered != "" {
			parts = append(parts, rendered)
		}
	}
	return strings.Join(parts, "\n")
}

func (model *richRootModel) preserveTrack() {
	if model.track != nil {
		document := model.track.finalDocument()
		if len(document.Blocks) > 0 && document.Blocks[0].Text != "" {
			model.notices = append(model.notices, document)
		}
	}
	model.track = nil
}

func (model *richRootModel) clearForm() {
	model.mode = richNoticeMode
	model.formID = 0
	model.form = nil
	model.answer = nil
	model.response = nil
}

func lineCount(value string) int {
	if value == "" {
		return 0
	}
	return len(strings.Split(value, "\n"))
}

func takeFirstLines(value string, count int) string {
	if count <= 0 || value == "" {
		return ""
	}
	lines := strings.Split(value, "\n")
	if len(lines) > count {
		lines = lines[:count]
	}
	return strings.Join(lines, "\n")
}

func takeLastLines(value string, count int) string {
	if count <= 0 || value == "" {
		return ""
	}
	lines := strings.Split(value, "\n")
	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}
	return strings.Join(lines, "\n")
}

type richReadyMsg struct{ ack chan struct{} }

type richNoticeMsg struct {
	document PresentationDocument
	ack      chan struct{}
}

type richMilestoneMsg struct {
	document PresentationDocument
	ack      chan struct{}
}

type richAskResult struct {
	answer InteractionAnswer
	err    error
}

type richShowFormMsg struct {
	id       uint64
	form     richFormModel
	answer   func() InteractionAnswer
	response chan<- richAskResult
	ack      chan struct{}
}

type richFormSubmittedMsg struct{ id uint64 }
type richFormCancelledMsg struct{ id uint64 }

type richCancelFormMsg struct {
	id  uint64
	err error
	ack chan struct{}
}

type richStartTrackMsg struct {
	label         string
	phases        []OperationPhase
	requestCancel func() error
	ack           chan struct{}
}

type richTrackPhaseMsg struct {
	phase OperationPhase
	ack   chan struct{}
}

type richCancelTrackMsg struct{ ack chan struct{} }
type richFinishTrackMsg struct{ ack chan struct{} }
