package cm

import "context"

// PhaseKind identifies one externally meaningful Git CM work phase.
type PhaseKind uint8

const (
	PhaseStage PhaseKind = iota
	PhaseCollect
	PhaseGenerate
	PhaseCommit
	PhasePush
)

// PhaseState records the command-owned state of one Git CM phase.
type PhaseState uint8

const (
	PhaseActive PhaseState = iota
	PhaseCompleted
	PhaseCancelled
	PhaseFailed
)

// Phase is one typed Git CM work update without terminal implementation details.
type Phase struct {
	Kind      PhaseKind
	State     PhaseState
	FileCount int
	Remote    string
}

// PhaseReporter consumes the updates for one contiguous Git CM work segment.
type PhaseReporter interface {
	Report(Phase)
	Close() error
}

// Tracker opens a command-owned Git CM work segment.
type Tracker interface {
	Start(context.Context) (PhaseReporter, error)
}

func (module *Module) track(ctx context.Context, work func(func(Phase)) (Result, error)) (result Result, err error) {
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
