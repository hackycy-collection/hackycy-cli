package main

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	cmcommand "github.com/hackycy/hackycy-cli/internal/commands/git/cm"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalGitCMAdapterTranslatesFormsPhasesAndPresentation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Values: []string{"one.go"}}},
		terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Confirmed: true}},
	)
	run := experience.Open(context.Background())
	adapter := newTerminalGitCMAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}, func() {})
	stagePrompt := cmcommand.StagePrompt{
		Message:       "Select files to stage",
		Options:       []cmcommand.StageOption{{Value: "one.go", Label: "M one.go"}, {Value: "two.go", Label: "A two.go"}},
		InitialValues: []string{"one.go", "two.go"},
	}
	selected, cancelled, err := adapter.SelectFiles(stagePrompt)
	if err != nil || cancelled || !reflect.DeepEqual(selected, []string{"one.go"}) {
		t.Fatalf("SelectFiles() = (%#v, %t, %v)", selected, cancelled, err)
	}
	commitPrompt := cmcommand.CommitPrompt{
		Message:   "Create this commit?",
		Generated: cmcommand.GeneratedMessage{Message: "feat(cm): present output", Evidence: cmcommand.EvidenceCoverage{EstimatedLocalPromptTokens: 456, RepresentedClusters: 1, TotalClusters: 1, IncludedFacts: 4}},
		Profile:   cmcommand.ProfileDiagnostic{Name: "work", Model: "model"},
	}
	confirmed, cancelled, err := adapter.ConfirmCommit(commitPrompt)
	if err != nil || cancelled || !confirmed {
		t.Fatalf("ConfirmCommit() = (%t, %t, %v)", confirmed, cancelled, err)
	}
	reporter, err := adapter.Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	reporter.Report(cmcommand.Phase{Kind: cmcommand.PhaseCommit, State: cmcommand.PhaseActive})
	reporter.Report(cmcommand.Phase{Kind: cmcommand.PhaseCommit, State: cmcommand.PhaseCompleted})
	if err := reporter.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	result := cmcommand.Result{Generated: &commitPrompt.Generated, Profile: commitPrompt.Profile, Committed: true}
	if err := adapter.PresentGenerated(result); err != nil {
		t.Fatalf("PresentGenerated() error = %v", err)
	}
	if err := adapter.PresentOutcome(result); err != nil {
		t.Fatalf("PresentOutcome() error = %v", err)
	}
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 6 || operations[0].Kind != terminaltest.AskOperation || operations[1].Kind != terminaltest.AskOperation || operations[2].Kind != terminaltest.TrackOperation || operations[5].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	stageRequest := operations[0].Value.(terminalexperience.InteractionRequest)
	if stageRequest.Kind != terminalexperience.InteractionMultiSelect || !stageRequest.HasDefault || !reflect.DeepEqual(stageRequest.Default.Values, stagePrompt.InitialValues) || !reflect.DeepEqual(stageRequest.CancelValues, []string{"q", "quit", "cancel"}) {
		t.Fatalf("stage request = %#v", stageRequest)
	}
	commitRequest := operations[1].Value.(terminalexperience.InteractionRequest)
	if commitRequest.Kind != terminalexperience.InteractionConfirm || !commitRequest.HasDefault || !commitRequest.Default.Confirmed || !strings.Contains(commitRequest.Description, "feat(cm): present output") || !strings.Contains(commitRequest.Description, "Profile: work (model)") {
		t.Fatalf("commit request = %#v", commitRequest)
	}
	operation := operations[2].Value.(terminalexperience.TrackedOperation)
	if operation.Label != "Git CM" {
		t.Fatalf("tracked operation = %#v", operation)
	}
	generated := operations[3].Value.(terminalexperience.PresentationDocument)
	if generated.Blocks[0].Role != terminalexperience.VisualRoleSuccess || !strings.Contains(generated.Blocks[0].Text, "feat(cm): present output") {
		t.Fatalf("generated document = %#v", generated)
	}
	outcome := operations[4].Value.(terminalexperience.PresentationDocument)
	if outcome.Blocks[0].Role != terminalexperience.VisualRoleSuccess || outcome.Blocks[0].Text != "Commit created" {
		t.Fatalf("outcome document = %#v", outcome)
	}
}

func TestTerminalGitCMAdapterMapsCancellationAndAutomationInteraction(t *testing.T) {
	cancelledExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	cancelledAdapter := newTerminalGitCMAdapter(cancelledExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.RichInteractive}, func() {})
	if _, cancelled, err := cancelledAdapter.SelectFiles(cmcommand.StagePrompt{}); err != nil || !cancelled {
		t.Fatalf("SelectFiles() = (%t, %v)", cancelled, err)
	}

	automationExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	automationAdapter := newTerminalGitCMAdapter(automationExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation}, func() {})
	if _, _, err := automationAdapter.ConfirmCommit(cmcommand.CommitPrompt{}); !errors.Is(err, errGitCMRequiresInteractive) {
		t.Fatalf("ConfirmCommit() error = %v", err)
	}
}

func TestGitCMPlainJourneyKeepsFormsAndPhasesOnStderr(t *testing.T) {
	repository := newGitCMRepository(t)
	withGitCMWorkingDirectory(t, repository)
	writeGitCMFile(t, filepath.Join(repository, "README.md"), "plain journey\n")
	server, provider := newGitCMMessageProvider(t, "feat(cm): plain tracked journey")
	defer server.Close()
	configureGitCMProvider(t, server.URL)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("1\n\n"),
		Output:      stdout,
		Diagnostics: stderr,
	})

	result, err := newGitCMHandler(experience)(context.Background(), cmcommand.Input{Stage: true})
	if err != nil || !result.Committed || result.Pushed || provider.calls != 1 {
		t.Fatalf("Run() = (%#v, %v), provider calls = %d", result, err, provider.calls)
	}
	for _, expected := range []string{"feat(cm): plain tracked journey", "Profile: env (fixture-model)", "Commit created"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("stdout omitted %q: %q", expected, stdout.String())
		}
	}
	for _, expected := range []string{"Select files to stage", "1) A README.md", "Staging selected files", "Collecting changes", "Generating commit message", "Create this commit? [Y/n]:", "Creating commit"} {
		if !strings.Contains(stderr.String(), expected) {
			t.Fatalf("stderr omitted %q: %q", expected, stderr.String())
		}
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("Plain streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
	}
	if subject := strings.TrimSpace(gitCMOutput(t, repository, "log", "-1", "--format=%s")); subject != "feat(cm): plain tracked journey" {
		t.Fatalf("HEAD subject = %q", subject)
	}
}

func TestGitCMAutomationFailsBeforePromptDependentMutationOrOutput(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		arguments  []string
		prepare    func(*testing.T, string)
		wantStatus string
	}{
		{
			name:      "file selection",
			arguments: []string{"git", "cm", "--stage"},
			prepare: func(t *testing.T, repository string) {
				writeGitCMFile(t, filepath.Join(repository, "README.md"), "selection\n")
			},
			wantStatus: "?? README.md\n",
		},
		{
			name:      "commit confirmation",
			arguments: []string{"git", "cm", "--staged"},
			prepare: func(t *testing.T, repository string) {
				writeGitCMFile(t, filepath.Join(repository, "README.md"), "confirmation\n")
				runGitCM(t, repository, "add", "README.md")
			},
			wantStatus: "A  README.md\n",
		},
		{
			name:      "stage all confirmation",
			arguments: []string{"git", "cm", "--stage-all"},
			prepare: func(t *testing.T, repository string) {
				writeGitCMFile(t, filepath.Join(repository, "README.md"), "stage all\n")
			},
			wantStatus: "?? README.md\n",
		},
		{
			name:      "push confirmation",
			arguments: []string{"git", "cm", "--staged", "--push"},
			prepare: func(t *testing.T, repository string) {
				writeGitCMFile(t, filepath.Join(repository, "README.md"), "push\n")
				runGitCM(t, repository, "add", "README.md")
			},
			wantStatus: "A  README.md\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := newGitCMRepository(t)
			withGitCMWorkingDirectory(t, repository)
			testCase.prepare(t, repository)
			beforeHead := gitCMOutput(t, repository, "rev-parse", "HEAD")
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
				Input:       panicGitCMReader{},
				Output:      stdout,
				Diagnostics: stderr,
			})
			app, err := newRootCommandForTest("0.0.0-dev", rootTestDependencies{
				Out:     stdout,
				Err:     stderr,
				Logging: logging.NewRuntime(logging.Options{Writer: stderr}),
				GitCM:   newGitCMHandler(experience),
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			outcome := app.Execute(context.Background(), testCase.arguments)
			if outcome.Code != 1 || !errors.Is(outcome.Err, errGitCMRequiresInteractive) || stdout.Len() != 0 || stderr.String() != "error: git cm requires an interactive terminal\n" {
				t.Fatalf("Automation outcome = %#v, streams = (%q, %q)", outcome, stdout.String(), stderr.String())
			}
			if afterHead := gitCMOutput(t, repository, "rev-parse", "HEAD"); afterHead != beforeHead {
				t.Fatalf("Automation failure changed HEAD from %q to %q", beforeHead, afterHead)
			}
			if status := gitCMOutput(t, repository, "status", "--short"); status != testCase.wantStatus {
				t.Fatalf("Automation failure status = %q, want %q", status, testCase.wantStatus)
			}
			if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
				t.Fatalf("Automation streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
			}
		})
	}
}

func TestGitCMGenerationOnlyAutomationRetainsTheDurableResult(t *testing.T) {
	repository := newGitCMRepository(t)
	withGitCMWorkingDirectory(t, repository)
	writeGitCMFile(t, filepath.Join(repository, "README.md"), "automation generation\n")
	server, provider := newGitCMMessageProvider(t, "feat(cm): automation generation")
	defer server.Close()
	configureGitCMProvider(t, server.URL)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
		Input:       panicGitCMReader{},
		Output:      stdout,
		Diagnostics: stderr,
	})

	result, err := newGitCMHandler(experience)(context.Background(), cmcommand.Input{DryRun: true})
	if err != nil || result.Generated == nil || result.PromptedCommit || provider.calls != 1 {
		t.Fatalf("Run() = (%#v, %v), provider calls = %d", result, err, provider.calls)
	}
	if !strings.Contains(stdout.String(), "feat(cm): automation generation") || stderr.Len() != 0 {
		t.Fatalf("Automation streams = (%q, %q)", stdout.String(), stderr.String())
	}
	if status := gitCMOutput(t, repository, "status", "--short"); status != "?? README.md\n" {
		t.Fatalf("generation-only status = %q", status)
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("Automation streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
	}
}

func TestGitCMHandlerPresentsCommittedPartialOutcomeAfterPushFailure(t *testing.T) {
	repository := newGitCMRepository(t)
	withGitCMWorkingDirectory(t, repository)
	runGitCM(t, repository, "remote", "add", "origin", filepath.Join(t.TempDir(), "missing.git"))
	writeGitCMFile(t, filepath.Join(repository, "README.md"), "partial push\n")
	runGitCM(t, repository, "add", "README.md")
	beforeHead := gitCMOutput(t, repository, "rev-parse", "HEAD")
	server, provider := newGitCMMessageProvider(t, "feat(cm): retain partial commit")
	defer server.Close()
	configureGitCMProvider(t, server.URL)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("\n"),
		Output:      stdout,
		Diagnostics: stderr,
	})

	remote := "origin"
	result, err := newGitCMHandler(experience)(context.Background(), cmcommand.Input{Staged: true, Push: &remote})
	if err == nil || !result.Committed || result.Pushed || result.PushRemote != "" || provider.calls != 1 {
		t.Fatalf("Run() = (%#v, %v), provider calls = %d", result, err, provider.calls)
	}
	if !strings.Contains(stdout.String(), "feat(cm): retain partial commit") || !strings.Contains(stdout.String(), "Commit created") || strings.Contains(stdout.String(), "Commit created and pushed") {
		t.Fatalf("partial stdout = %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Pushing commit") {
		t.Fatalf("partial stderr = %q", stderr.String())
	}
	if afterHead := gitCMOutput(t, repository, "rev-parse", "HEAD"); afterHead == beforeHead {
		t.Fatalf("partial result did not retain a commit: HEAD = %q", afterHead)
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("Plain streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
	}
}

func TestGitCMDocumentsPreserveTheExistingPlainResults(t *testing.T) {
	for _, testCase := range []struct {
		result cmcommand.Result
		want   string
	}{
		{result: cmcommand.Result{NoChanges: true, NoChangeScope: cmcommand.ScopeAllUncommitted}, want: "No uncommitted changes.\n"},
		{result: cmcommand.Result{NoChanges: true, NoChangeScope: cmcommand.ScopeStaged}, want: "No staged changes.\n"},
		{result: cmcommand.Result{NothingSelected: true}, want: "Nothing selected.\n"},
		{result: cmcommand.Result{Cancelled: true}, want: "Cancelled\n"},
		{result: cmcommand.Result{Committed: true}, want: "Commit created\n"},
		{result: cmcommand.Result{Pushed: true}, want: "Commit created and pushed\n"},
	} {
		if got := terminalexperience.RenderPlain(gitCMOutcomeDocument(terminalexperience.Session{Kind: terminalexperience.PlainInteractive}, testCase.result)); got != testCase.want {
			t.Fatalf("Outcome(%#v) = %q, want %q", testCase.result, got, testCase.want)
		}
	}
	generated := gitCMGeneratedDocument(terminalexperience.Session{Kind: terminalexperience.PlainInteractive}, cmcommand.GeneratedMessage{Message: "feat(cm): compact", Evidence: cmcommand.EvidenceCoverage{EstimatedLocalPromptTokens: 4000, RepresentedClusters: 2, TotalClusters: 3, IncludedFacts: 18, OmittedFacts: 13, ContentCompacted: true}}, cmcommand.ProfileDiagnostic{Name: "work", Model: "model"})
	if got := terminalexperience.RenderPlain(generated); !strings.Contains(got, "Provider tokens: unavailable") || !strings.Contains(got, "4,000") || !strings.Contains(got, "3 clusters represented with compacted semantic evidence") {
		t.Fatalf("generated output = %q", got)
	}
}

func configureGitCMProvider(t *testing.T, baseURL string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", "")
	t.Setenv("YCY_CM_PROFILE", "")
	t.Setenv("YCY_CM_BASE_URL", baseURL)
	t.Setenv("YCY_CM_MODEL", "fixture-model")
	t.Setenv("YCY_CM_API_KEY", "fixture-api-key")
}

type panicGitCMReader struct{}

func (panicGitCMReader) Read([]byte) (int, error) {
	panic("git cm Automation must not read stdin")
}
