package upgrade

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunUpgradeSchedulesOnlyVerifiedCandidate(t *testing.T) {
	content := []byte("new executable")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/latest":
			_, _ = io.WriteString(writer, `{"tag_name":"v2.0.0","assets":[{"name":"ycy-linux-x64","digest":"sha256:`+sha256Bytes(content)+`"}]}`)
		case "/download/v2.0.0/ycy-linux-x64":
			_, _ = writer.Write(content)
		default:
			t.Errorf("unexpected request path %s", request.URL.Path)
		}
	}))
	defer server.Close()
	directory := t.TempDir()
	target := filepath.Join(directory, "ycy")
	if err := os.WriteFile(target, []byte("old executable"), 0o755); err != nil {
		t.Fatal(err)
	}
	var spawnedPath string
	var spawnedArgs []string
	output := &strings.Builder{}
	result, err := RunUpgrade(context.Background(), UpgradeOptions{
		Resolver: ReleaseResolverOptions{LatestURL: server.URL + "/latest", DownloadBaseURL: server.URL + "/download", CurrentVersion: "1.0.0", GOOS: "linux", GOARCH: "amd64"},
		Candidate: CandidateOptions{
			TransactionID: func() (string, error) { return "tx", nil },
			Executor: func(context.Context, string, []string, []string) (ProcessResult, error) {
				return ProcessResult{Stdout: []byte("2.0.0\n")}, nil
			},
		},
		Output:     output,
		Executable: func() (string, error) { return target, nil },
		Copy: func(source, destination string) error {
			return os.WriteFile(destination, []byte("updater"), 0o700)
		},
		Spawn: func(_ context.Context, path string, args []string) error {
			spawnedPath, spawnedArgs = path, append([]string(nil), args...)
			return nil
		},
		TempDirectory: func() string { return directory },
		PID:           func() int { return 123 },
		Now:           func() time.Time { return time.Unix(1, 0) },
	})
	if err != nil || !result.Scheduled {
		t.Fatalf("run = %#v, %v", result, err)
	}
	if spawnedPath != filepath.Join(directory, "ycy-updater-tx") || len(spawnedArgs) == 0 || FindInternalMarker(spawnedArgs) != 0 {
		t.Fatalf("spawn = %s %#v", spawnedPath, spawnedArgs)
	}
	state, err := ReadState(StatePath(target))
	if err != nil || state == nil || state.Status != StatusPending {
		t.Fatalf("pending state = %#v, %v", state, err)
	}
	if !strings.Contains(output.String(), "scheduled") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRunUpgradeAlreadyCurrentAndFailureCleanup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = io.WriteString(writer, `{"tag_name":"v1.0.0"}`)
	}))
	defer server.Close()
	output := &strings.Builder{}
	result, err := RunUpgrade(context.Background(), UpgradeOptions{
		Resolver:   ReleaseResolverOptions{LatestURL: server.URL, CurrentVersion: "1.0.0", GOOS: "linux", GOARCH: "amd64"},
		Output:     output,
		Executable: func() (string, error) { return filepath.Join(t.TempDir(), "ycy"), nil },
	})
	if err != nil || !result.AlreadyCurrent || !strings.Contains(output.String(), "No update needed") {
		t.Fatalf("already current = %#v, %v, output %q", result, err, output.String())
	}

	content := []byte("new")
	failureServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/latest" {
			_, _ = io.WriteString(writer, `{"tag_name":"v2.0.0","assets":[{"name":"ycy-linux-x64","digest":"sha256:`+sha256Bytes(content)+`"}]}`)
			return
		}
		_, _ = writer.Write(content)
	}))
	defer failureServer.Close()
	directory := t.TempDir()
	target := filepath.Join(directory, "ycy")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	updater := filepath.Join(directory, "updater")
	_, err = RunUpgrade(context.Background(), UpgradeOptions{
		Resolver: ReleaseResolverOptions{LatestURL: failureServer.URL + "/latest", DownloadBaseURL: failureServer.URL, CurrentVersion: "1.0.0", GOOS: "linux", GOARCH: "amd64"},
		Candidate: CandidateOptions{TransactionID: func() (string, error) { return "fail", nil }, Executor: func(context.Context, string, []string, []string) (ProcessResult, error) {
			return ProcessResult{Stdout: []byte("2.0.0")}, nil
		}},
		Executable:    func() (string, error) { return target, nil },
		Copy:          func(source, destination string) error { return os.WriteFile(destination, []byte("updater"), 0o700) },
		Spawn:         func(context.Context, string, []string) error { return errors.New("spawn refused") },
		Remove:        func(path string) error { return os.Remove(path) },
		TempDirectory: func() string { return directory },
		Now:           time.Now,
		PID:           func() int { return 123 },
	})
	if err == nil || !strings.Contains(err.Error(), "spawn refused") {
		t.Fatalf("spawn failure = %v", err)
	}
	if fileExists(StatePath(target)) || fileExists(filepath.Join(directory, "ycy.new.fail")) || fileExists(updater) {
		t.Fatal("failed scheduling left transaction files")
	}
}

func TestConsumeStartupResultDoesNotReadAdjacentState(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "ycy")
	legacyPath := target + ".update-state.json"
	if err := os.WriteFile(legacyPath, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ConsumeStartupResult(target, io.Discard); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(legacyPath)
	if err != nil || string(contents) != "preserve" {
		t.Fatalf("adjacent state = %q, %v", contents, err)
	}
}
