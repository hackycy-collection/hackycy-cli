package cm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/gitprocess"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestGitCMRichPTYFourWayStageAndCommitJourney(t *testing.T) {
	const helperEnvironment = "YCY_GIT_CM_STAGE_COMMIT_PTY_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		runGitCMStageCommitPTYHelper(t)
		return
	}
	for _, testCase := range []struct {
		name          string
		width, height uint16
		color         bool
	}{
		{name: "wide color", width: 120, height: 40, color: true},
		{name: "wide no color", width: 120, height: 40, color: false},
		{name: "compact color", width: 40, height: 15, color: true},
		{name: "compact no color", width: 40, height: 15, color: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			providerDone := filepath.Join(t.TempDir(), "provider-done")
			command := exec.Command(os.Args[0], "-test.run=^TestGitCMRichPTYFourWayStageAndCommitJourney$")
			command.Env = gitCMPTYEnvironment(map[string]string{
				"NO_COLOR":                          map[bool]string{true: "", false: "1"}[testCase.color],
				"TERM":                              "xterm-256color",
				helperEnvironment:                   "1",
				"YCY_GIT_CM_STAGE_COMMIT_PTY_START": "1",
				"YCY_GIT_CM_PROVIDER_DONE":          providerDone,
			})
			output := runGitCMPTYProcess(t, command, testCase.width, testCase.height, providerDone)
			assertGitCMStageCommitPTYOutput(t, output, testCase.color, testCase.width >= 70)
		})
	}
}

func TestGitCMRichPTYFourWayPushJourney(t *testing.T) {
	const helperEnvironment = "YCY_GIT_CM_PUSH_PTY_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		runGitCMPushPTYHelper(t)
		return
	}
	for _, testCase := range []struct {
		name          string
		width, height uint16
		color         bool
	}{
		{name: "wide color", width: 120, height: 40, color: true},
		{name: "wide no color", width: 120, height: 40, color: false},
		{name: "compact color", width: 40, height: 15, color: true},
		{name: "compact no color", width: 40, height: 15, color: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			providerDone := filepath.Join(t.TempDir(), "provider-done")
			command := exec.Command(os.Args[0], "-test.run=^TestGitCMRichPTYFourWayPushJourney$")
			command.Env = gitCMPTYEnvironment(map[string]string{
				"NO_COLOR":                  map[bool]string{true: "", false: "1"}[testCase.color],
				"TERM":                      "xterm-256color",
				helperEnvironment:           "1",
				"YCY_GIT_CM_PUSH_PTY_START": "1",
				"YCY_GIT_CM_PROVIDER_DONE":  providerDone,
			})
			output := runGitCMPushPTYProcess(t, command, testCase.width, testCase.height, providerDone)
			assertGitCMPushPTYOutput(t, output, testCase.color, testCase.width >= 70)
		})
	}
}

func runGitCMPushPTYHelper(t *testing.T) {
	t.Helper()
	if os.Getenv("YCY_GIT_CM_PUSH_PTY_START") == "1" {
		var start [2]byte
		if _, err := io.ReadFull(os.Stdin, start[:]); err != nil {
			t.Fatalf("wait for PTY sizing: %v", err)
		}
	}
	repository := newGitCMRepository(t)
	withGitCMWorkingDirectory(t, repository)
	remote := filepath.Join(t.TempDir(), "origin.git")
	if err := os.MkdirAll(remote, 0o700); err != nil {
		t.Fatalf("create Git CM push remote: %v", err)
	}
	runGitCM(t, remote, "init", "--bare", "-q")
	runGitCM(t, repository, "remote", "add", "origin", remote)
	runGitCM(t, repository, "push", "-q", "-u", "origin", "HEAD")
	writeGitCMFile(t, filepath.Join(repository, "README.md"), "rich push\n")
	runGitCM(t, repository, "add", "README.md")
	providerDone := os.Getenv("YCY_GIT_CM_PROVIDER_DONE")
	providerCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		providerCalls++
		if request.Method != http.MethodPost || request.URL.Path != "/chat/completions" {
			t.Errorf("provider request = %s %s", request.Method, request.URL.Path)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = io.Copy(io.Discard, request.Body)
		response.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(response, `{"choices":[{"message":{"content":%q}}]}`, "feat(cm): rich push")
		if providerDone != "" {
			_ = os.WriteFile(providerDone, []byte("done"), 0o600)
		}
	}))
	defer server.Close()
	configureGitCMProvider(t, server.URL)
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{
			Interaction: terminalexperience.RichInteractive,
			Stdin:       terminalexperience.StreamCapability{Terminal: true},
			Stdout:      terminalexperience.StreamCapability{Terminal: true, Color: os.Getenv("NO_COLOR") == ""},
			Stderr:      terminalexperience.StreamCapability{Terminal: true, Color: os.Getenv("NO_COLOR") == ""},
		},
		Input: os.Stdin, Output: os.Stdout, Diagnostics: os.Stderr,
	})
	remoteName := "origin"
	result, err := executeCM(&Options{Context: context.Background(), Input: Input{Staged: true, Push: &remoteName}, Config: func() (ProfileResolver, error) {
		return appconfig.New(appconfig.Dependencies{})
	}, HTTP: server.Client(), Terminal: experience, Git: &gitprocess.Runner{}})
	if err != nil || !result.Committed || !result.Pushed || providerCalls != 1 {
		t.Fatalf("executeCM() = (%#v, %v), provider calls = %d", result, err, providerCalls)
	}
	if status := gitCMOutput(t, repository, "status", "--short"); status != "" {
		t.Fatalf("working tree after push = %q", status)
	}
	_, _ = fmt.Fprintln(os.Stderr, "GIT_CM_PUSH_OK")
}

func runGitCMPushPTYProcess(t *testing.T, command *exec.Cmd, width, height uint16, providerDone string) string {
	t.Helper()
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start Git CM push PTY helper: %v", err)
	}
	defer process.Close()
	if err := process.Resize(width, height); err != nil {
		t.Fatalf("resize PTY to %dx%d: %v", width, height, err)
	}
	var output gitCMPTYBuffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, process.Terminal())
		close(readDone)
	}()
	if _, err := process.Terminal().Write([]byte("go\n")); err != nil {
		t.Fatalf("release Git CM push PTY helper after sizing: %v", err)
	}
	waitForGitCMPTYFile(t, providerDone)
	time.Sleep(150 * time.Millisecond)
	if _, err := process.Terminal().Write([]byte("\r")); err != nil {
		t.Fatalf("submit Git CM push confirmation: %v", err)
	}
	waitForGitCMPTYText(t, &output, "GIT_CM_PUSH_OK")
	if err := process.Wait(); err != nil {
		t.Fatalf("wait Git CM push PTY helper: %v\n%s", err, output.String())
	}
	if err := process.Close(); err != nil {
		t.Fatalf("close Git CM push PTY helper: %v", err)
	}
	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out reading Git CM push PTY output: %q", output.String())
	}
	return output.String()
}

func assertGitCMPushPTYOutput(t *testing.T, output string, color, wide bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	enter := strings.Index(visible, "\x1b[?1049h")
	leave := strings.LastIndex(visible, "\x1b[?1049l")
	if strings.Count(visible, "\x1b[?1049h") != 1 || strings.Count(visible, "\x1b[?1049l") != 1 || enter < 0 || leave < enter || !strings.Contains(visible, "\x1b[?25h") {
		t.Fatalf("git cm push did not restore primary screen: %q", output)
	}
	live := strings.Join(strings.Fields(terminaltest.StripANSI(visible[enter:leave])), " ")
	for _, expected := range []string{"YCY / git cm", "STATE", "PHASE", "DETAIL"} {
		if !strings.Contains(live, expected) {
			t.Fatalf("git cm push live Console missing %q: %q", expected, output)
		}
	}
	if wide && !strings.Contains(live, "Push commit") {
		t.Fatalf("wide git cm push live Console omitted push phase: %q", output)
	}
	for _, expected := range []string{"feat(cm): rich push", "GIT_CM_PUSH_OK", "Commit created and pushed"} {
		if !strings.Contains(visible, expected) {
			t.Fatalf("git cm push output missing %q: %q", expected, output)
		}
	}
	if strings.Contains(output, "fixture-api-key") || strings.Contains(output, "Authorization: Bearer") {
		t.Fatalf("git cm push leaked provider credential: %q", output)
	}
	transcript := visible[leave:]
	for _, expected := range []string{"Verify unchanged scope (completed)", "Create commit (completed)", "Push commit (completed)", "succeeded", "Commit created and pushed"} {
		if !strings.Contains(transcript, expected) {
			t.Fatalf("git cm push Transcript missing %q: %q", expected, output)
		}
	}
	if strings.Index(transcript, "Create commit (completed)") > strings.Index(transcript, "Push commit (completed)") {
		t.Fatalf("git cm push phase ordering = %q", transcript)
	}
	if !color {
		for _, prefix := range []string{"\x1b[38;", "\x1b[3m", "\x1b[9m"} {
			if strings.Contains(output, prefix) {
				t.Fatalf("NO_COLOR git cm push output contains %q: %q", prefix, output)
			}
		}
	}
}

func runGitCMStageCommitPTYHelper(t *testing.T) {
	t.Helper()
	if os.Getenv("YCY_GIT_CM_STAGE_COMMIT_PTY_START") == "1" {
		var start [2]byte
		if _, err := io.ReadFull(os.Stdin, start[:]); err != nil {
			t.Fatalf("wait for PTY sizing: %v", err)
		}
	}
	repository := newGitCMRepository(t)
	withGitCMWorkingDirectory(t, repository)
	writeGitCMFile(t, repository+"/README.md", "rich stage and commit\n")
	providerDone := os.Getenv("YCY_GIT_CM_PROVIDER_DONE")
	providerCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		providerCalls++
		if request.Method != http.MethodPost || request.URL.Path != "/chat/completions" {
			t.Errorf("provider request = %s %s", request.Method, request.URL.Path)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = io.Copy(io.Discard, request.Body)
		response.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(response, `{"choices":[{"message":{"content":%q}}]}`, "feat(cm): rich stage and commit")
		if providerDone != "" {
			_ = os.WriteFile(providerDone, []byte("done"), 0o600)
		}
	}))
	defer server.Close()
	configureGitCMProvider(t, server.URL)
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{
			Interaction: terminalexperience.RichInteractive,
			Stdin:       terminalexperience.StreamCapability{Terminal: true},
			Stdout:      terminalexperience.StreamCapability{Terminal: true, Color: os.Getenv("NO_COLOR") == ""},
			Stderr:      terminalexperience.StreamCapability{Terminal: true, Color: os.Getenv("NO_COLOR") == ""},
		},
		Input: os.Stdin, Output: os.Stdout, Diagnostics: os.Stderr,
	})
	result, err := executeCM(&Options{Context: context.Background(), Input: Input{Stage: true}, Config: func() (ProfileResolver, error) {
		return appconfig.New(appconfig.Dependencies{})
	}, HTTP: server.Client(), Terminal: experience, Git: &gitprocess.Runner{}})
	if err != nil || !result.Committed || result.Pushed || providerCalls != 1 {
		t.Fatalf("executeCM() = (%#v, %v), provider calls = %d", result, err, providerCalls)
	}
	if status := gitCMOutput(t, repository, "status", "--short"); status != "" {
		t.Fatalf("working tree after commit = %q", status)
	}
	if subject := strings.TrimSpace(gitCMOutput(t, repository, "log", "-1", "--format=%s")); subject != "feat(cm): rich stage and commit" {
		t.Fatalf("commit subject = %q", subject)
	}
	_, _ = fmt.Fprintln(os.Stderr, "GIT_CM_STAGE_COMMIT_OK")
}

func assertGitCMStageCommitPTYOutput(t *testing.T, output string, color, wide bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	enter := strings.Index(visible, "\x1b[?1049h")
	leave := strings.LastIndex(visible, "\x1b[?1049l")
	if strings.Count(visible, "\x1b[?1049h") != 1 || strings.Count(visible, "\x1b[?1049l") != 1 || enter < 0 || leave < enter || !strings.Contains(visible, "\x1b[?25h") {
		t.Fatalf("git cm Rich PTY did not restore primary screen: %q", output)
	}
	live := strings.Join(strings.Fields(terminaltest.StripANSI(visible[enter:leave])), " ")
	for _, expected := range []string{"YCY / git cm", "STATE", "PHASE", "DETAIL"} {
		if !strings.Contains(live, expected) {
			t.Fatalf("git cm live Console missing %q: %q", expected, output)
		}
	}
	if wide {
		for _, expected := range []string{"Generate and optionally create a commit", "stage and commit", "Inspect changes", "Capture commit evidence", "Generate commit message", "Verify unchanged scope", "Create commit"} {
			if !strings.Contains(live, expected) {
				t.Fatalf("wide git cm live Console missing %q: %q", expected, output)
			}
		}
	} else if !strings.Contains(live, "selection") && !strings.Contains(live, "commit") {
		t.Fatalf("compact git cm live Console omitted bounded active context: %q", output)
	}
	for _, expected := range []string{"Select files to stage", "feat(cm): rich stage and commit", "GIT_CM_STAGE_COMMIT_OK", "Commit created"} {
		if !strings.Contains(visible, expected) {
			t.Fatalf("git cm PTY output missing %q: %q", expected, output)
		}
	}
	if wide && !strings.Contains(visible, "Create this commit?") {
		t.Fatalf("wide git cm PTY omitted commit confirmation: %q", output)
	}
	if strings.Contains(output, "fixture-api-key") || strings.Contains(output, "Authorization: Bearer") || strings.Contains(output, "http://127.") {
		t.Fatalf("git cm PTY leaked provider credential or URL: %q", output)
	}
	transcript := visible[leave:]
	ordered := []string{"Inspect changes (completed)", "Stage selected files (completed)", "Capture commit evidence (completed)", "Resolve provider profile (completed)", "Generate commit message (completed)", "Verify unchanged scope (completed)", "Create commit (completed)", "succeeded", "Commit created"}
	last := 0
	for _, expected := range ordered {
		next := strings.Index(transcript[last:], expected)
		if next < 0 {
			t.Fatalf("git cm Transcript missing ordered event %q: %q", expected, output)
		}
		last += next + len(expected)
	}
	if !color {
		for _, prefix := range []string{"\x1b[38;", "\x1b[3m", "\x1b[9m"} {
			if strings.Contains(output, prefix) {
				t.Fatalf("NO_COLOR git cm PTY output contains %q: %q", prefix, output)
			}
		}
	}
}

func runGitCMPTYProcess(t *testing.T, command *exec.Cmd, width, height uint16, providerDone string) string {
	t.Helper()
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start git cm PTY helper: %v", err)
	}
	defer process.Close()
	if err := process.Resize(width, height); err != nil {
		t.Fatalf("resize PTY to %dx%d: %v", width, height, err)
	}
	var output gitCMPTYBuffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, process.Terminal())
		close(readDone)
	}()
	if _, err := process.Terminal().Write([]byte("go\n")); err != nil {
		t.Fatalf("release PTY helper after sizing: %v", err)
	}
	waitForGitCMPTYText(t, &output, "Select files to stage")
	if _, err := process.Terminal().Write([]byte("\r")); err != nil {
		t.Fatalf("commit Git CM selection filter: %v", err)
	}
	// Huh's searchable MultiSelect may need a second Enter. Wait for the
	// command-owned staging phase first; this avoids sending a buffered Enter
	// into the subsequent confirmation form on compact terminals.
	selectionDone := false
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		text := output.String()
		if strings.Contains(text, "Stage selected files") || strings.Contains(text, "files staged") {
			selectionDone = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !selectionDone {
		if _, err := process.Terminal().Write([]byte("\r")); err != nil {
			t.Fatalf("submit Git CM file selection: %v", err)
		}
	}
	waitForGitCMPTYFile(t, providerDone)
	time.Sleep(100 * time.Millisecond)
	if _, err := process.Terminal().Write([]byte("\r")); err != nil {
		t.Fatalf("submit Git CM confirmation: %v", err)
	}
	waitForGitCMPTYText(t, &output, "GIT_CM_STAGE_COMMIT_OK")
	if err := process.Wait(); err != nil {
		t.Fatalf("wait git cm PTY helper: %v\n%s", err, output.String())
	}
	if err := process.Close(); err != nil {
		t.Fatalf("close git cm PTY helper: %v", err)
	}
	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out reading git cm PTY output: %q", output.String())
	}
	return output.String()
}

func waitForGitCMPTYFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for Git CM provider marker %q", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitForGitCMPTYText(t *testing.T, output *gitCMPTYBuffer, needle string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if strings.Contains(output.String(), needle) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for git cm PTY text %q: %q", needle, output.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

type gitCMPTYBuffer struct {
	mu    sync.Mutex
	value strings.Builder
}

func (buffer *gitCMPTYBuffer) Write(value []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.value.Write(value)
}

func (buffer *gitCMPTYBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.value.String()
}

func gitCMPTYEnvironment(overrides map[string]string) []string {
	ignored := map[string]struct{}{"CI": {}, "CLICOLOR": {}, "CLICOLOR_FORCE": {}, "COLORTERM": {}, "NO_COLOR": {}, "TERM": {}}
	environment := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		if _, skip := ignored[key]; skip {
			continue
		}
		if _, replaced := overrides[key]; !replaced {
			environment = append(environment, entry)
		}
	}
	for key, value := range overrides {
		environment = append(environment, key+"="+value)
	}
	return environment
}
