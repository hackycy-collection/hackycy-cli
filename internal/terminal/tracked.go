package terminal

import (
	"io"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
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

func (run *runtimeRun) trackRich(controller *richController, operation TrackedOperation) error {
	if operation.Updates == nil {
		return nil
	}
	var cancelOnce sync.Once
	requestCancel := func() {
		if operation.RequestCancel != nil {
			cancelOnce.Do(operation.RequestCancel)
		}
	}
	if err := controller.startTrack(operation.Label, requestCancel); err != nil {
		return err
	}

	contextDone := run.ctx.Done()
	for {
		select {
		case <-contextDone:
			requestCancel()
			if err := controller.cancelTrack(); err != nil {
				return err
			}
			contextDone = nil
		case phase, open := <-operation.Updates:
			if !open {
				return controller.finishTrack()
			}
			if err := controller.updateTrack(phase); err != nil {
				return err
			}
		}
	}
}

type trackedState struct {
	label             string
	phases            []OperationPhase
	cancelArmed       bool
	cancellationState bool
	requestStop       func()
}

func (state *trackedState) requestCancellation() {
	if state.cancellationState {
		return
	}
	state.cancellationState = true
	if state.requestStop != nil {
		state.requestStop()
	}
}

func (state *trackedState) applyPhase(phase OperationPhase) {
	for index := len(state.phases) - 1; index >= 0; index-- {
		if state.phases[index].Name == phase.Name {
			state.phases[index] = phase
			return
		}
	}
	state.phases = append(state.phases, phase)
}

func (state *trackedState) view(width int, styles map[VisualRole]lipgloss.Style) string {
	lines := make([]string, 0, len(state.phases)*2+3)
	if state.label != "" {
		lines = append(lines, styles[VisualRoleTitle].Render(wrapText(state.label, width)))
	}
	if width > 0 && width < 48 {
		phase := state.currentPhase()
		if phase.Name != "" {
			lines = append(lines, styles[phaseRole(phase.State)].Render(wrapText(phase.Name, width)))
			if phase.Detail != "" {
				lines = append(lines, styles[VisualRoleMuted].Render(wrapText(phase.Detail, width)))
			}
		}
	} else {
		for _, phase := range state.phases {
			role := phaseRole(phase.State)
			lines = append(lines, styles[role].Render(trackedPhasePrefix(phase.State)+" "+wrapText(phase.Name, width)))
			if phase.Detail != "" {
				lines = append(lines, styles[VisualRoleMuted].Render("  "+wrapText(phase.Detail, width)))
			}
		}
	}
	if state.cancelArmed && !state.cancellationState {
		lines = append(lines, styles[VisualRoleWarning].Render("Press Esc again to cancel"))
	}
	if state.cancellationState {
		lines = append(lines, styles[VisualRoleError].Render("Cancelling..."))
	}
	return strings.Join(lines, "\n")
}

func (state *trackedState) finalDocument() PresentationDocument {
	phase := state.currentPhase()
	if phase.Name == "" {
		phase.Name = state.label
	}
	if state.cancellationState && (phase.State == PhaseActive || phase.State == PhasePending) {
		phase.State = PhaseCancelled
	}
	blocks := []PresentationBlock{{Role: phaseRole(phase.State), Text: phase.Name}}
	if phase.Detail != "" {
		blocks = append(blocks, PresentationBlock{Role: VisualRoleMuted, Text: phase.Detail})
	}
	return PresentationDocument{Blocks: blocks}
}

func (state *trackedState) currentPhase() OperationPhase {
	for index := len(state.phases) - 1; index >= 0; index-- {
		if state.phases[index].State == PhaseActive || state.phases[index].State == PhasePending {
			return state.phases[index]
		}
	}
	if len(state.phases) > 0 {
		return state.phases[len(state.phases)-1]
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
