package fork

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/gitprocess"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalGitForkAdapterTranslatesConfirmationPhasesAndPresentation(t *testing.T) {
	experience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Value: terminalexperience.InteractionAnswer{Confirmed: true}})
	run := experience.Open(context.Background())
	adapter := newTerminalGitForkAdapter(run, terminalexperience.Session{Kind: terminalexperience.RichInteractive, Color: true}, func() {})
	prompt := OverwritePrompt{Destination: "project", Message: "Directory \"project\" is not empty. Overwrite?"}

	adapter.Introduction()
	confirmed, cancelled, err := adapter.ConfirmOverwrite(prompt)
	if err != nil || cancelled || !confirmed {
		t.Fatalf("ConfirmOverwrite() = (%t, %t, %v)", confirmed, cancelled, err)
	}
	reporter, err := adapter.Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	reporter.Report(Phase{Kind: PhaseArchive, State: PhaseActive, Ref: "main", Destination: "chosen"})
	if err := reporter.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	adapter.Outcome(Result{
		Repository:  Repository{Host: "github.example", Owner: "group", Name: "project", ProviderType: "github"},
		Destination: "chosen",
		Ref:         "main",
		Acquisition: "archive",
	})
	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 5 || operations[0].Kind != terminaltest.PresentOperation || operations[1].Kind != terminaltest.AskOperation || operations[2].Kind != terminaltest.TrackOperation || operations[3].Kind != terminaltest.PresentOperation || operations[4].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	confirmation := operations[1].Value.(terminalexperience.InteractionRequest)
	if confirmation.Kind != terminalexperience.InteractionConfirm || confirmation.Message != prompt.Message || !confirmation.HasDefault || !confirmation.Default.Confirmed || confirmation.PlainPrompt != prompt.Message+" [Y/n]: " || !reflect.DeepEqual(confirmation.CancelValues, []string{"q", "quit", "cancel"}) {
		t.Fatalf("confirmation request = %#v", confirmation)
	}
	if _, err := confirmation.ParsePlain("maybe"); err == nil || err.Error() != "Invalid confirmation" {
		t.Fatalf("invalid confirmation parse error = %v", err)
	}
	tracked := operations[2].Value.(terminalexperience.TrackedOperation)
	if tracked.Label != "Git Fork" || tracked.RequestCancel == nil {
		t.Fatalf("tracked operation = %#v", tracked)
	}
	var phases []terminalexperience.OperationPhase
	for phase := range tracked.Updates {
		phases = append(phases, phase)
	}
	if got, want := phases, []terminalexperience.OperationPhase{{Name: "Downloading archive", Detail: "main", State: terminalexperience.PhaseActive}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tracked phases = %#v, want %#v", got, want)
	}
	outcome := operations[3].Value.(terminalexperience.PresentationDocument)
	if got, want := outcome.Blocks, []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleMuted, Text: "Resolved: github.example/group/project (github)"},
		{Role: terminalexperience.VisualRoleActive, Text: "Branch: main"},
		{Role: terminalexperience.VisualRoleSuccess, Text: "Archive downloaded and extracted"},
		{Role: terminalexperience.VisualRoleSuccess, Text: "Done! Project created at chosen"},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("outcome = %#v, want %#v", got, want)
	}
}

func TestTerminalGitForkAdapterPreservesPlainConfirmationSyntax(t *testing.T) {
	prompt := OverwritePrompt{Destination: "project", Message: "Directory \"project\" is not empty. Overwrite?"}
	for _, testCase := range []struct {
		name          string
		input         string
		wantConfirmed bool
		wantCancelled bool
		invalid       bool
	}{
		{name: "default yes", input: "\n", wantConfirmed: true},
		{name: "declined", input: "no\n"},
		{name: "cancelled", input: "cancel\n", wantCancelled: true},
		{name: "invalid then confirmed", input: "unexpected\nyes\n", wantConfirmed: true, invalid: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stdout, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
			experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
				Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
				Input:       strings.NewReader(testCase.input),
				Output:      stdout,
				Diagnostics: diagnostics,
			})
			run := experience.Open(context.Background())
			confirmed, cancelled, err := newTerminalGitForkAdapter(run, experience.Session(), func() {}).ConfirmOverwrite(prompt)
			if closeErr := run.Close(); closeErr != nil {
				t.Fatalf("Close() error = %v", closeErr)
			}
			if err != nil || confirmed != testCase.wantConfirmed || cancelled != testCase.wantCancelled {
				t.Fatalf("ConfirmOverwrite() = (%t, %t, %v)", confirmed, cancelled, err)
			}
			if !strings.HasPrefix(diagnostics.String(), prompt.Message+" [Y/n]: ") {
				t.Fatalf("confirmation prompt = %q", diagnostics.String())
			}
			if testCase.invalid != strings.Contains(diagnostics.String(), "Invalid confirmation") {
				t.Fatalf("confirmation diagnostics = %q", diagnostics.String())
			}
			if stdout.Len() != 0 || terminaltest.ContainsTerminalControl(diagnostics.Bytes()) {
				t.Fatalf("Plain streams = (%q, %q)", stdout.String(), diagnostics.String())
			}
		})
	}
}

func TestTerminalGitForkAdapterMapsCancellationAutomationAndRedactsFallbackFacts(t *testing.T) {
	cancelledExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrInteractionCancelled})
	cancelledAdapter := newTerminalGitForkAdapter(cancelledExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.RichInteractive}, func() {})
	if confirmed, cancelled, err := cancelledAdapter.ConfirmOverwrite(OverwritePrompt{}); err != nil || confirmed || !cancelled {
		t.Fatalf("cancelled ConfirmOverwrite() = (%t, %t, %v)", confirmed, cancelled, err)
	}

	automationExperience := terminaltest.NewRecordingExperience(terminaltest.SemanticAnswer{Err: terminalexperience.ErrAutomationInteraction})
	automationAdapter := newTerminalGitForkAdapter(automationExperience.Open(context.Background()), terminalexperience.Session{Kind: terminalexperience.Automation}, func() {})
	if _, _, err := automationAdapter.ConfirmOverwrite(OverwritePrompt{}); !errors.Is(err, errGitForkRequiresInteractive) {
		t.Fatalf("Automation ConfirmOverwrite() error = %v", err)
	}

	document := gitForkOutcomeDocument(terminalexperience.Session{Kind: terminalexperience.PlainInteractive}, Result{
		Repository:         Repository{Host: "github.example", Owner: "group", Name: "project", ProviderType: "github"},
		Destination:        "chosen",
		DefaultBranchError: errors.New("default request failed: token=branch-secret"),
		ArchiveError:       errors.New("archive request failed: Authorization: Bearer archive-secret"),
		Acquisition:        "clone",
	})
	for _, block := range document.Blocks {
		if block.Role != terminalexperience.VisualRolePlain || strings.Contains(block.Text, "branch-secret") || strings.Contains(block.Text, "archive-secret") {
			t.Fatalf("redacted fallback block = %#v", block)
		}
	}
	if text := terminalexperience.RenderPlain(document); !strings.Contains(text, "[REDACTED]") {
		t.Fatalf("redacted fallback output = %q", text)
	}
}

func TestOSForkGitRunnerUsesTheSharedArgvProcessBoundary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("host shell fixture is Unix-specific")
	}
	root := t.TempDir()
	argumentsPath := filepath.Join(root, "arguments")
	t.Setenv("FORK_GIT_ARGUMENTS", argumentsPath)
	script := writeForkGitScript(t, root, "git", "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$FORK_GIT_ARGUMENTS\"\nprintf 'fork stderr' >&2\n")
	runner := gitRunnerAdapter{runner: &gitprocess.Runner{Executable: script}}

	output, err := runner.Run(context.Background(), []string{"clone", "--depth=1", "remote", "/tmp/destination"})
	if err != nil || output.ExitCode != 0 || string(output.Stderr) != "fork stderr" {
		t.Fatalf("Run() = (%#v, %v)", output, err)
	}
	arguments, err := os.ReadFile(argumentsPath)
	if err != nil {
		t.Fatalf("read arguments: %v", err)
	}
	if got, want := string(arguments), "clone\n--depth=1\nremote\n/tmp/destination\n"; got != want {
		t.Fatalf("arguments = %q, want %q", got, want)
	}

	missing := gitRunnerAdapter{runner: &gitprocess.Runner{Executable: filepath.Join(root, "missing-git")}}
	if _, err := missing.Run(context.Background(), nil); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing runner error = %v", err)
	}
}

func writeForkGitScript(t *testing.T, directory, name, contents string) string {
	t.Helper()
	path := filepath.Join(directory, name+".sh")
	if err := os.WriteFile(path, []byte(contents), 0o700); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}

func TestGitForkHandlerUsesEncryptedConfigAndALocalProviderArchive(t *testing.T) {
	archive := gitForkFixtureArchive(t, map[string]string{
		"project-main/README.md": "archive contents\n",
		"project-main/bin/run":   "#!/bin/sh\n",
	})
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v3/repos/group/project":
			if request.Header.Get("Authorization") != "Bearer fixture-token" {
				t.Errorf("default-branch Authorization = %q", request.Header.Get("Authorization"))
			}
			_, _ = io.WriteString(response, `{"default_branch":"main"}`)
		case "/api/v3/repos/group/project/tarball/main":
			if request.Header.Get("Authorization") != "Bearer fixture-token" {
				t.Errorf("archive Authorization = %q", request.Header.Get("Authorization"))
			}
			_, _ = response.Write(archive)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")
	configureGitForkFixture(t, home, server.URL)
	destination := filepath.Join(t.TempDir(), "destination")
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatalf("create destination: %v", err)
	}
	if err := os.WriteFile(filepath.Join(destination, "old.txt"), []byte("replace"), 0o600); err != nil {
		t.Fatalf("write existing destination: %v", err)
	}
	output, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("unexpected\nyes\n"),
		Output:      output,
		Diagnostics: diagnostics,
	})
	result, err := executeForkForTest(context.Background(), experience, "fixture:group/project", destination)
	if err != nil || string(result.Acquisition) != "archive" || result.Ref != "main" {
		t.Fatalf("handler result = (%#v, %v)", result, err)
	}
	if contents, err := os.ReadFile(filepath.Join(destination, "README.md")); err != nil || string(contents) != "archive contents\n" {
		t.Fatalf("archive contents = %q, %v", contents, err)
	}
	if info, err := os.Stat(filepath.Join(destination, "bin", "run")); err != nil || info.Mode()&0o100 != 0 {
		t.Fatalf("archive mode = (%v, %v), want no executable bit", info, err)
	}
	text := output.String()
	for _, expected := range []string{"HACKYCY CLI", "Git Fork", "Resolved:", "Branch: main", "Archive downloaded and extracted", "Done! Project created at"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, text)
		}
	}
	if strings.Contains(text, "fixture-token") {
		t.Fatalf("output disclosed the configured token: %q", text)
	}
	for _, expected := range []string{"Directory \"", "Invalid confirmation", "Resolving repository", "Fetching default branch", "Downloading archive", "Project ready"} {
		if !strings.Contains(diagnostics.String(), expected) {
			t.Fatalf("diagnostics does not contain %q:\n%s", expected, diagnostics.String())
		}
	}
	if terminaltest.ContainsTerminalControl(append(output.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("Plain streams contain terminal control: (%q, %q)", output.String(), diagnostics.String())
	}
}

func TestGitForkHandlerLeavesANonemptyDestinationOnDeclinedOverwrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")
	destination := filepath.Join(t.TempDir(), "destination")
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatalf("create destination: %v", err)
	}
	kept := filepath.Join(destination, "kept.txt")
	if err := os.WriteFile(kept, []byte("keep"), 0o600); err != nil {
		t.Fatalf("write destination: %v", err)
	}
	output, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader("no\n"),
		Output:      output,
		Diagnostics: diagnostics,
	})
	result, err := executeForkForTest(context.Background(), experience, "owner/project", destination)
	if err != nil || !result.Cancelled {
		t.Fatalf("handler result = (%#v, %v)", result, err)
	}
	if contents, err := os.ReadFile(kept); err != nil || string(contents) != "keep" {
		t.Fatalf("kept destination = %q, %v", contents, err)
	}
	if !strings.Contains(output.String(), "Cancelled") {
		t.Fatalf("cancellation output = %q", output.String())
	}
	if !strings.Contains(diagnostics.String(), "[Y/n]: ") || terminaltest.ContainsTerminalControl(append(output.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("Plain cancellation streams = (%q, %q)", output.String(), diagnostics.String())
	}
}

func TestExecuteForkAutomationFailsBeforeReadingOrRemovingANonemptyDestination(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "destination")
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatalf("create destination: %v", err)
	}
	kept := filepath.Join(destination, "kept.txt")
	if err := os.WriteFile(kept, []byte("keep"), 0o600); err != nil {
		t.Fatalf("write destination: %v", err)
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.Automation},
		Input:       panicGitForkReader{},
		Output:      stdout,
		Diagnostics: stderr,
	})
	_, err := executeForkForTest(context.Background(), experience, "owner/project", destination)
	if !errors.Is(err, errGitForkRequiresInteractive) || stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("Automation error = %v, streams = (%q, %q)", err, stdout.String(), stderr.String())
	}
	if contents, err := os.ReadFile(kept); err != nil || string(contents) != "keep" {
		t.Fatalf("Automation failure mutated destination = (%q, %v)", contents, err)
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("Automation streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
	}
}

func TestGitForkHandlerFallsBackToTheLocalGitRunnerAndCleansMetadata(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("host shell fixture is Unix-specific")
	}
	root := t.TempDir()
	argumentsPath := filepath.Join(root, "arguments")
	script := filepath.Join(root, "git")
	scriptContents := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$FORK_GIT_ARGUMENTS\"\nfor value do destination=$value; done\nmkdir -p \"$destination/.git\"\nprintf 'clone contents\\n' > \"$destination/README.md\"\n"
	if err := os.WriteFile(script, []byte(scriptContents), 0o700); err != nil {
		t.Fatalf("write Git fixture: %v", err)
	}
	t.Setenv("FORK_GIT_ARGUMENTS", argumentsPath)
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v3/repos/group/project":
			_, _ = io.WriteString(response, `{"default_branch":"main"}`)
		case "/api/v3/repos/group/project/tarball/main":
			response.WriteHeader(http.StatusBadGateway)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")
	configureGitForkFixture(t, home, server.URL)
	destination := filepath.Join(root, "destination")
	output, diagnostics := &bytes.Buffer{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session:     terminalexperience.Session{Kind: terminalexperience.PlainInteractive},
		Input:       strings.NewReader(""),
		Output:      output,
		Diagnostics: diagnostics,
	})
	result, err := executeForkForTest(context.Background(), experience, "fixture:group/project", destination)
	if err != nil || string(result.Acquisition) != "clone" {
		t.Fatalf("handler result = (%#v, %v)", result, err)
	}
	if contents, err := os.ReadFile(filepath.Join(destination, "README.md")); err != nil || string(contents) != "clone contents\n" {
		t.Fatalf("clone contents = %q, %v", contents, err)
	}
	if _, err := os.Lstat(filepath.Join(destination, ".git")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("clone metadata remained: %v", err)
	}
	arguments, err := os.ReadFile(argumentsPath)
	if err != nil {
		t.Fatalf("read Git arguments: %v", err)
	}
	if got := strings.Split(strings.TrimSpace(string(arguments)), "\n"); len(got) != 7 || got[0] != "clone" || got[1] != "--depth=1" || got[2] != "--single-branch" || got[3] != "--branch" || got[4] != "main" || got[6] != destination {
		t.Fatalf("clone arguments = %q", arguments)
	}
	if !strings.Contains(output.String(), "Archive download failed") || !strings.Contains(output.String(), "Cloned and cleaned up") || strings.Contains(output.String(), "fixture-token") {
		t.Fatalf("clone fallback output = %q", output.String())
	}
	if !strings.Contains(diagnostics.String(), "Falling back to git clone") || terminaltest.ContainsTerminalControl(append(output.Bytes(), diagnostics.Bytes()...)) {
		t.Fatalf("Plain clone streams = (%q, %q)", output.String(), diagnostics.String())
	}
}

func configureGitForkFixture(t *testing.T, home, serverURL string) {
	t.Helper()
	store, err := appconfig.New(appconfig.Dependencies{
		Environment: func(key string) string {
			switch key {
			case "HOME":
				return home
			case "USERPROFILE":
				return ""
			default:
				return os.Getenv(key)
			}
		},
	})
	if err != nil {
		t.Fatalf("new appconfig store: %v", err)
	}
	if err := store.SaveForkInstance("fixture", appconfig.ForkInput{
		Host: strings.TrimPrefix(serverURL, "http://"), Scheme: "http", Type: "github", Token: "fixture-token",
	}); err != nil {
		t.Fatalf("save Fork fixture: %v", err)
	}
}

func executeForkForTest(ctx context.Context, experience *terminalexperience.Runtime, repository, destination string) (Result, error) {
	return executeFork(&Options{
		Context:          ctx,
		Repository:       repository,
		Destination:      destination,
		Config:           forkTestConfigStore,
		WorkingDirectory: os.Getwd,
		HTTP:             http.DefaultClient,
		Terminal:         experience,
		Git:              &gitprocess.Runner{},
	})
}

func forkTestConfigStore() (ConfigReader, error) {
	return appconfig.New(appconfig.Dependencies{})
}

func gitForkFixtureArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(gzipWriter)
	for name, contents := range files {
		if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(contents))}); err != nil {
			t.Fatalf("write TAR header: %v", err)
		}
		if _, err := io.WriteString(tarWriter, contents); err != nil {
			t.Fatalf("write TAR contents: %v", err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close TAR: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return compressed.Bytes()
}

type panicGitForkReader struct{}

func (panicGitForkReader) Read([]byte) (int, error) {
	panic("git fork Automation must not read stdin")
}
