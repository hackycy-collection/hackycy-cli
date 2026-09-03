package test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestRunTestRichPTYRestoresPrimaryScreenAndReplaysSafeTranscript(t *testing.T) {
	const helperEnvironment = "YCY_CONFIG_CM_TEST_RICH_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		runCMTestRichPTYHelper(t)
		return
	}

	for _, testCase := range []struct {
		name  string
		extra string
		color bool
	}{
		{name: "color", color: true},
		{name: "no color", extra: "NO_COLOR=1", color: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestRunTestRichPTYRestoresPrimaryScreenAndReplaysSafeTranscript$")
			command.Env = append(cmTestPTYEnvironment(), helperEnvironment+"=1", "TERM=xterm-256color")
			if testCase.extra != "" {
				command.Env = append(command.Env, testCase.extra)
			}
			output := runCMTestPTYProcess(t, command)
			assertCMTestRichPTYOutput(t, output, testCase.color)
		})
	}
}

func runCMTestRichPTYHelper(t *testing.T) {
	t.Helper()
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{
			Interaction: terminalexperience.RichInteractive,
			Stdin:       terminalexperience.StreamCapability{Terminal: true},
			Stdout:      terminalexperience.StreamCapability{Terminal: true, Color: os.Getenv("NO_COLOR") == ""},
			Stderr:      terminalexperience.StreamCapability{Terminal: true, Color: os.Getenv("NO_COLOR") == ""},
		},
		Input:       os.Stdin,
		Output:      os.Stdout,
		Diagnostics: os.Stderr,
	})
	err := runTest(&Options{
		Context: context.Background(),
		Store: func() (TestProfileResolver, error) {
			return &recordingCMTestResolver{profile: cmTestProfile()}, nil
		},
		HTTP: cmTestHTTPClient(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"ok"}}],"usage":{"prompt_tokens":3,"completion_tokens":2}}`)),
			}, nil
		}),
		Terminal: experience,
	})
	if err != nil {
		t.Fatalf("runTest() error = %v", err)
	}
}

func runCMTestPTYProcess(t *testing.T, command *exec.Cmd) string {
	t.Helper()
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY helper: %v", err)
	}
	defer process.Close()
	var output lockedCMTestPTYBuffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, process.Terminal())
		close(readDone)
	}()
	if err := process.Wait(); err != nil {
		t.Fatalf("wait PTY helper: %v\n%s", err, output.String())
	}
	if err := process.Close(); err != nil {
		t.Fatalf("close PTY helper: %v", err)
	}
	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out reading PTY output: %q", output.String())
	}
	return output.String()
}

func assertCMTestRichPTYOutput(t *testing.T, output string, color bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	if !strings.Contains(visible, "YCY / config cm test") || !strings.Contains(visible, "Resolve CM test profile") || !strings.Contains(visible, "Test CM provider") || !strings.Contains(visible, "Response received") {
		t.Fatalf("Rich PTY output missing phase/context evidence: %q", output)
	}
	for _, expected := range []string{"Test commit message provider", "Response:\nok", "Done", "Prompt tokens: 3", "Completion tokens: 2", "Total tokens: 5"} {
		if !strings.Contains(visible, expected) {
			t.Fatalf("Rich PTY output missing %q: %q", expected, output)
		}
	}
	enter := strings.LastIndex(visible, "\x1b[?1049h")
	leave := strings.LastIndex(visible, "\x1b[?1049l")
	if strings.Count(visible, "\x1b[?1049h") != 1 || strings.Count(visible, "\x1b[?1049l") != 1 || enter < 0 || leave < enter || !strings.Contains(visible, "\x1b[?25h") {
		t.Fatalf("Rich PTY output did not restore the primary screen: %q", output)
	}
	transcript := strings.Index(visible[leave:], "Resolve CM test profile (completed)")
	result := strings.LastIndex(visible, "Response:\nok")
	if transcript < 0 || result < 0 || leave+transcript > result {
		t.Fatalf("Rich PTY transcript/result ordering = %q", output)
	}
	if strings.Contains(output, "test-api-key") {
		t.Fatalf("Rich PTY output leaked API key: %q", output)
	}
	if !color {
		for _, prefix := range []string{"\x1b[38;", "\x1b[3m", "\x1b[9m"} {
			if strings.Contains(output, prefix) {
				t.Fatalf("no-color Rich PTY output contains %q: %q", prefix, output)
			}
		}
	}
}

func cmTestPTYEnvironment() []string {
	ignored := map[string]struct{}{"CI": {}, "CLICOLOR": {}, "CLICOLOR_FORCE": {}, "COLORTERM": {}, "NO_COLOR": {}, "TERM": {}}
	environment := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if _, skip := ignored[key]; !skip {
			environment = append(environment, entry)
		}
	}
	return environment
}

type lockedCMTestPTYBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (buffer *lockedCMTestPTYBuffer) Write(value []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.Write(value)
}

func (buffer *lockedCMTestPTYBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.String()
}
