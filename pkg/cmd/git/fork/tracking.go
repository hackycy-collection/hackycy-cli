package fork

import "context"

// PhaseKind identifies one externally meaningful Git Fork acquisition phase.
type PhaseKind uint8

const (
	PhaseResolve PhaseKind = iota
	PhaseDefaultBranch
	PhaseArchive
	PhaseClone
	PhaseReady
)

// PhaseState records the command-owned state of one Git Fork phase.
type PhaseState uint8

const (
	PhaseActive PhaseState = iota
	PhaseCompleted
	PhaseCancelled
	PhaseFailed
)

// Phase is one typed Git Fork acquisition update without terminal implementation details.
type Phase struct {
	Kind        PhaseKind
	State       PhaseState
	Repository  string
	Ref         string
	Destination string
}

// PhaseReporter consumes the updates for one contiguous Git Fork acquisition segment.
type PhaseReporter interface {
	Report(Phase)
	Close() error
}

// Tracker opens the command-owned Git Fork acquisition segment.
type Tracker interface {
	Start(context.Context) (PhaseReporter, error)
}

func (module *Module) track(ctx context.Context, result Result, work func(func(Phase)) (Result, error)) (tracked Result, err error) {
	reporter, err := module.tracker.Start(ctx)
	if err != nil {
		return Result{}, err
	}
	last := Phase{}
	reported := false
	report := func(phase Phase) {
		last = phase
		reported = true
		reporter.Report(phase)
	}
	defer func() {
		if reported && last.State == PhaseActive {
			if ctx.Err() != nil {
				last.State = PhaseCancelled
			} else if err != nil {
				last.State = PhaseFailed
			} else {
				last.State = PhaseCompleted
			}
			reporter.Report(last)
		}
		if closeErr := reporter.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	return work(report)
}
