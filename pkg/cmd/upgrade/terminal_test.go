package upgrade

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
	"github.com/hackycy/hackycy-cli/internal/updater"
)

func TestUpgradeConsoleDescriptorProvidesSafeCurrentVersionContext(t *testing.T) {
	want := terminalexperience.ConsoleDescriptor{
		Command: "YCY / upgrade",
		Target:  "release update",
		Status:  "READY",
		Metadata: []terminalexperience.ConsoleMetadata{
			{Label: "scope", Value: "detached updater"},
			{Label: "current", Value: "1.2.3"},
		},
	}
	if got := terminalUpgradeConsoleDescriptor(" 1.2.3 "); !reflect.DeepEqual(got, want) {
		t.Fatalf("Console descriptor = %#v, want %#v", got, want)
	}
	unsafe := terminalUpgradeConsoleDescriptor("bad\x1b[31m\nversion")
	for _, field := range []string{unsafe.Command, unsafe.Target, unsafe.Status, unsafe.Metadata[0].Label, unsafe.Metadata[0].Value, unsafe.Metadata[1].Label, unsafe.Metadata[1].Value} {
		if strings.ContainsAny(field, "\r\n\t\x1b") {
			t.Fatalf("unsafe descriptor field contains terminal control: %q", field)
		}
	}
}

func TestRunUpgradeProjectsParentPhasesAndSubmitsOneResult(t *testing.T) {
	stdout, stderr := &countingUpgradeWriter{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Output:       stdout,
		Diagnostics:  stderr,
	})
	err := runUpgrade(&Options{
		Context:        context.Background(),
		Terminal:       experience,
		CurrentVersion: "1.0.0",
		run: func(_ context.Context, options updater.UpgradeOptions) (updater.UpgradeResult, error) {
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
	if got := stdout.String(); got != "Update to v2.0.0 has been scheduled and will finish after ycy exits.\n" || stdout.writes != 1 {
		t.Fatalf("stdout = (%d writes, %q)", stdout.writes, got)
	}
	for _, expected := range []string{"Consume startup transaction", "Resolve release", "Current v1.0.0; latest v2.0.0", "Resolve artifact", "checksum: release-digest", "Complete"} {
		if !strings.Contains(stderr.String(), expected) {
			t.Fatalf("stderr missing %q: %q", expected, stderr.String())
		}
	}
	if terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("plain streams contain terminal control: (%q, %q)", stdout.String(), stderr.String())
	}
}

func TestRunUpgradeWritesConsumedStateBeforeTheFinalResult(t *testing.T) {
	stdout, stderr := &countingUpgradeWriter{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.Automation},
		Output:       stdout,
		Diagnostics:  stderr,
	})
	previous := updater.UpdateTransaction{ExpectedVersion: "1.0.1", Status: updater.StatusSucceeded}
	err := runUpgrade(&Options{
		Context:        context.Background(),
		Terminal:       experience,
		CurrentVersion: "1.0.1",
		run: func(_ context.Context, options updater.UpgradeOptions) (updater.UpgradeResult, error) {
			options.Observer.PreviousState(previous)
			return updater.UpgradeResult{PreviousState: &previous, AlreadyCurrent: true, CurrentVersion: "1.0.1"}, nil
		},
	})
	if err != nil {
		t.Fatalf("runUpgrade() error = %v", err)
	}
	want := "Updated ycy to v1.0.1.\nCurrent version v1.0.1 is the latest.\nNo update needed.\n"
	if got := stdout.String(); got != want || stdout.writes != 1 {
		t.Fatalf("stdout = (%d writes, %q), want %q", stdout.writes, got, want)
	}
	if stderr.Len() != 0 || terminaltest.ContainsTerminalControl(stdout.Bytes()) {
		t.Fatalf("Automation streams = (%q, %q)", stdout.String(), stderr.String())
	}
}

func TestRunUpgradeRedactsClassifiedAbortAndFinishesOnce(t *testing.T) {
	stdout, stderr := &countingUpgradeWriter{}, &bytes.Buffer{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{Interaction: terminalexperience.PlainInteractive},
		Output:       stdout,
		Diagnostics:  stderr,
	})
	err := runUpgrade(&Options{
		Context:        context.Background(),
		Terminal:       experience,
		CurrentVersion: "1.0.0",
		run: func(_ context.Context, options updater.UpgradeOptions) (updater.UpgradeResult, error) {
			options.Observer.Phase(updater.UpgradePhaseEvent{Phase: updater.UpgradePhaseResolveRelease, State: updater.UpgradePhaseActive})
			options.Observer.Phase(updater.UpgradePhaseEvent{Phase: updater.UpgradePhaseResolveRelease, State: updater.UpgradePhaseFailed, Detail: "Release resolution failed"})
			options.Observer.Phase(updater.UpgradePhaseEvent{Phase: updater.UpgradePhaseComplete, State: updater.UpgradePhaseActive})
			options.Observer.Phase(updater.UpgradePhaseEvent{Phase: updater.UpgradePhaseComplete, State: updater.UpgradePhaseFailed, Detail: "Update failed"})
			return updater.UpgradeResult{Aborted: true}, &updater.ExitCodeError{Code: 0, Err: errors.New("download failed Bearer upgrade-secret")}
		},
	})
	if err == nil {
		t.Fatal("classified abort error was lost")
	}
	if got := stdout.String(); got != "Update aborted.\n" || stdout.writes != 1 {
		t.Fatalf("stdout = (%d writes, %q)", stdout.writes, got)
	}
	if !strings.Contains(stderr.String(), "error: download failed Bearer [REDACTED]") || strings.Contains(stderr.String(), "upgrade-secret") || terminaltest.ContainsTerminalControl(append(stdout.Bytes(), stderr.Bytes()...)) {
		t.Fatalf("diagnostics = %q", stderr.String())
	}
}

type countingUpgradeWriter struct {
	bytes.Buffer
	writes int
}

func (writer *countingUpgradeWriter) Write(value []byte) (int, error) {
	writer.writes++
	return writer.Buffer.Write(value)
}

func (writer *countingUpgradeWriter) WriteString(value string) (int, error) {
	writer.writes++
	return writer.Buffer.WriteString(value)
}
