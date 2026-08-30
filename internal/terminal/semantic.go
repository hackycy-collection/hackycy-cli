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
	Role VisualRole
	Text string
}

// PresentationDocument is presentation content assembled from semantic roles.
type PresentationDocument struct {
	Blocks []PresentationBlock
}

// PhaseState describes one command-owned state in a tracked operation.
type PhaseState uint8

const (
	PhasePending PhaseState = iota
	PhaseActive
	PhaseCompleted
	PhaseCancelled
	PhaseFailed
)

// OperationPhase is one externally meaningful progress update from a command.
type OperationPhase struct {
	Name   string
	Detail string
	State  PhaseState
}

// TrackedOperation supplies terminal presentation with command-owned updates.
// It never carries a business-work callback: command orchestration remains outside
// the terminal module.
type TrackedOperation struct {
	Label         string
	Updates       <-chan OperationPhase
	RequestCancel func()
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
	Result(PresentationDocument) error
	Close() error
}
