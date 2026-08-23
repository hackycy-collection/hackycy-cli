package exportenv

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestModuleExportsNamedEnvironmentToStdout(t *testing.T) {
	workingDirectory := t.TempDir()
	directory := filepath.Join(workingDirectory, "project")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatalf("make input directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".env.production"), []byte("VALUE=production\n"), 0o600); err != nil {
		t.Fatalf("write environment file: %v", err)
	}
	presenter := &recordingPresenter{}
	module, err := New(Dependencies{
		WorkingDirectory: func() (string, error) { return workingDirectory, nil },
		Selector:         &recordingSelector{},
		Reader:           osReader{},
		Writer:           osWriter{},
		Presenter:        presenter,
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	result, err := module.Run(context.Background(), Input{Directory: "project", Environment: "production"})

	if err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}
	if result.Cancelled {
		t.Fatal("Run reported cancellation")
	}
	if len(presenter.outros) != 1 || presenter.outros[0] != "Exported variables:" {
		t.Fatalf("outros = %#v", presenter.outros)
	}
	if len(presenter.printed) != 1 || presenter.printed[0] != "{\n  \"VALUE\": \"production\"\n}" {
		t.Fatalf("printed = %#v", presenter.printed)
	}
}

func TestModuleWritesOutputFromWorkingDirectoryAndOverwrites(t *testing.T) {
	workingDirectory := t.TempDir()
	directory := filepath.Join(workingDirectory, "input")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatalf("make input directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".env"), []byte("BASE=base\nSHARED=base\n"), 0o600); err != nil {
		t.Fatalf("write base file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".env.production"), []byte("PROD=production\nSHARED=production\n"), 0o600); err != nil {
		t.Fatalf("write environment file: %v", err)
	}
	outputPath := filepath.Join(workingDirectory, "output.json")
	if err := os.WriteFile(outputPath, []byte("obsolete"), 0o600); err != nil {
		t.Fatalf("write old output: %v", err)
	}
	presenter := &recordingPresenter{}
	module, err := New(Dependencies{
		WorkingDirectory: func() (string, error) { return workingDirectory, nil },
		Selector:         &recordingSelector{},
		Reader:           osReader{},
		Writer:           osWriter{},
		Presenter:        presenter,
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	_, err = module.Run(context.Background(), Input{
		Directory:   "input",
		Environment: "production",
		Merge:       true,
		Output:      "output.json",
	})

	if err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	want := "{\n  \"BASE\": \"base\",\n  \"PROD\": \"production\",\n  \"SHARED\": \"production\"\n}"
	if string(content) != want {
		t.Fatalf("output = %q, want %q", content, want)
	}
	if len(presenter.outros) != 1 || presenter.outros[0] != "Writing output to output.json" {
		t.Fatalf("outros = %#v", presenter.outros)
	}
	if len(presenter.printed) != 0 {
		t.Fatalf("printed = %#v", presenter.printed)
	}
}

func TestModuleReportsCancellationWithoutReadingOrWriting(t *testing.T) {
	workingDirectory := t.TempDir()
	if err := os.WriteFile(filepath.Join(workingDirectory, ".env.production"), []byte("VALUE=production\n"), 0o600); err != nil {
		t.Fatalf("write environment file: %v", err)
	}
	selector := &recordingSelector{cancel: true}
	reader := &countingFailureReader{}
	writer := &countingFailureWriter{}
	presenter := &recordingPresenter{}
	module, err := New(Dependencies{
		WorkingDirectory: func() (string, error) { return workingDirectory, nil },
		Selector:         selector,
		Reader:           reader,
		Writer:           writer,
		Presenter:        presenter,
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	result, err := module.Run(context.Background(), Input{})

	if err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}
	if !result.Cancelled {
		t.Fatal("Run did not report cancellation")
	}
	if reader.calls != 0 || writer.calls != 0 {
		t.Fatalf("reader calls = %d, writer calls = %d", reader.calls, writer.calls)
	}
	if len(presenter.alerts) != 1 || presenter.alerts[0] != "Cancelled" {
		t.Fatalf("alerts = %#v", presenter.alerts)
	}
}

func TestModuleReportsOutputTargetBeforeMissingParentFailure(t *testing.T) {
	workingDirectory := t.TempDir()
	if err := os.WriteFile(filepath.Join(workingDirectory, ".env"), []byte("VALUE=value\n"), 0o600); err != nil {
		t.Fatalf("write environment file: %v", err)
	}
	presenter := &recordingPresenter{}
	module, err := New(Dependencies{
		WorkingDirectory: func() (string, error) { return workingDirectory, nil },
		Selector:         &recordingSelector{},
		Reader:           osReader{},
		Writer:           osWriter{},
		Presenter:        presenter,
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	_, err = module.Run(context.Background(), Input{Output: "missing/output.json"})

	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if len(presenter.outros) != 1 || presenter.outros[0] != "Writing output to missing/output.json" {
		t.Fatalf("outros = %#v", presenter.outros)
	}
}

func TestModuleReturnsPermissionFailureFromWriter(t *testing.T) {
	workingDirectory := t.TempDir()
	if err := os.WriteFile(filepath.Join(workingDirectory, ".env"), []byte("VALUE=value\n"), 0o600); err != nil {
		t.Fatalf("write environment file: %v", err)
	}
	module, err := New(Dependencies{
		WorkingDirectory: func() (string, error) { return workingDirectory, nil },
		Selector:         &recordingSelector{},
		Reader:           osReader{},
		Writer:           permissionWriter{},
		Presenter:        &recordingPresenter{},
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	_, err = module.Run(context.Background(), Input{Output: "output.json"})

	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("Run error = %v, want permission error", err)
	}
}

type osReader struct{}

func (osReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

type osWriter struct{}

func (osWriter) WriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0o666)
}

type countingFailureReader struct {
	calls int
}

func (reader *countingFailureReader) ReadFile(string) ([]byte, error) {
	reader.calls++
	return nil, errors.New("reader must not run")
}

type countingFailureWriter struct {
	calls int
}

func (writer *countingFailureWriter) WriteFile(string, []byte) error {
	writer.calls++
	return errors.New("writer must not run")
}

type permissionWriter struct{}

func (permissionWriter) WriteFile(string, []byte) error {
	return fs.ErrPermission
}
