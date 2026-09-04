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

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestRunForkAddRichPTYRestoresScreenAndRedactsTranscript(t *testing.T) {
	const helperEnvironment = "YCY_CONFIG_FORK_ADD_RICH_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		runForkAddRichPTYHelper(t)
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
			command := exec.Command(os.Args[0], "-test.run=^TestRunForkAddRichPTYRestoresScreenAndRedactsTranscript$")
			command.Env = forkAddPTYEnvironment()
			command.Env = append(command.Env, helperEnvironment+"=1", "TERM=xterm-256color")
			if !testCase.color {
				command.Env = append(command.Env, "NO_COLOR=1")
			}
			output := runForkAddPTYProcess(t, command, testCase.width, testCase.height)
			assertForkAddRichPTYOutput(t, output, testCase.color, testCase.width >= 70)
		})
	}
}

func runForkAddRichPTYHelper(t *testing.T) {
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
			return forkAddWriterFunc(func(_ string, _ appconfig.ForkInput) error {
				_, _ = fmt.Fprintln(os.Stderr, "FORK_ADD_WRITE_OK")
				return nil
			}), nil
		},
	})
	if err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}
}

func runForkAddPTYProcess(t *testing.T, command *exec.Cmd, width, height uint16) string {
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
	var output lockedForkAddPTYBuffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, process.Terminal())
		close(readDone)
	}()
	for _, step := range []struct {
		needle string
		input  string
	}{
		{needle: "Instance name (alias)", input: "work\r"},
		{needle: "Host", input: "gitlab.example\r"},
		{needle: "Provider type", input: "\r"},
		{needle: "Protocol", input: "\r"},
		{needle: "Access token", input: "secret-token\r"},
	} {
		waitForForkAddPTYText(t, &output, step.needle)
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

func assertForkAddRichPTYOutput(t *testing.T, output string, color, wide bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	expected := []string{
		"YCY / config fork add",
		"Add fork provider instance",
		"Store a provider connection for git fork operations",
		"Collect provider details",
		"Save provider instance",
		"Instance name (alias): work",
		"Host: gitlab.example",
		"Provider type: GitLab",
		"Protocol: HTTPS",
		"Access token: [redacted]",
		"Instance work (gitlab.example) added successfully",
		"FORK_ADD_WRITE_OK",
	}
	if !wide {
		expected = []string{"YCY / config fork add", "Collect provider details", "Save provider instance", "Access token: [redacted]", "FORK_ADD_WRITE_OK"}
	}
	for _, expected := range expected {
		if !strings.Contains(visible, expected) {
			t.Fatalf("Rich PTY output missing %q: %q", expected, output)
		}
	}
	if strings.Contains(output, "secret-token") {
		t.Fatalf("Rich PTY output leaked access token: %q", output)
	}
	enter := strings.LastIndex(visible, "\x1b[?1049h")
	leave := strings.LastIndex(visible, "\x1b[?1049l")
	if strings.Count(visible, "\x1b[?1049h") != 1 || strings.Count(visible, "\x1b[?1049l") != 1 || enter < 0 || leave < enter || !strings.Contains(visible, "\x1b[?25h") {
		t.Fatalf("Rich PTY output did not restore the primary screen: %q", output)
	}
	transcript := strings.Index(visible[leave:], "Collect provider details (completed)")
	resultNeedle := "Instance work (gitlab.example) added successfully"
	if !wide {
		resultNeedle = "Instance work (gitlab.example) added"
	}
	result := strings.LastIndex(visible, resultNeedle)
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

func forkAddPTYEnvironment() []string {
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

type lockedForkAddPTYBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (buffer *lockedForkAddPTYBuffer) Write(value []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.Write(value)
}

func (buffer *lockedForkAddPTYBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.String()
}

func waitForForkAddPTYText(t *testing.T, output *lockedForkAddPTYBuffer, needle string) {
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
