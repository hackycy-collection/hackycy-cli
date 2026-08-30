package terminal

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

var (
	// ErrExperienceRunClosed reports use after a terminal run has returned control.
	ErrExperienceRunClosed = errors.New("terminal experience run is closed")
	// ErrExperienceRunFinished reports interactive use after the first durable result.
	ErrExperienceRunFinished = errors.New("terminal experience run has emitted its result")
)

// ExperienceOptions supplies terminal-owned dependencies for one invocation.
type ExperienceOptions struct {
	Capabilities Capabilities
	Input        io.Reader
	Output       io.Writer
	Diagnostics  io.Writer
	Width        int
	Height       int
}

// Runtime is the concrete terminal Experience for one invocation.
type Runtime struct {
	capabilities       Capabilities
	input              io.Reader
	output             io.Writer
	diagnostics        *LeaseAwareDiagnosticWriter
	inputTerminal      *os.File
	outputTerminal     *os.File
	diagnosticTerminal *os.File
	width              int
	height             int
}

// NewExperience constructs a terminal Experience from explicit inherited streams.
func NewExperience(options ExperienceOptions) *Runtime {
	if options.Input == nil {
		options.Input = emptyInput{}
	}
	if options.Output == nil {
		options.Output = io.Discard
	}
	runtime := &Runtime{
		capabilities: options.Capabilities,
		input:        options.Input,
		output:       options.Output,
		diagnostics:  NewLeaseAwareDiagnosticWriter(options.Diagnostics),
		width:        options.Width,
		height:       options.Height,
	}
	runtime.inputTerminal, _ = options.Input.(*os.File)
	runtime.outputTerminal, _ = options.Output.(*os.File)
	runtime.diagnosticTerminal, _ = options.Diagnostics.(*os.File)
	return runtime
}

// Capabilities returns the immutable capabilities selected for this Experience.
func (runtime *Runtime) Capabilities() Capabilities {
	return runtime.capabilities
}

// Open starts a per-command terminal run.
func (runtime *Runtime) Open(ctx context.Context) ExperienceRun {
	if ctx == nil {
		ctx = context.Background()
	}
	return &runtimeRun{
		runtime: runtime,
		ctx:     ctx,
		interactions: NewInteractionHandler(InteractionOptions{
			Capabilities: runtime.capabilities,
			Input:        runtime.input,
			Diagnostics:  runtime.diagnostics,
		}),
	}
}

// DiagnosticWriter coordinates normal diagnostic records with the active Rich UI.
func (runtime *Runtime) DiagnosticWriter() io.Writer {
	return runtime.diagnostics
}

type runtimeRun struct {
	runtime      *Runtime
	ctx          context.Context
	interactions *InteractionHandler

	operation    sync.Mutex
	state        runState
	richDisabled bool
	controller   *richController
}

type runState uint8

const (
	runActive runState = iota
	runFinished
	runClosed
)

func (run *runtimeRun) Ask(request InteractionRequest) (InteractionAnswer, error) {
	run.operation.Lock()
	defer run.operation.Unlock()
	if err := run.interactiveAvailable(); err != nil {
		return InteractionAnswer{}, err
	}
	if run.richEnabled() {
		controller, err := run.ensureRich()
		if err == nil {
			return controller.ask(run.ctx, run.interactions, request)
		}
		if !errors.Is(err, errRichUnavailable) {
			return InteractionAnswer{}, err
		}
		run.disableRich()
	}
	return run.interactions.Ask(run.ctx, request)
}

func (run *runtimeRun) Track(operation TrackedOperation) error {
	run.operation.Lock()
	defer run.operation.Unlock()
	if err := run.interactiveAvailable(); err != nil {
		return err
	}
	if run.richEnabled() {
		controller, err := run.ensureRich()
		if err == nil {
			return run.trackRich(controller, operation)
		}
		if !errors.Is(err, errRichUnavailable) {
			return err
		}
		run.disableRich()
	}
	if run.runtime.capabilities.Interaction == Automation {
		return run.trackSilently(operation)
	}
	return run.trackPlain(run.runtime.diagnostics, operation)
}

func (run *runtimeRun) Notice(document PresentationDocument) error {
	run.operation.Lock()
	defer run.operation.Unlock()
	if err := run.interactiveAvailable(); err != nil {
		return err
	}
	if run.runtime.capabilities.Interaction == Automation {
		return nil
	}
	if run.richEnabled() {
		controller, err := run.ensureRich()
		if err == nil {
			return controller.notice(document)
		}
		if !errors.Is(err, errRichUnavailable) {
			return err
		}
		run.disableRich()
	}
	return WritePlain(run.runtime.diagnostics, document)
}

func (run *runtimeRun) Result(document PresentationDocument) error {
	run.operation.Lock()
	defer run.operation.Unlock()
	if run.state == runClosed {
		return ErrExperienceRunClosed
	}

	var restoreErr error
	if run.state == runActive {
		run.state = runFinished
		restoreErr = run.stopRich()
	}
	return errors.Join(restoreErr, run.writeResult(document))
}

func (run *runtimeRun) Close() error {
	run.operation.Lock()
	defer run.operation.Unlock()
	if run.state == runClosed {
		return nil
	}
	run.state = runClosed
	return run.stopRich()
}

func (run *runtimeRun) interactiveAvailable() error {
	switch run.state {
	case runFinished:
		return ErrExperienceRunFinished
	case runClosed:
		return ErrExperienceRunClosed
	default:
		return nil
	}
}

func (run *runtimeRun) richEnabled() bool {
	return run.runtime.capabilities.Interaction == RichInteractive && !run.richDisabled
}

func (run *runtimeRun) disableRich() {
	run.richDisabled = true
	capabilities := run.interactions.capabilities
	capabilities.Interaction = PlainInteractive
	run.interactions.capabilities = capabilities
}

func (run *runtimeRun) ensureRich() (*richController, error) {
	if run.controller != nil {
		return run.controller, nil
	}
	controller := newRichController(run.runtime)
	if err := controller.start(); err != nil {
		return nil, err
	}
	run.controller = controller
	return controller, nil
}

func (run *runtimeRun) stopRich() error {
	if run.controller == nil {
		return nil
	}
	controller := run.controller
	run.controller = nil
	return controller.close()
}

func (run *runtimeRun) writeResult(document PresentationDocument) error {
	if !run.runtime.capabilities.Stdout.Terminal {
		return WritePlain(run.runtime.output, document)
	}
	width := run.runtime.width
	if run.runtime.outputTerminal != nil {
		if terminalWidth, _, err := term.GetSize(int(run.runtime.outputTerminal.Fd())); err == nil {
			width = terminalWidth
		}
	}
	return WriteRich(run.runtime.output, document, RichOptions{
		Width: width,
		Color: run.runtime.capabilities.Stdout.Color,
	})
}

func (run *runtimeRun) presentPhase(output io.Writer, phase OperationPhase) error {
	document := PresentationDocument{Blocks: []PresentationBlock{{
		Role: phaseRole(phase.State),
		Text: phase.Name,
	}}}
	if phase.Detail != "" {
		document.Blocks = append(document.Blocks, PresentationBlock{Role: VisualRoleMuted, Text: phase.Detail})
	}
	return WritePlain(output, document)
}

func phaseRole(state PhaseState) VisualRole {
	switch state {
	case PhaseCompleted:
		return VisualRoleSuccess
	case PhaseCancelled, PhaseFailed:
		return VisualRoleError
	case PhasePending, PhaseActive:
		return VisualRoleActive
	default:
		return VisualRolePlain
	}
}

type emptyInput struct{}

func (emptyInput) Read([]byte) (int, error) {
	return 0, io.EOF
}

var _ Experience = (*Runtime)(nil)
var _ ExperienceRun = (*runtimeRun)(nil)
