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
	PreviousState    *UpdateTransaction
	AlreadyCurrent   bool
	CurrentVersion   string
	Scheduled        bool
	ScheduledVersion string
	Aborted          bool
	State            UpdateTransaction
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
	result := UpgradeResult{}
	if state, consumeErr := ConsumeState(targetPath); consumeErr != nil {
		return UpgradeResult{}, consumeErr
	} else if state != nil {
		if state.Status == StatusPending {
			return UpgradeResult{}, errors.New("an update is already in progress")
		}
		result.PreviousState = state
	}
	if options.Resolver.CurrentVersion == "" {
		return result, errors.New("current CLI version is required")
	}
	resolution, err := ResolveRelease(ctx, options.Resolver)
	if err != nil {
		var already *AlreadyCurrentError
		if errors.As(err, &already) {
			result.AlreadyCurrent = true
			result.CurrentVersion = already.Current
			return result, nil
		}
		result.Aborted = true
		return result, classifyUpgradeError(err)
	}
	candidate, err := DownloadCandidate(ctx, resolution, targetPath, options.Candidate)
	if err != nil {
		result.Aborted = true
		return result, classifyUpgradeError(err)
	}
	updaterPath := updaterBinaryPath(options.TempDirectory(), candidate.TransactionID)
	if err := options.Copy(targetPath, updaterPath); err != nil {
		cleanupUpgradePaths(options.Remove, candidate.Path)
		return result, fmt.Errorf("copy updater: %w", err)
	}
	if err := protectUpgradePath(updaterPath, 0o755, os.Chmod); err != nil {
		cleanupUpgradePaths(options.Remove, candidate.Path, updaterPath)
		return result, fmt.Errorf("protect updater: %w", err)
	}
	state := NewUpdateTransaction(targetPath, candidate, resolution.Version, options.PID(), updaterPath, options.Now())
	if err := WriteState(state); err != nil {
		cleanupUpgradePaths(options.Remove, candidate.Path, updaterPath)
		return result, fmt.Errorf("publish pending update state: %w", err)
	}
	if err := options.Spawn(ctx, updaterPath, InternalUpdateArgs(state)); err != nil {
		cleanupUpgradePaths(options.Remove, candidate.Path, updaterPath, state.StatePath)
		return result, fmt.Errorf("schedule updater: %w", err)
	}
	result.Scheduled = true
	result.ScheduledVersion = resolution.Version
	result.State = state
	return result, nil
}

func normalizeUpgradeOptions(options UpgradeOptions) UpgradeOptions {
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

func classifyUpgradeError(err error) error {
	var status *HTTPStatusError
	if errors.As(err, &status) || strings.Contains(strings.ToLower(err.Error()), "checksum") || strings.Contains(strings.ToLower(err.Error()), "downloaded file is empty") {
		return &ExitCodeError{Code: 0, Err: err}
	}
	return &ExitCodeError{Code: 1, Err: err}
}

func cleanupUpgradePaths(remove func(string) error, paths ...string) {
	for _, path := range paths {
		_ = retryFileOperation(fileRetryCount, defaultFileSleep, func() error {
			return remove(path)
		})
	}
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
