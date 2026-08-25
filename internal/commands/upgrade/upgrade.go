package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ExitCodeError lets the CLI preserve the historical status of an abort without process exit in the module.
type ExitCodeError struct {
	Code int
	Err  error
}

func (err *ExitCodeError) Error() string {
	if err.Err == nil {
		return "upgrade aborted"
	}
	return err.Err.Error()
}
func (err *ExitCodeError) Unwrap() error { return err.Err }
func (err *ExitCodeError) ExitCode() int { return err.Code }

// SpawnDetached starts the copied updater without retaining a child process handle.
type SpawnDetached func(context.Context, string, []string) error

// UpgradeOptions keeps release, process, and filesystem facts injectable.
type UpgradeOptions struct {
	Resolver      ReleaseResolverOptions
	Candidate     CandidateOptions
	Replacement   ReplacementOptions
	Output        io.Writer
	ErrorOutput   io.Writer
	Executable    func() (string, error)
	Spawn         SpawnDetached
	Copy          func(string, string) error
	Remove        func(string) error
	Now           func() time.Time
	PID           func() int
	TempDirectory func() string
}

// UpgradeResult describes a successful no-op or scheduled transaction.
type UpgradeResult struct {
	AlreadyCurrent bool
	Scheduled      bool
	State          UpdateTransaction
}

// RunUpgrade performs one Go-to-Go scheduling transaction.
func RunUpgrade(ctx context.Context, options UpgradeOptions) (UpgradeResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	options = normalizeUpgradeOptions(options)
	targetPath, err := options.Executable()
	if err != nil {
		return UpgradeResult{}, fmt.Errorf("resolve current executable: %w", err)
	}
	targetPath, err = filepath.Abs(targetPath)
	if err != nil {
		return UpgradeResult{}, fmt.Errorf("resolve current executable: %w", err)
	}
	if state, consumeErr := ConsumeState(targetPath); consumeErr != nil {
		return UpgradeResult{}, consumeErr
	} else if state != nil {
		if state.Status == StatusPending {
			return UpgradeResult{}, errors.New("an update is already in progress")
		}
		writeUpgradeResult(options.Output, FormatStateResult(*state))
	}
	if options.Resolver.CurrentVersion == "" {
		return UpgradeResult{}, errors.New("current CLI version is required")
	}
	resolution, err := ResolveRelease(ctx, options.Resolver)
	if err != nil {
		var already *AlreadyCurrentError
		if errors.As(err, &already) {
			writeUpgradeResult(options.Output, fmt.Sprintf("Current version v%s is the latest.\nNo update needed.", already.Current))
			return UpgradeResult{AlreadyCurrent: true}, nil
		}
		return UpgradeResult{}, classifyUpgradeError(options.Output, options.ErrorOutput, err)
	}
	candidate, err := DownloadCandidate(ctx, resolution, targetPath, options.Candidate)
	if err != nil {
		return UpgradeResult{}, classifyUpgradeError(options.Output, options.ErrorOutput, err)
	}
	updaterPath := filepath.Join(options.TempDirectory(), "ycy-updater-"+candidate.TransactionID+filepath.Ext(targetPath))
	if err := options.Copy(targetPath, updaterPath); err != nil {
		_ = options.Remove(candidate.Path)
		return UpgradeResult{}, fmt.Errorf("copy updater: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(updaterPath, 0o755); err != nil {
			_ = options.Remove(candidate.Path)
			_ = options.Remove(updaterPath)
			return UpgradeResult{}, fmt.Errorf("make updater executable: %w", err)
		}
	}
	state := NewUpdateTransaction(targetPath, candidate, resolution.Version, options.PID(), updaterPath, options.Now())
	if err := WriteState(state); err != nil {
		_ = options.Remove(candidate.Path)
		_ = options.Remove(updaterPath)
		return UpgradeResult{}, fmt.Errorf("publish pending update state: %w", err)
	}
	if err := options.Spawn(ctx, updaterPath, InternalUpdateArgs(state)); err != nil {
		_ = options.Remove(candidate.Path)
		_ = options.Remove(updaterPath)
		_ = options.Remove(state.StatePath)
		return UpgradeResult{}, fmt.Errorf("schedule updater: %w", err)
	}
	writeUpgradeResult(options.Output, fmt.Sprintf("Update to v%s has been scheduled and will finish after ycy exits.", resolution.Version))
	return UpgradeResult{Scheduled: true, State: state}, nil
}

// ConsumeStartupResult presents one completed Go result while leaving version self-check output clean.
func ConsumeStartupResult(targetPath string, output io.Writer) error {
	state, err := ConsumeState(targetPath)
	if err != nil {
		return err
	}
	if state != nil && state.Status != StatusPending {
		writeUpgradeResult(output, FormatStateResult(*state))
	}
	return nil
}

func normalizeUpgradeOptions(options UpgradeOptions) UpgradeOptions {
	if options.Output == nil {
		options.Output = io.Discard
	}
	if options.ErrorOutput == nil {
		options.ErrorOutput = io.Discard
	}
	if options.Executable == nil {
		options.Executable = os.Executable
	}
	if options.Spawn == nil {
		options.Spawn = spawnDetached
	}
	if options.Copy == nil {
		options.Copy = copyFile
	}
	if options.Remove == nil {
		options.Remove = os.Remove
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.PID == nil {
		options.PID = os.Getpid
	}
	if options.TempDirectory == nil {
		options.TempDirectory = os.TempDir
	}
	if options.Resolver.GOOS == "" {
		options.Resolver.GOOS = runtime.GOOS
	}
	if options.Resolver.GOARCH == "" {
		options.Resolver.GOARCH = runtime.GOARCH
	}
	if options.Candidate.Client == nil {
		options.Candidate.Client = options.Resolver.Client
	}
	return options
}

func classifyUpgradeError(output, errorOutput io.Writer, err error) error {
	writeUpgradeResult(errorOutput, "error: "+err.Error())
	writeUpgradeResult(output, "Update aborted.")
	var status *HTTPStatusError
	if errors.As(err, &status) || strings.Contains(strings.ToLower(err.Error()), "checksum") || strings.Contains(strings.ToLower(err.Error()), "downloaded file is empty") {
		return &ExitCodeError{Code: 0, Err: err}
	}
	return &ExitCodeError{Code: 1, Err: err}
}

func writeUpgradeResult(output io.Writer, message string) {
	if output == nil || message == "" {
		return
	}
	_, _ = fmt.Fprintln(output, message)
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o700)
	if err != nil {
		return err
	}
	if _, copyErr := io.Copy(output, input); copyErr != nil {
		_ = output.Close()
		_ = os.Remove(destination)
		return copyErr
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(destination)
		return err
	}
	return nil
}

func spawnDetached(ctx context.Context, path string, arguments []string) error {
	command := exec.CommandContext(ctx, path, arguments...)
	command.Stdin = nil
	command.Stdout = nil
	command.Stderr = nil
	configureDetachedCommand(command)
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
