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
			command := exec.Command(os.Args[0], "-test.run=^TestRunCMAddRichPTYRestoresScreenAndRedactsTranscript$")
			command.Env = append(cmAddPTYEnvironment(), helperEnvironment+"=1", "TERM=xterm-256color")
			if !testCase.color {
				command.Env = append(command.Env, "NO_COLOR=1")
			}
			output := runCMAddPTYProcess(t, command, testCase.width, testCase.height)
			assertCMAddRichPTYOutput(t, output, testCase.color, testCase.width >= 70)
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

func runCMAddPTYProcess(t *testing.T, command *exec.Cmd, width, height uint16) string {
	t.Helper()
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY helper: %v", err)
	}
	defer process.Close()
	if err := process.Resize(width, height); err != nil {
		t.Fatalf("resize PTY to %dx%d: %v", width, height, err)
	}
	var output lockedCMAddPTYBuffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, process.Terminal())
		close(readDone)
	}()
	needles := []string{"Profile name", "OpenAI-compatible base URL", "Model", "API key"}
	if width < 70 {
		// Compact B activity rows wrap long Huh labels across several lines.
		needles = []string{"Profile name", "base", "Model", "API key"}
	}
	for index, step := range []struct {
		needle string
		input  string
	}{
		{needle: "Profile name", input: "work\r"},
		{needle: "OpenAI-compatible base URL", input: "https://provider.example/v1\r"},
		{needle: "Model", input: "gpt-4.1-mini\r"},
		{needle: "API key", input: "secret-api-key\r"},
	} {
		waitForCMAddPTYText(t, &output, needles[index])
		if _, err := process.Terminal().Write([]byte(step.input)); err != nil {
			t.Fatalf("write PTY input for %q: %v", needles[index], err)
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

func assertCMAddRichPTYOutput(t *testing.T, output string, color, wide bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	expected := []string{
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
	}
	if !wide {
		expected = []string{"YCY / config cm add", "Collect CM profile details", "Save CM profile", "API key: [redacted]", "CM_ADD_WRITE_OK"}
	}
	for _, expected := range expected {
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
