package main

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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	forkcommand "github.com/hackycy/hackycy-cli/internal/commands/git/fork"
)

func TestTerminalGitForkPrompterUsesTheLegacyDefaultYesAndCancellationForms(t *testing.T) {
	prompt := forkcommand.OverwritePrompt{Destination: "project", Message: "Directory \"project\" is not empty. Overwrite?"}

	defaultOutput := &bytes.Buffer{}
	confirmed, cancelled := newTerminalGitForkPrompter(strings.NewReader("\n"), defaultOutput).ConfirmOverwrite(prompt)
	if !confirmed || cancelled || defaultOutput.String() != "Directory \"project\" is not empty. Overwrite? [Y/n]: " {
		t.Fatalf("default overwrite = (%t, %t, %q)", confirmed, cancelled, defaultOutput.String())
	}

	for _, test := range []struct {
		input         string
		wantConfirmed bool
		wantCancelled bool
	}{
		{input: "no\n"},
		{input: "cancel\n", wantCancelled: true},
		{input: "unexpected\nyes\n", wantConfirmed: true},
	} {
		output := &bytes.Buffer{}
		confirmed, cancelled := newTerminalGitForkPrompter(strings.NewReader(test.input), output).ConfirmOverwrite(prompt)
		if confirmed != test.wantConfirmed || cancelled != test.wantCancelled {
			t.Fatalf("input %q overwrite = (%t, %t)", test.input, confirmed, cancelled)
		}
		if strings.HasPrefix(test.input, "unexpected") && !strings.Contains(output.String(), "Invalid confirmation") {
			t.Fatalf("invalid confirmation output = %q", output.String())
		}
	}
}

func TestTerminalGitForkPresenterRendersTheCommandMilestones(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalGitForkPresenter{output: output}
	presenter.Introduction()
	presenter.Resolved(forkcommand.Repository{Host: "github.example", Owner: "group", Name: "project", ProviderType: "github"})
	presenter.DefaultBranchStarted()
	presenter.DefaultBranchResolved("main")
	presenter.ArchiveStarted()
	presenter.ArchiveSucceeded()
	presenter.Completed("chosen")
	presenter.DefaultBranchFailed(errors.New("default unavailable"))
	presenter.ArchiveFailed(errors.New("archive unavailable"))
	presenter.CloneStarted()
	presenter.CloneSucceeded()
	presenter.Cancelled()

	want := "HACKYCY CLI\n\nGit Fork\n" +
		"Resolved: github.example/group/project (github)\n" +
		"Fetching default branch...\nBranch: main\nDownloading archive...\nArchive downloaded and extracted\nDone! Project created at chosen\n" +
		"Failed to get default branch: default unavailable\nFalling back to git clone with remote default branch.\n" +
		"Archive download failed: archive unavailable\nFalling back to git clone...\nCloned and cleaned up\nCancelled\n"
	if output.String() != want {
		t.Fatalf("presentation = %q, want %q", output.String(), want)
	}
}

func TestOSForkGitRunnerUsesTheSharedArgvProcessBoundary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("host shell fixture is Unix-specific")
	}
	root := t.TempDir()
	argumentsPath := filepath.Join(root, "arguments")
	t.Setenv("FORK_GIT_ARGUMENTS", argumentsPath)
	script := writeHeatGitScript(t, root, "git", "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$FORK_GIT_ARGUMENTS\"\nprintf 'fork stderr' >&2\n")
	runner := &osForkGitRunner{executable: script}

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

	missing := &osForkGitRunner{executable: filepath.Join(root, "missing-git")}
	if _, err := missing.Run(context.Background(), nil); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing runner error = %v", err)
	}
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
	output := &bytes.Buffer{}
	handler := newGitForkHandler(strings.NewReader(""), output)
	result, err := handler(context.Background(), forkcommand.Input{Repository: "fixture:group/project", Destination: destination})
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
	output := &bytes.Buffer{}
	handler := newGitForkHandler(strings.NewReader("no\n"), output)
	result, err := handler(context.Background(), forkcommand.Input{Repository: "owner/project", Destination: destination})
	if err != nil || !result.Cancelled {
		t.Fatalf("handler result = (%#v, %v)", result, err)
	}
	if contents, err := os.ReadFile(kept); err != nil || string(contents) != "keep" {
		t.Fatalf("kept destination = %q, %v", contents, err)
	}
	if !strings.Contains(output.String(), "Cancelled") {
		t.Fatalf("cancellation output = %q", output.String())
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
	output := &bytes.Buffer{}
	result, err := newGitForkHandler(strings.NewReader(""), output)(context.Background(), forkcommand.Input{Repository: "fixture:group/project", Destination: destination})
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
}

func TestGitForkStandaloneBinaryDownloadsALocalProviderArchive(t *testing.T) {
	archive := gitForkFixtureArchive(t, map[string]string{"project-main/README.md": "standalone archive\n"})
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v3/repos/group/project":
			_, _ = io.WriteString(response, `{"default_branch":"main"}`)
		case "/api/v3/repos/group/project/tarball/main":
			_, _ = response.Write(archive)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	home := t.TempDir()
	configureGitForkFixture(t, home, server.URL)
	binary := standaloneBinaryOutputPath(filepath.Join(t.TempDir(), "ycy"))
	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/ycy")
	build.Dir = repositoryRoot(t)
	build.Env = environmentWith(map[string]string{"CGO_ENABLED": "0", "GOTOOLCHAIN": "go1.26.7", "GOWORK": "off"})
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build standalone binary: %v\n%s", err, output)
	}
	destination := filepath.Join(t.TempDir(), "destination")
	command := exec.Command(resolveStandaloneBinary(binary), "git", "fork", "fixture:group/project", destination)
	command.Dir = t.TempDir()
	command.Env = environmentWith(map[string]string{"HOME": home, "USERPROFILE": ""})
	output, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "Done! Project created at") || strings.Contains(string(output), "fixture-token") {
		t.Fatalf("standalone git fork = (%v, %q)", err, output)
	}
	if contents, err := os.ReadFile(filepath.Join(destination, "README.md")); err != nil || string(contents) != "standalone archive\n" {
		t.Fatalf("standalone archive contents = %q, %v", contents, err)
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
