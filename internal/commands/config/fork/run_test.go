package fork

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

func TestModuleReadsAndRendersTheSafeForkList(t *testing.T) {
	output := &bytes.Buffer{}
	module, err := New(Dependencies{
		Reader: fakeReader{instances: []appconfig.ForkInstance{{
			Name:         "work",
			Host:         "gitlab.example",
			Scheme:       "https",
			Type:         "gitlab",
			TokenPreview: "MDEy***",
		}}},
		Output: output,
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	if _, err := module.Run(context.Background(), Input{}); err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}
	for _, field := range []string{"work", "gitlab", "https", "gitlab.example", "MDEy***", "1 instance configured"} {
		if !strings.Contains(output.String(), field) {
			t.Fatalf("Run() output missing %q: %q", field, output.String())
		}
	}
}

func TestModuleReturnsReaderAndOutputFailures(t *testing.T) {
	readFailure := errors.New("read configuration")
	module, err := New(Dependencies{Reader: fakeReader{err: readFailure}, Output: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}
	if _, err := module.Run(context.Background(), Input{}); !errors.Is(err, readFailure) {
		t.Fatalf("Run() error = %v, want %v", err, readFailure)
	}

	writeFailure := errors.New("write output")
	module, err = New(Dependencies{Reader: fakeReader{}, Output: failingWriter{err: writeFailure}})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}
	if _, err := module.Run(context.Background(), Input{}); !errors.Is(err, writeFailure) {
		t.Fatalf("Run() error = %v, want %v", err, writeFailure)
	}
}

func TestNewRequiresReadAndOutputAdapters(t *testing.T) {
	if _, err := New(Dependencies{Output: &bytes.Buffer{}}); err == nil {
		t.Fatal("New() accepted a nil Reader")
	}
	if _, err := New(Dependencies{Reader: fakeReader{}}); err == nil {
		t.Fatal("New() accepted a nil Output")
	}
}

type failingWriter struct {
	err error
}

func (writer failingWriter) Write([]byte) (int, error) {
	return 0, writer.err
}
