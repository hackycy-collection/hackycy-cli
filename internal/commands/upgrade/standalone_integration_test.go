package upgrade

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDetachedGoToGoStandaloneReplacementRollbackAndSelfCheck(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native Windows detached replacement is reserved for the artifact gate")
	}
	repository := repositoryRootForUpgradeTest(t)
	if _, err := os.Stat(filepath.Join(repository, "web", "dist")); errors.Is(err, os.ErrNotExist) {
		t.Skip("production Web output is required; run make build first")
	}
	directory := t.TempDir()
	first := filepath.Join(directory, "first")
	second := filepath.Join(directory, "second")
	buildStandaloneUpgradeArtifact(t, repository, first, "1.0.0")
	buildStandaloneUpgradeArtifact(t, repository, second, "2.0.0")

	target := filepath.Join(directory, "ycy")
	staged := target + ".new.success"
	backup := target + ".backup.success"
	updater := filepath.Join(directory, "ycy-updater-success")
	copyUpgradeFile(t, first, target)
	copyUpgradeFile(t, second, staged)
	copyUpgradeFile(t, first, updater)
	parentPID := exitedUpgradeParent(t)
	success := UpdateTransaction{
		TransactionID: "success-success-success-success-successsuccess",
		ParentPID:     parentPID, TargetPath: target, StagedPath: staged, BackupPath: backup,
		ExpectedHash: mustUpgradeFileHash(t, staged), ExpectedVersion: "2.0.0",
		StatePath: StatePath(target), UpdaterPath: updater,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Status: StatusPending,
	}
	if err := WriteState(success); err != nil {
		t.Fatal(err)
	}
	runUpgradeUpdater(t, updater, InternalUpdateArgs(success))
	state, err := ReadState(success.StatePath)
	if err != nil || state == nil || state.Status != StatusSucceeded {
		t.Fatalf("success state = %#v, %v", state, err)
	}
	assertUpgradeVersion(t, target, "2.0.0")
	if fileExists(backup) || fileExists(staged) {
		t.Fatal("successful replacement left staged or backup files")
	}

	rollbackTarget := filepath.Join(directory, "ycy-rollback")
	rollbackStaged := rollbackTarget + ".new.failure"
	rollbackBackup := rollbackTarget + ".backup.failure"
	rollbackUpdater := filepath.Join(directory, "ycy-updater-failure")
	copyUpgradeFile(t, second, rollbackTarget)
	copyUpgradeFile(t, first, rollbackStaged)
	copyUpgradeFile(t, second, rollbackUpdater)
	failure := UpdateTransaction{
		TransactionID: "failure-failure-failure-failure-failurefailure",
		ParentPID:     exitedUpgradeParent(t), TargetPath: rollbackTarget, StagedPath: rollbackStaged, BackupPath: rollbackBackup,
		ExpectedHash: mustUpgradeFileHash(t, rollbackStaged), ExpectedVersion: "2.0.0",
		StatePath: StatePath(rollbackTarget), UpdaterPath: rollbackUpdater,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Status: StatusPending,
	}
	if err := WriteState(failure); err != nil {
		t.Fatal(err)
	}
	failedCommand := exec.Command(rollbackUpdater, InternalUpdateArgs(failure)...)
	if output, runErr := failedCommand.CombinedOutput(); runErr == nil {
		t.Fatal("failed updater unexpectedly succeeded")
	} else if len(output) > 0 {
		t.Logf("failed updater output: %s", output)
	}
	state, err = ReadState(failure.StatePath)
	if err != nil || state == nil || state.Status != StatusFailed {
		t.Fatalf("failure state = %#v, %v", state, err)
	}
	assertUpgradeVersion(t, rollbackTarget, "2.0.0")

	// Completed state must not contaminate --version, then is consumed by the next normal command.
	versionCommand := exec.Command(target, "--version")
	versionOutput, err := versionCommand.CombinedOutput()
	if err != nil || string(versionOutput) != "2.0.0\n" {
		t.Fatalf("plain self-check = %q, %v", versionOutput, err)
	}
	if !fileExists(success.StatePath) {
		t.Fatal("--version consumed completed state")
	}
	helpCommand := exec.Command(target, "--help")
	helpOutput, err := helpCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("help after update = %v\n%s", err, helpOutput)
	}
	if !strings.Contains(string(helpOutput), "Updated ycy to v2.0.0.") || fileExists(success.StatePath) {
		t.Fatalf("result consumption = %q, state remains=%v", helpOutput, fileExists(success.StatePath))
	}
}

func buildStandaloneUpgradeArtifact(t *testing.T, repository, output, version string) {
	t.Helper()
	command := exec.Command("go", "build", "-trimpath", "-ldflags", "-X main.version="+version, "-o", output, "./cmd/ycy")
	command.Dir = repository
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.26.7", "GOWORK=off", "CGO_ENABLED=0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build %s artifact: %v\n%s", version, err, output)
	}
}

func copyUpgradeFile(t *testing.T, source, destination string) {
	t.Helper()
	input, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		t.Fatal(err)
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(destination, 0o755); err != nil {
		t.Fatal(err)
	}
}

func exitedUpgradeParent(t *testing.T) int {
	t.Helper()
	command := exec.Command("sh", "-c", "exit 0")
	if err := command.Run(); err != nil {
		t.Fatal(err)
	}
	return command.Process.Pid
}

func runUpgradeUpdater(t *testing.T, updater string, arguments []string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, updater, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run hidden updater: %v\n%s", err, output)
	}
	return string(output)
}

func assertUpgradeVersion(t *testing.T, path, expected string) {
	t.Helper()
	command := exec.Command(path, "--version")
	output, err := command.CombinedOutput()
	if err != nil || string(output) != expected+"\n" {
		t.Fatalf("%s version = %q, %v", path, output, err)
	}
}

func mustUpgradeFileHash(t *testing.T, path string) string {
	t.Helper()
	hash, err := hashFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}
