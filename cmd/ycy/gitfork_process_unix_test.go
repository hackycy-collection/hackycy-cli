//go:build !windows

package main

import (
	"context"
	"errors"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestOSForkGitRunnerStopsTheChildProcessGroupWithoutAnOrphan(t *testing.T) {
	root := t.TempDir()
	grandchild := writeHeatGitScript(t, root, "grandchild", "#!/bin/sh\nwhile :; do\n  sleep 1\ndone\n")
	pidPath := filepath.Join(root, "grandchild-pid")
	parent := writeHeatGitScript(t, root, "parent", "#!/bin/sh\n\"$FORK_GIT_GRANDCHILD\" &\nchild=$!\nprintf '%s\\n' \"$child\" > \"$FORK_GIT_GRANDCHILD_PID\"\nwait \"$child\"\n")
	t.Setenv("FORK_GIT_GRANDCHILD", grandchild)
	t.Setenv("FORK_GIT_GRANDCHILD_PID", pidPath)

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	runner := &osForkGitRunner{executable: parent}
	results := make(chan error, 1)
	go func() {
		_, err := runner.Run(ctx, nil)
		results <- err
	}()

	pid := waitForHeatGitPID(t, pidPath)
	cancel(ycySignalCause{signal: syscall.SIGTERM})
	select {
	case err := <-results:
		var outcome *forkGitSignalOutcome
		if !errors.As(err, &outcome) || !errors.Is(err, context.Canceled) || outcome.ExitCode() != 143 {
			t.Fatalf("Run() error = %v, want SIGTERM outcome with code 143", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Git child process group did not stop")
	}
	if !waitForHeatGitGone(pid, 2*time.Second) {
		t.Fatalf("grandchild %d remained after Git cancellation", pid)
	}
}
