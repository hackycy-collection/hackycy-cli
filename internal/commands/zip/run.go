package zip

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Input is the typed command request after a later CLI binder applies defaults.
type Input struct {
	Directory string
	Open      bool
	WithDir   string
}

// ResultKind distinguishes legacy-visible normal outcomes from unexpected Go failures.
type ResultKind string

const (
	ResultCompleted         ResultKind = "completed"
	ResultCancelled         ResultKind = "cancelled"
	ResultDirectoryNotFound ResultKind = "directory-not-found"
	ResultPathNotDirectory  ResultKind = "path-not-directory"
	ResultCollectionFailed  ResultKind = "collection-failed"
	ResultNoFiles           ResultKind = "no-files"
	ResultNoValidFiles      ResultKind = "no-valid-files"
	ResultCompressionFailed ResultKind = "compression-failed"
	ResultWriteFailed       ResultKind = "write-failed"
)

// Result records the observable command result while preserving the legacy success-status branches.
type Result struct {
	Kind           ResultKind
	Plan           *ZipPlan
	OutputPath     string
	CollectedCount int
	IncludedCount  int
	Cause          error
	RevealFailed   bool
}

// Prompter owns the four command-owned interactions without knowing terminal implementation details.
type Prompter interface {
	SelectPackage(SelectPackageStep) (string, bool, error)
	SelectSource(SelectSourceStep) (string, bool, error)
	SelectGlob(SelectGlobStep) ([]string, bool, error)
	EditOutputFile(EditOutputFileStep) (string, bool, error)
}

// PathStater validates the selected source only after the interactive plan is complete.
type PathStater interface {
	Stat(string) (fs.FileInfo, error)
}

// ArchiveCollector owns glob expansion and regular-file selection.
type ArchiveCollector interface {
	CollectArchiveFiles(string, []string) ([]ArchiveEntry, error)
}

// ArchiveBuilder owns complete in-memory ZIP creation.
type ArchiveBuilder interface {
	BuildZipData([]ArchiveEntry, string, string) ([]byte, int, error)
}

// ArchiveWriter owns the final direct output publication.
type ArchiveWriter interface {
	WriteZipFile(string, []byte) error
}

// Revealer opens a completed archive in the host shell when enabled.
type Revealer interface {
	Reveal(string) error
}

// Dependencies are the command-owned boundaries required by the complete flow.
type Dependencies struct {
	Prompter           Prompter
	Presenter          Presenter
	RemoteNameResolver RemoteNameResolver
	Stater             PathStater
	Collector          ArchiveCollector
	Builder            ArchiveBuilder
	Writer             ArchiveWriter
	Revealer           Revealer
}

// Module composes planning, archiving, and result mapping while remaining terminal-independent.
type Module struct {
	prompter           Prompter
	presenter          Presenter
	remoteNameResolver RemoteNameResolver
	stater             PathStater
	collector          ArchiveCollector
	builder            ArchiveBuilder
	writer             ArchiveWriter
	revealer           Revealer
}

// New creates an unregistered ZIP command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.Prompter == nil {
		return nil, errors.New("zip prompter is required")
	}
	if dependencies.Stater == nil {
		dependencies.Stater = osPathStater{}
	}
	if dependencies.Collector == nil {
		dependencies.Collector = archiveCollectorFunc(CollectArchiveFiles)
	}
	if dependencies.Builder == nil {
		dependencies.Builder = archiveBuilderFunc(BuildZipData)
	}
	if dependencies.Writer == nil {
		dependencies.Writer = archiveWriterFunc(WriteZipFile)
	}
	if dependencies.Presenter == nil {
		dependencies.Presenter = discardPresenter{}
	}
	return &Module{
		prompter:           dependencies.Prompter,
		presenter:          dependencies.Presenter,
		remoteNameResolver: dependencies.RemoteNameResolver,
		stater:             dependencies.Stater,
		collector:          dependencies.Collector,
		builder:            dependencies.Builder,
		writer:             dependencies.Writer,
		revealer:           dependencies.Revealer,
	}, nil
}

// Run executes the full legacy planning and archive flow without assigning a process exit code.
func (module *Module) Run(input Input) (Result, error) {
	presentIntroduction(module.presenter)
	plan, cancelled, err := module.resolvePlan(input.Directory)
	if err != nil {
		return Result{}, err
	}
	if cancelled {
		presentCancellation(module.presenter)
		return Result{Kind: ResultCancelled}, nil
	}

	info, err := module.stater.Stat(plan.Input)
	if err != nil {
		presentDirectoryNotFound(module.presenter, plan.Input)
		return Result{Kind: ResultDirectoryNotFound, Plan: &plan, Cause: err}, nil
	}
	if !info.IsDir() {
		presentPathNotDirectory(module.presenter, plan.Input)
		return Result{Kind: ResultPathNotDirectory, Plan: &plan}, nil
	}

	outputPath := filepath.Join(plan.Input, plan.File+".zip")
	presentCollectionStart(module.presenter)
	entries, err := module.collector.CollectArchiveFiles(plan.Input, plan.Glob)
	if err != nil {
		presentCollectionFailure(module.presenter, err)
		return Result{Kind: ResultCollectionFailed, Plan: &plan, OutputPath: outputPath, Cause: err}, nil
	}
	if len(entries) == 0 {
		presentNoFiles(module.presenter)
		return Result{Kind: ResultNoFiles, Plan: &plan, OutputPath: outputPath}, nil
	}
	presentCollectedFiles(module.presenter, len(entries))

	presentCompressionStart(module.presenter)
	data, included, err := module.builder.BuildZipData(entries, outputPath, input.WithDir)
	if err != nil {
		if errors.Is(err, errNoValidArchiveFiles) {
			presentNoValidFiles(module.presenter)
			return Result{Kind: ResultNoValidFiles, Plan: &plan, OutputPath: outputPath, CollectedCount: len(entries), Cause: err}, nil
		}
		presentCompressionFailure(module.presenter, err)
		return Result{Kind: ResultCompressionFailed, Plan: &plan, OutputPath: outputPath, CollectedCount: len(entries), Cause: err}, nil
	}
	presentCompressionComplete(module.presenter, included)
	presentWritingStart(module.presenter)
	if err := module.writer.WriteZipFile(outputPath, data); err != nil {
		presentWriteFailure(module.presenter, err)
		return Result{Kind: ResultWriteFailed, Plan: &plan, OutputPath: outputPath, CollectedCount: len(entries), IncludedCount: included, Cause: err}, nil
	}

	result := Result{
		Kind:           ResultCompleted,
		Plan:           &plan,
		OutputPath:     outputPath,
		CollectedCount: len(entries),
		IncludedCount:  included,
	}
	if input.Open && module.revealer != nil && module.revealer.Reveal(outputPath) != nil {
		result.RevealFailed = true
	}
	presentSavedArchive(module.presenter, outputPath)
	return result, nil
}

func (module *Module) resolvePlan(directory string) (ZipPlan, bool, error) {
	session, err := CreatePlanningSession(directory)
	if err != nil {
		return ZipPlan{}, false, err
	}
	for {
		resolution, err := ResolvePlanningStep(session, module.remoteNameResolver)
		if err != nil {
			return ZipPlan{}, false, err
		}
		session = resolution.Session
		var answer PlanningAnswer
		switch step := resolution.Step.(type) {
		case SelectPackageStep:
			presentPlanningNote(module.presenter, step.Note)
			value, cancelled, err := module.prompter.SelectPackage(step)
			if err != nil {
				return ZipPlan{}, false, err
			}
			if cancelled {
				return ZipPlan{}, true, nil
			}
			answer = PlanningAnswer{Type: PlanningAnswerPackage, Value: value}
		case SelectSourceStep:
			presentPlanningNote(module.presenter, step.Note)
			value, cancelled, err := module.prompter.SelectSource(step)
			if err != nil {
				return ZipPlan{}, false, err
			}
			if cancelled {
				return ZipPlan{}, true, nil
			}
			answer = PlanningAnswer{Type: PlanningAnswerSource, Value: value}
		case SelectGlobStep:
			values, cancelled, err := module.prompter.SelectGlob(step)
			if err != nil {
				return ZipPlan{}, false, err
			}
			if cancelled {
				return ZipPlan{}, true, nil
			}
			answer = PlanningAnswer{Type: PlanningAnswerGlob, Values: values}
		case EditOutputFileStep:
			value, cancelled, err := module.prompter.EditOutputFile(step)
			if err != nil {
				return ZipPlan{}, false, err
			}
			if cancelled {
				return ZipPlan{}, true, nil
			}
			answer = PlanningAnswer{Type: PlanningAnswerOutput, Value: value}
		case CompleteStep:
			presentPlanningNote(module.presenter, step.Note)
			return step.Plan, false, nil
		default:
			return ZipPlan{}, false, fmt.Errorf("unknown zip planning step %T", resolution.Step)
		}
		session, err = ApplyPlanningAnswer(session, answer)
		if err != nil {
			return ZipPlan{}, false, err
		}
	}
}

type osPathStater struct{}

func (osPathStater) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

type archiveCollectorFunc func(string, []string) ([]ArchiveEntry, error)

func (function archiveCollectorFunc) CollectArchiveFiles(directory string, patterns []string) ([]ArchiveEntry, error) {
	return function(directory, patterns)
}

type archiveBuilderFunc func([]ArchiveEntry, string, string) ([]byte, int, error)

func (function archiveBuilderFunc) BuildZipData(entries []ArchiveEntry, outputPath, withDir string) ([]byte, int, error) {
	return function(entries, outputPath, withDir)
}

type archiveWriterFunc func(string, []byte) error

func (function archiveWriterFunc) WriteZipFile(outputPath string, data []byte) error {
	return function(outputPath, data)
}
