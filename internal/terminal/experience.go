package terminal

import (
	"context"
	"errors"
	"io"
	"sync"
)

// ErrExperienceRunClosed reports use after a terminal run has returned control.
var ErrExperienceRunClosed = errors.New("terminal experience run is closed")

// ExperienceOptions supplies the terminal-owned dependencies for one invocation.
type ExperienceOptions struct {
	Session     Session
	Input       io.Reader
	Output      io.Writer
	Diagnostics io.Writer
	Width       int
}

// Runtime is the concrete terminal Experience for one invocation.
type Runtime struct {
	session     Session
	input       io.Reader
	output      io.Writer
	diagnostics *LeaseAwareDiagnosticWriter
	width       int
}

// NewExperience constructs a terminal Experience from explicit inherited streams.
func NewExperience(options ExperienceOptions) *Runtime {
	if options.Input == nil {
		options.Input = emptyInput{}
	}
	if options.Output == nil {
		options.Output = io.Discard
	}
	return &Runtime{
		session:     options.Session,
		input:       options.Input,
		output:      options.Output,
		diagnostics: NewLeaseAwareDiagnosticWriter(options.Diagnostics),
		width:       options.Width,
	}
}

// Session returns the immutable capability selected for this Experience.
func (runtime *Runtime) Session() Session {
	return runtime.session
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
			Session: runtime.session,
			Input:   runtime.input,
		}),
	}
}

// DiagnosticWriter returns the lease-aware diagnostic stream for normal records.
func (runtime *Runtime) DiagnosticWriter() io.Writer {
	return runtime.diagnostics
}

type runtimeRun struct {
	runtime      *Runtime
	ctx          context.Context
	interactions *InteractionHandler

	mu        sync.Mutex
	operation sync.Mutex
	closed    bool
}

func (run *runtimeRun) Ask(request InteractionRequest) (answer InteractionAnswer, err error) {
	run.operation.Lock()
	defer run.operation.Unlock()
	if err := run.available(); err != nil {
		return InteractionAnswer{}, err
	}
	lease := run.runtime.diagnostics.AcquireRendererLease()
	defer func() {
		if closeErr := lease.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	return run.interactions.withDiagnostics(lease.Writer()).Ask(run.ctx, request)
}

func (run *runtimeRun) Present(document PresentationDocument) error {
	run.operation.Lock()
	defer run.operation.Unlock()
	if err := run.available(); err != nil {
		return err
	}
	if run.runtime.session.Kind == RichInteractive {
		return WriteRich(run.runtime.output, document, RichOptions{
			Width: run.runtime.width,
			Color: run.runtime.session.Color,
		})
	}
	return WritePlain(run.runtime.output, document)
}

func (run *runtimeRun) Track(operation TrackedOperation) (err error) {
	run.operation.Lock()
	defer run.operation.Unlock()
	if err := run.available(); err != nil {
		return err
	}
	lease := run.runtime.diagnostics.AcquireRendererLease()
	defer func() {
		if closeErr := lease.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if operation.Updates == nil {
		return nil
	}
	for {
		select {
		case <-run.ctx.Done():
			if operation.RequestCancel != nil {
				operation.RequestCancel()
			}
			return run.ctx.Err()
		case phase, open := <-operation.Updates:
			if !open {
				return nil
			}
			if err := run.presentPhase(lease.Writer(), phase); err != nil {
				return err
			}
		}
	}
}

func (run *runtimeRun) Close() error {
	run.operation.Lock()
	defer run.operation.Unlock()
	run.mu.Lock()
	defer run.mu.Unlock()
	run.closed = true
	return nil
}

func (run *runtimeRun) available() error {
	run.mu.Lock()
	defer run.mu.Unlock()
	if run.closed {
		return ErrExperienceRunClosed
	}
	return nil
}

func (run *runtimeRun) presentPhase(output io.Writer, phase OperationPhase) error {
	document := PresentationDocument{Blocks: []PresentationBlock{{
		Role: phaseRole(phase.State),
		Text: phase.Name,
	}}}
	if phase.Detail != "" {
		document.Blocks = append(document.Blocks, PresentationBlock{Role: VisualRoleMuted, Text: phase.Detail})
	}
	if run.runtime.session.Kind == RichInteractive {
		return WriteRich(output, document, RichOptions{
			Width: run.runtime.width,
			Color: run.runtime.session.Color,
		})
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
