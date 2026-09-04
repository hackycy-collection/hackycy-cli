package add

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestRunCMAddRichPTYRestoresScreenAndRedactsTranscript(t *testing.T) {
	const helperEnvironment = "YCY_CONFIG_CM_ADD_RICH_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		runCMAddRichPTYHelper(t)
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
			command := exec.Command(os.Args[0], "-test.run=^TestRunCMAddRichPTYRestoresScreenAndRedactsTranscript$")
			command.Env = append(cmAddPTYEnvironment(), helperEnvironment+"=1", "TERM=xterm-256color")
			if testCase.extra != "" {
				command.Env = append(command.Env, testCase.extra)
			}
			output := runCMAddPTYProcess(t, command)
			assertCMAddRichPTYOutput(t, output, testCase.color)
		})
	}
}

func runCMAddRichPTYHelper(t *testing.T) {
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
	err := runAdd(&Options{
		Context:  context.Background(),
		Terminal: experience,
		Store: func() (AddWriter, error) {
			return cmAddWriterFunc(func(_, _, _, _ string) error {
				_, _ = fmt.Fprintln(os.Stderr, "CM_ADD_WRITE_OK")
				return nil
			}), nil
		},
	})
	if err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}
}

func runCMAddPTYProcess(t *testing.T, command *exec.Cmd) string {
	t.Helper()
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY helper: %v", err)
	}
	defer process.Close()
	var output lockedCMAddPTYBuffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, process.Terminal())
		close(readDone)
	}()
	for _, step := range []struct {
		needle string
		input  string
	}{
		{needle: "Profile name", input: "work\r"},
		{needle: "OpenAI-compatible base URL", input: "https://provider.example/v1\r"},
		{needle: "Model", input: "gpt-4.1-mini\r"},
		{needle: "API key", input: "secret-api-key\r"},
	} {
		waitForCMAddPTYText(t, &output, step.needle)
		if _, err := process.Terminal().Write([]byte(step.input)); err != nil {
			t.Fatalf("write PTY input for %q: %v", step.needle, err)
		}
	}
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

func assertCMAddRichPTYOutput(t *testing.T, output string, color bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	for _, expected := range []string{
		"YCY / config cm add",
		"Add commit message profile",
		"Configure an OpenAI-compatible provider",
		"Collect CM profile details",
		"Save CM profile",
		"Profile name: work",
		"OpenAI-compatible base URL: https://provider.example/v1",
		"Model: gpt-4.1-mini",
		"API key: [redacted]",
		"Profile work added",
		"CM_ADD_WRITE_OK",
	} {
		if !strings.Contains(visible, expected) {
			t.Fatalf("Rich PTY output missing %q: %q", expected, output)
		}
	}
	if strings.Contains(output, "secret-api-key") {
		t.Fatalf("Rich PTY output leaked API key: %q", output)
	}
	enter := strings.LastIndex(visible, "\x1b[?1049h")
	leave := strings.LastIndex(visible, "\x1b[?1049l")
	if strings.Count(visible, "\x1b[?1049h") != 1 || strings.Count(visible, "\x1b[?1049l") != 1 || enter < 0 || leave < enter || !strings.Contains(visible, "\x1b[?25h") {
		t.Fatalf("Rich PTY output did not restore the primary screen: %q", output)
	}
	transcript := strings.Index(visible[leave:], "Collect CM profile details (completed)")
	result := strings.LastIndex(visible, "Profile work added")
	if transcript < 0 || result < 0 || leave+transcript > result {
		t.Fatalf("Rich PTY transcript/result ordering = %q", output)
	}
	if !color {
		for _, prefix := range []string{"\x1b[38;", "\x1b[3m", "\x1b[9m"} {
			if strings.Contains(output, prefix) {
				t.Fatalf("no-color Rich PTY output contains %q: %q", prefix, output)
			}
		}
	}
}

func cmAddPTYEnvironment() []string {
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

type lockedCMAddPTYBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (buffer *lockedCMAddPTYBuffer) Write(value []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.Write(value)
}

func (buffer *lockedCMAddPTYBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.String()
}

func waitForCMAddPTYText(t *testing.T, output *lockedCMAddPTYBuffer, needle string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if strings.Contains(output.String(), needle) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for PTY text %q: %q", needle, output.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
}
