package cm

import (
	"context"
	"errors"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

// ProfileResolver is the shared encrypted-configuration boundary used by Git CM.
type ProfileResolver interface {
	ResolveCMProfile(appconfig.CMResolveOptions) (appconfig.ResolvedCMProfile, error)
}

// CommitPrompt contains the generated message and safe diagnostics shown before committing.
type CommitPrompt struct {
	Message   string
	Generated GeneratedMessage
	Profile   ProfileDiagnostic
}

// CommitPrompter owns the default-yes commit confirmation boundary.
type CommitPrompter interface {
	ConfirmCommit(CommitPrompt) (confirmed bool, cancelled bool)
}

// Dependencies are Git CM's command-owned collaboration points.
type Dependencies struct {
	Git       GitInputRunner
	Files     SnapshotFileSystem
	Prompter  StagePrompter
	Committer CommitPrompter
	Resolver  ProfileResolver
	Transport ProviderTransport
}

// Module owns Git CM generation before confirmation, commit, and push handling.
type Module struct {
	git       GitInputRunner
	files     SnapshotFileSystem
	prompter  StagePrompter
	committer CommitPrompter
	resolver  ProfileResolver
	transport ProviderTransport
}

// New constructs an unregistered Git CM module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.Git == nil {
		return nil, errors.New("Git CM runner is required")
	}
	if dependencies.Files == nil {
		return nil, errors.New("Git CM filesystem is required")
	}
	if dependencies.Prompter == nil {
		return nil, errors.New("Git CM staging prompt is required")
	}
	if dependencies.Committer == nil {
		return nil, errors.New("Git CM commit prompt is required")
	}
	if dependencies.Resolver == nil {
		return nil, errors.New("Git CM profile resolver is required")
	}
	if dependencies.Transport == nil {
		return nil, errors.New("Git CM provider transport is required")
	}
	return &Module{
		git:       dependencies.Git,
		files:     dependencies.Files,
		prompter:  dependencies.Prompter,
		committer: dependencies.Committer,
		resolver:  dependencies.Resolver,
		transport: dependencies.Transport,
	}, nil
}

// Run applies optional staging and generates one validated commit message.
func (module *Module) Run(ctx context.Context, input Input) (Result, error) {
	mode, err := resolveExecutionMode(input)
	if err != nil {
		return Result{}, err
	}
	result := Result{Scope: mode.Scope}
	if mode.PromptStage {
		staged, err := StageSelectedChanges(ctx, module.git, module.files, module.prompter)
		if err != nil {
			return Result{}, err
		}
		result.RepositoryRoot = staged.RepositoryRoot
		if staged.Cancelled {
			result.Cancelled = true
			return result, nil
		}
		if staged.NothingSelected {
			result.NothingSelected = true
			return result, nil
		}
		if staged.NoChanges {
			result.NoChanges = true
			result.NoChangeScope = ScopeAllUncommitted
			return result, nil
		}
	}
	if mode.StageAll {
		root, err := StageAllChanges(ctx, module.git)
		if err != nil {
			return Result{}, err
		}
		result.RepositoryRoot = root
	}

	snapshot, err := CaptureSnapshot(ctx, module.git, module.files, mode.Scope)
	if err != nil {
		return Result{}, err
	}
	result.RepositoryRoot = snapshot.RepositoryRoot
	if len(snapshot.Files) == 0 {
		result.NoChanges = true
		result.NoChangeScope = mode.Scope
		return result, nil
	}
	profile, err := module.resolver.ResolveCMProfile(appconfig.CMResolveOptions{
		ProfileName:       input.Profile,
		TimeoutOverrideMS: input.TimeoutMS,
	})
	if err != nil {
		return Result{}, err
	}
	language, err := normalizeLanguage(input.Language)
	if err != nil {
		return Result{}, err
	}
	result.Profile = profileDiagnostic(profile)
	model, err := NewOpenAICompatibleModel(profile, module.transport)
	if err != nil {
		return Result{}, err
	}
	generated, err := GenerateCommitMessage(ctx, model, GenerationInput{
		Snapshot:    snapshot,
		Language:    language,
		IncludeBody: input.Body,
	})
	if err != nil {
		return result, err
	}
	result.Generated = &generated
	if !mode.CreateCommit {
		return result, nil
	}
	result.PromptedCommit = true
	confirmed, cancelled := module.committer.ConfirmCommit(CommitPrompt{
		Message:   "Create this commit?",
		Generated: generated,
		Profile:   result.Profile,
	})
	if cancelled || !confirmed {
		result.Cancelled = true
		return result, nil
	}
	if err := CommitSnapshot(ctx, module.git, module.files, CommitRequest{
		RepositoryRoot: result.RepositoryRoot,
		Scope:          result.Scope,
		SnapshotID:     generated.SnapshotID,
		Message:        generated.Message,
	}); err != nil {
		return result, err
	}
	result.Committed = true
	if !mode.Push {
		return result, nil
	}
	if err := PushCommit(ctx, module.git, result.RepositoryRoot, mode.PushRemote); err != nil {
		return result, err
	}
	result.Pushed = true
	result.PushRemote = mode.PushRemote
	return result, nil
}

func normalizeLanguage(value string) (string, error) {
	if value == "" {
		return "en", nil
	}
	if value != "en" && value != "zh" {
		return "", errors.New("Unsupported language. Use \"en\" or \"zh\".")
	}
	return value, nil
}
