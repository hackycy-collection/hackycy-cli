package upgrade

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

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
	"github.com/hackycy/hackycy-cli/internal/updater"
)

func TestRunUpgradeRichPTYRestoresPrimaryScreenAndReplaysSafeTranscript(t *testing.T) {
	const helperEnvironment = "YCY_UPGRADE_RICH_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		runUpgradeRichPTYHelper(t)
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
			command := exec.Command(os.Args[0], "-test.run=^TestRunUpgradeRichPTYRestoresPrimaryScreenAndReplaysSafeTranscript$")
			command.Env = append(upgradePTYEnvironment(), helperEnvironment+"=1", "TERM=xterm-256color")
			if testCase.extra != "" {
				command.Env = append(command.Env, testCase.extra)
			}
			output := runUpgradePTYProcess(t, command)
			assertUpgradeRichPTYOutput(t, output, testCase.color)
		})
	}
}

func runUpgradeRichPTYHelper(t *testing.T) {
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
	err := runUpgrade(&Options{
		Context:        context.Background(),
		Terminal:       experience,
		CurrentVersion: "1.0.0",
		run: func(_ context.Context, options updater.UpgradeOptions) (updater.UpgradeResult, error) {
			options.Observer.PreviousState(updater.UpdateTransaction{ExpectedVersion: "1.0.1", Status: updater.StatusSucceeded})
			for _, event := range []updater.UpgradePhaseEvent{
				{Phase: updater.UpgradePhaseConsumeStartupTransaction, State: updater.UpgradePhaseActive},
				{Phase: updater.UpgradePhaseConsumeStartupTransaction, State: updater.UpgradePhaseCompleted, Detail: "Startup transaction checked"},
				{Phase: updater.UpgradePhaseResolveRelease, State: updater.UpgradePhaseActive},
				{Phase: updater.UpgradePhaseResolveRelease, State: updater.UpgradePhaseCompleted, Detail: "Release metadata resolved", CurrentVersion: "1.0.0", CandidateVersion: "2.0.0", TargetOS: "linux", TargetArchitecture: "amd64"},
				{Phase: updater.UpgradePhaseResolveArtifact, State: updater.UpgradePhaseActive},
				{Phase: updater.UpgradePhaseResolveArtifact, State: updater.UpgradePhaseCompleted, Detail: "Artifact and checksum resolved", ArtifactName: "ycy-linux-x64", ChecksumSource: updater.ChecksumReleaseDigest},
				{Phase: updater.UpgradePhaseComplete, State: updater.UpgradePhaseActive},
				{Phase: updater.UpgradePhaseComplete, State: updater.UpgradePhaseCompleted, Detail: "Update scheduled", CandidateVersion: "2.0.0"},
			} {
				options.Observer.Phase(event)
			}
			return updater.UpgradeResult{Scheduled: true, ScheduledVersion: "2.0.0"}, nil
		},
	})
	if err != nil {
		t.Fatalf("runUpgrade() error = %v", err)
	}
}

func runUpgradePTYProcess(t *testing.T, command *exec.Cmd) string {
	t.Helper()
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY helper: %v", err)
	}
	defer process.Close()
	var output lockedUpgradePTYBuffer
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

func assertUpgradeRichPTYOutput(t *testing.T, output string, color bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	for _, expected := range []string{
		"Upgrade ycy",
		"Updated ycy to v1.0.1.",
		"Consume startup transaction",
		"Resolve release",
		"Resolve artifact",
		"checksum: release-digest",
		"Complete",
		"Update to v2.0.0 has been scheduled and will finish after ycy exits.",
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
	transcript := strings.Index(visible[leave:], "Resolve release (completed)")
	result := strings.LastIndex(visible, "Update to v2.0.0 has been scheduled and will finish after ycy exits.")
	if transcript < 0 || result < 0 || leave+transcript > result {
		t.Fatalf("Rich PTY transcript/result ordering = %q", output)
	}
	previous := strings.Index(visible[leave:], "Updated ycy to v1.0.1.")
	consume := strings.Index(visible[leave:], "Consume startup transaction (completed)")
	if previous < 0 || consume < 0 || previous > consume {
		t.Fatalf("Rich PTY prior-state Transcript ordering = %q", output)
	}
	if strings.Contains(output, "upgrade-secret") {
		t.Fatalf("Rich PTY output leaked a secret: %q", output)
	}
	if !color {
		for _, prefix := range []string{"\x1b[38;", "\x1b[3m", "\x1b[9m"} {
			if strings.Contains(output, prefix) {
				t.Fatalf("no-color Rich PTY output contains %q: %q", prefix, output)
			}
		}
	}
}

func upgradePTYEnvironment() []string {
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

type lockedUpgradePTYBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (buffer *lockedUpgradePTYBuffer) Write(value []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.Write(value)
}

func (buffer *lockedUpgradePTYBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.buf.String()
}
