package pulse

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/gitprocess"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestRunPulseRichPTYRestoresPrimaryScreenAndReplaysFourPhaseTranscript(t *testing.T) {
	const helperEnvironment = "YCY_GIT_PULSE_RICH_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		runPulseRichPTYHelper(t)
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
			command := exec.Command(os.Args[0], "-test.run=^TestRunPulseRichPTYRestoresPrimaryScreenAndReplaysFourPhaseTranscript$")
			command.Env = append(pulsePTYEnvironment(), helperEnvironment+"=1", "TERM=xterm-256color")
			if testCase.extra != "" {
				command.Env = append(command.Env, testCase.extra)
			}
			output := runPulsePTYProcess(t, command)
			assertPulseRichPTYOutput(t, output, testCase.color)
		})
	}
}

func runPulseRichPTYHelper(t *testing.T) {
	t.Helper()
	workspace := t.TempDir()
	initializeStandalonePulseRepository(t, workspace, "Ada", "ada@example.test", "safe commit")
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
	err := runPulse(&Options{
		Context:          context.Background(),
		Directory:        workspace,
		Days:             pulseInt(1),
		WorkingDirectory: func() (string, error) { return workspace, nil },
		Terminal:         experience,
		Git:              &gitprocess.Runner{},
		Now:              time.Now,
	})
	if err != nil {
		t.Fatalf("runPulse() error = %v", err)
	}
}

func runPulsePTYProcess(t *testing.T, command *exec.Cmd) string {
	t.Helper()
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY helper: %v", err)
	}
	defer process.Close()
	var output lockedPulsePTYBuffer
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

func assertPulseRichPTYOutput(t *testing.T, output string, color bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	for _, expected := range []string{
		"YCY / git pulse",
		"Prepare workspace",
		"Scan repositories",
		"Fetch commits",
		"Build commit tree",
		"Workspace commit activity",
		"safe commit",
	} {
		if !strings.Contains(visible, expected) {
			t.Fatalf("Rich PTY output missing %q: %q", expected, output)
		}
	}
	enter := strings.LastIndex(visible, "\x1b[?1049h")
	leave := strings.LastIndex(visible, "\x1b[?1049l")
	if strings.Count(visible, "\x1b[?1049h") != 1 || strings.Count(visible, "\x1b[?1049l") != 1 || enter < 0 || leave < enter || !strings.Contains(visible, "\x1b[?25h") {
		t.Fatalf("Rich PTY output did not restore the primary screen: %q", output)
	}
	transcript := strings.Index(visible[leave:], "Prepare workspace (completed)")
	result := strings.LastIndex(visible, "safe commit")
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

func pulsePTYEnvironment() []string {
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

type lockedPulsePTYBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (buffer *lockedPulsePTYBuffer) Write(value []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.Write(value)
}

func (buffer *lockedPulsePTYBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.String()
}
