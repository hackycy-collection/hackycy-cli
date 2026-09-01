package terminal

import (
	"context"
	"io"
)

// InteractionKind identifies the answer shape a terminal interaction needs.
type InteractionKind uint8

const (
	InteractionText InteractionKind = iota
	InteractionSecret
	InteractionSelect
	InteractionMultiSelect
	InteractionConfirm
)

// InteractionOption is one semantic choice supplied by a command adapter.
type InteractionOption struct {
	Label       string
	Value       string
	Description string
}

// InteractionAnswer contains the value returned by one semantic interaction.
// The request kind determines which field is meaningful.
type InteractionAnswer struct {
	Value     string
	Values    []string
	Confirmed bool
}

// FinishOutcome is the durable semantic outcome for one finite command run.
type FinishOutcome uint8

const (
	// Succeeded reports that the command completed its intended work.
	Succeeded FinishOutcome = iota + 1
	// Cancelled reports that the caller or user cancelled the command.
	Cancelled
	// Failed reports that the command finished with a known failure.
	Failed
)

func (outcome FinishOutcome) String() string {
	switch outcome {
	case Succeeded:
		return "succeeded"
	case Cancelled:
		return "cancelled"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}

// InteractionRequest describes command intent without choosing a prompt toolkit.
type InteractionRequest struct {
	Kind         InteractionKind
	Message      string
	Description  string
	Placeholder  string
	Options      []InteractionOption
	Default      InteractionAnswer
	HasDefault   bool
	CancelValues []string
	PlainLead    string
	PlainPrompt  string
	// TranscriptLabel is the safe label used for the completed answer marker.
	TranscriptLabel string
	// Sensitive prevents the request value from entering a Live View or transcript.
	Sensitive bool
	// ParsePlain preserves a command-owned established Plain Interactive input grammar.
	// It is not used by Rich Interactive forms or Automation mode.
	ParsePlain func(string) (InteractionAnswer, error)
	Validate   func(InteractionAnswer) error
}

// VisualRole identifies the semantic meaning of presentation text.
type VisualRole uint8

const (
	VisualRolePlain VisualRole = iota
	VisualRoleTitle
	VisualRoleActive
	VisualRoleSuccess
	VisualRoleWarning
	VisualRoleError
	VisualRoleMuted
)

// PresentationBlock is one block of terminal presentation text.
type PresentationBlock struct {
	Role      VisualRole
	Text      string
	Sensitive bool
}

// PresentationDocument is presentation content assembled from semantic roles.
type PresentationDocument struct {
	Blocks []PresentationBlock
}

// PhaseDefinition is one immutable command-defined entry in a tracked phase catalog.
type PhaseDefinition struct {
	ID   string
	Name string
}

// Phase is a compatibility alias for callers that use the shorter catalog name.
type Phase = PhaseDefinition

// PhaseState describes one command-owned state in a tracked operation.
type PhaseState uint8

const (
	PhasePending PhaseState = iota
	PhaseActive
	PhaseCompleted
	PhaseCancelled
	PhaseFailed
)

func (state PhaseState) String() string {
	switch state {
	case PhasePending:
		return "pending"
	case PhaseActive:
		return "active"
	case PhaseCompleted:
		return "completed"
	case PhaseCancelled:
		return "cancelled"
	case PhaseFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// OperationPhase is one externally meaningful progress update from a command.
type OperationPhase struct {
	ID      string
	PhaseID string
	Name    string
	Detail  string
	State   PhaseState
}

// TrackedOperation supplies terminal presentation with command-owned updates.
// It never carries a business-work callback: command orchestration remains outside
// the terminal module.
type TrackedOperation struct {
	ID               string
	OperationID      string
	Label            string
	Phases           []PhaseDefinition
	PhaseDefinitions []PhaseDefinition
	Updates          <-chan OperationPhase
	RequestCancel    func()
}

// Experience opens independently closable terminal runs and owns diagnostics.
type Experience interface {
	Open(context.Context) ExperienceRun
	DiagnosticWriter() io.Writer
}

// ExperienceRun exposes only semantic terminal operations for one command context.
type ExperienceRun interface {
	Ask(InteractionRequest) (InteractionAnswer, error)
	Track(TrackedOperation) error
	Notice(PresentationDocument) error
	Milestone(PresentationDocument) error
	Finish(FinishOutcome, *PresentationDocument) error
	// Result remains for command adapters that have not yet migrated to Finish.
	Result(PresentationDocument) error
	Close() error
}
