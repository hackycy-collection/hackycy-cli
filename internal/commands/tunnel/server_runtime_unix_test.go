//go:build darwin || linux

package tunnel

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServerRuntimeActivatesPreparedFRPSAndPreparesRestart(t *testing.T) {
	artifact, err := CurrentFRPArtifact()
	if err != nil {
		t.Fatalf("CurrentFRPArtifact() error = %v", err)
	}
	preparations := make(chan struct{}, 2)
	runtimeDirectory := filepath.Join(t.TempDir(), "frp", FRPVersion)
	options := ServerRuntimeOptions{
		Settings: ServerHTTPServerSettings{
			Address:     "127.0.0.1",
			ControlPort: 0,
			FRPPort:     17000,
			HTTPPort:    18080,
			PortRange:   ServerHTTPPortRange{Start: 20000, End: 20100},
			DataDir:     t.TempDir(),
			AdminUser:   "admin",
		},
		AdminPassword:       "environment-password",
		frpArtifact:         &artifact,
		frpRuntimeDirectory: runtimeDirectory,
		ensureFRPRuntime: func(_ context.Context, directory string, received FRPArtifact) (FRPRuntimePaths, error) {
			paths := frpRuntimePaths(directory, received.Target)
			if err := os.MkdirAll(paths.Directory, 0o755); err != nil {
				return FRPRuntimePaths{}, err
			}
			if err := os.WriteFile(paths.FRPC, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
				return FRPRuntimePaths{}, err
			}
			if err := os.WriteFile(paths.FRPS, []byte("#!/bin/sh\nif [ \"$1\" = verify ]; then exit 0; fi\nif [ \"$1\" = -c ]; then trap 'exit 0' TERM INT; while :; do sleep 1; done; fi\nexit 9\n"), 0o755); err != nil {
				return FRPRuntimePaths{}, err
			}
			preparations <- struct{}{}
			return paths, nil
		},
	}
	runtime, err := NewServerRuntime(context.Background(), options)
	if err != nil {
		t.Fatalf("NewServerRuntime() error = %v", err)
	}
	server, err := runtime.Start()
	if err != nil {
		_ = runtime.Close()
		t.Fatalf("Runtime.Start() error = %v", err)
	}

	waitForFRPSupervisor(t, 5*time.Second, func() bool {
		return runtime.frps.FRPSState().State == FRPProcessRunning
	})
	<-preparations
	if err := runtime.frps.Restart(context.Background()); err != nil {
		_ = server.Close()
		t.Fatalf("ManagedFRPS.Restart() error = %v", err)
	}
	<-preparations
	if state := runtime.frps.FRPSState(); state.State != FRPProcessRunning || state.PID == nil {
		_ = server.Close()
		t.Fatalf("FRPS state after restart = %#v", state)
	}
	if err := server.Close(); err != nil {
		t.Fatalf("RunningServer.Close() error = %v", err)
	}
	if err := server.Wait(); err != nil {
		t.Fatalf("RunningServer.Wait() error = %v", err)
	}
}
