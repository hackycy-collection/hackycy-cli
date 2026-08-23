package exportenv

import (
	"errors"
	"testing"
)

func TestWriteOutputResolvesTargetFromWorkingDirectory(t *testing.T) {
	writer := &recordingWriter{}

	err := WriteOutput("/working-directory", "nested/output.json", "{\"VALUE\":\"value\"}", writer)

	if err != nil {
		t.Fatalf("WriteOutput returned an error: %v", err)
	}
	if writer.path != "/working-directory/nested/output.json" {
		t.Fatalf("writer path = %q", writer.path)
	}
	if writer.content != "{\"VALUE\":\"value\"}" {
		t.Fatalf("writer content = %q", writer.content)
	}
}

func TestWriteOutputKeepsAbsoluteTarget(t *testing.T) {
	writer := &recordingWriter{}

	err := WriteOutput("/working-directory", "/outside/output.json", "{}", writer)

	if err != nil {
		t.Fatalf("WriteOutput returned an error: %v", err)
	}
	if writer.path != "/outside/output.json" {
		t.Fatalf("writer path = %q", writer.path)
	}
}

func TestWriteOutputReturnsWriterFailure(t *testing.T) {
	want := errors.New("write failure")

	err := WriteOutput("/working-directory", "output.json", "{}", failingWriter{err: want})

	if !errors.Is(err, want) {
		t.Fatalf("WriteOutput error = %v, want %v", err, want)
	}
}

type recordingWriter struct {
	path    string
	content string
}

func (writer *recordingWriter) WriteFile(path string, content []byte) error {
	writer.path = path
	writer.content = string(content)
	return nil
}

type failingWriter struct {
	err error
}

func (writer failingWriter) WriteFile(string, []byte) error {
	return writer.err
}
