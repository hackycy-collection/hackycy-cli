package terminal

import (
	"io"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (run *runtimeRun) trackPlain(output io.Writer, operation TrackedOperation) error {
	return run.consumeTracked(operation, func(phase OperationPhase) error {
		return run.presentPhase(output, phase)
	})
}

func (run *runtimeRun) trackSilently(operation TrackedOperation) error {
	return run.consumeTracked(operation, nil)
}

func (run *runtimeRun) consumeTracked(operation TrackedOperation, render func(OperationPhase) error) error {
	if operation.Updates == nil {
		return nil
	}

	contextDone := run.ctx.Done()
	cancelled := false
	for {
		select {
		case <-contextDone:
			if !cancelled && operation.RequestCancel != nil {
				operation.RequestCancel()
			}
			cancelled = true
			contextDone = nil
		case phase, open := <-operation.Updates:
			if !open {
				return nil
			}
			if render != nil {
				if err := render(phase); err != nil {
					return err
				}
			}
		}
	}
}

func (run *runtimeRun) trackRich(leaseOutput io.Writer, operation TrackedOperation) error {
	if operation.Updates == nil {
		return nil
	}

	rendererOutput := leaseOutput
	if run.runtime.diagnosticTerminal != nil {
		rendererOutput = run.runtime.diagnosticTerminal
	}
	var cancelOnce sync.Once
	requestCancel := func() {
		if operation.RequestCancel != nil {
			cancelOnce.Do(operation.RequestCancel)
		}
	}
	model := newTrackedTeaModel(operation.Label, run.runtime.width, run.runtime.session.Color, rendererOutput, requestCancel)
	program := tea.NewProgram(
		model,
		tea.WithInput(run.runtime.input),
		tea.WithOutput(rendererOutput),
		tea.WithoutSignalHandler(),
		tea.WithoutBracketedPaste(),
	)
	finished := make(chan error, 1)
	go func() {
		_, err := program.Run()
		finished <- err
	}()

	contextDone := run.ctx.Done()
	cancellationRequested := false
	for {
		select {
		case <-contextDone:
			if !cancellationRequested {
				requestCancel()
				program.Send(trackedCancellationMsg{confirmed: true})
			}
			cancellationRequested = true
			contextDone = nil
		case phase, open := <-operation.Updates:
			if !open {
				program.Send(trackedCompleteMsg{})
				err := <-finished
				if err != nil {
					return err
				}
				return WriteRich(rendererOutput, model.finalDocument(), RichOptions{
					Width: run.runtime.width,
					Color: run.runtime.session.Color,
				})
			}
			program.Send(trackedPhaseMsg{phase: phase})
		case err := <-finished:
			return err
		}
	}
}

type trackedPhaseMsg struct {
	phase OperationPhase
}

type trackedCancellationMsg struct {
	confirmed bool
}

type trackedCompleteMsg struct{}

type trackedTeaModel struct {
	label       string
	width       int
	color       bool
	renderer    *lipgloss.Renderer
	requestStop func()

	phases            []OperationPhase
	cancelArmed       bool
	cancellationState bool
}

func newTrackedTeaModel(label string, width int, color bool, output io.Writer, requestStop func()) *trackedTeaModel {
	return &trackedTeaModel{
		label:       label,
		width:       width,
		color:       color,
		renderer:    lipgloss.NewRenderer(output),
		requestStop: requestStop,
	}
}

func (*trackedTeaModel) Init() tea.Cmd {
	return nil
}

func (model *trackedTeaModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch value := message.(type) {
	case tea.WindowSizeMsg:
		model.width = value.Width
	case tea.KeyMsg:
		switch value.String() {
		case "ctrl+c":
			model.requestCancellation()
		case "esc":
			if model.cancelArmed {
				model.requestCancellation()
			} else {
				model.cancelArmed = true
			}
		}
	case trackedPhaseMsg:
		model.applyPhase(value.phase)
	case trackedCancellationMsg:
		if value.confirmed {
			model.cancellationState = true
		}
	case trackedCompleteMsg:
		return model, tea.Quit
	}
	return model, nil
}

func (model *trackedTeaModel) requestCancellation() {
	if model.cancellationState {
		return
	}
	model.cancellationState = true
	if model.requestStop != nil {
		model.requestStop()
	}
}

func (model *trackedTeaModel) applyPhase(phase OperationPhase) {
	for index := len(model.phases) - 1; index >= 0; index-- {
		if model.phases[index].Name == phase.Name {
			model.phases[index] = phase
			return
		}
	}
	model.phases = append(model.phases, phase)
}

func (model *trackedTeaModel) View() string {
	styles := richStyles(model.renderer, model.color)
	lines := make([]string, 0, len(model.phases)*2+3)
	if model.label != "" {
		lines = append(lines, styles[VisualRoleTitle].Render(wrapText(model.label, model.width)))
	}
	if model.width > 0 && model.width < 48 {
		phase := model.currentPhase()
		if phase.Name != "" {
			lines = append(lines, styles[phaseRole(phase.State)].Render(wrapText(phase.Name, model.width)))
			if phase.Detail != "" {
				lines = append(lines, styles[VisualRoleMuted].Render(wrapText(phase.Detail, model.width)))
			}
		}
	} else {
		for _, phase := range model.phases {
			role := phaseRole(phase.State)
			lines = append(lines, styles[role].Render(trackedPhasePrefix(phase.State)+" "+wrapText(phase.Name, model.width)))
			if phase.Detail != "" {
				lines = append(lines, styles[VisualRoleMuted].Render("  "+wrapText(phase.Detail, model.width)))
			}
		}
	}
	if model.cancelArmed && !model.cancellationState {
		lines = append(lines, styles[VisualRoleWarning].Render("Press Esc again to cancel"))
	}
	if model.cancellationState {
		lines = append(lines, styles[VisualRoleError].Render("Cancelling..."))
	}
	return strings.Join(lines, "\n")
}

func (model *trackedTeaModel) finalDocument() PresentationDocument {
	phase := model.currentPhase()
	if phase.Name == "" {
		phase.Name = model.label
	}
	if model.cancellationState && phase.State == PhaseActive {
		phase.State = PhaseCancelled
	}
	blocks := []PresentationBlock{{Role: phaseRole(phase.State), Text: phase.Name}}
	if phase.Detail != "" {
		blocks = append(blocks, PresentationBlock{Role: VisualRoleMuted, Text: phase.Detail})
	}
	return PresentationDocument{Blocks: blocks}
}

func (model *trackedTeaModel) currentPhase() OperationPhase {
	for index := len(model.phases) - 1; index >= 0; index-- {
		if model.phases[index].State == PhaseActive || model.phases[index].State == PhasePending {
			return model.phases[index]
		}
	}
	if len(model.phases) > 0 {
		return model.phases[len(model.phases)-1]
	}
	return OperationPhase{}
}

func trackedPhasePrefix(state PhaseState) string {
	switch state {
	case PhaseCompleted:
		return "[done]"
	case PhaseCancelled:
		return "[cancelled]"
	case PhaseFailed:
		return "[failed]"
	case PhasePending:
		return "[pending]"
	default:
		return "[active]"
	}
}
