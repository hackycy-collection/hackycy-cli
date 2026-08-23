package cm

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

func TestModuleReadsAndRendersTheSafeCMList(t *testing.T) {
	output := &bytes.Buffer{}
	module, err := New(Dependencies{
		Reader: fakeReader{profiles: appconfig.CMProfileList{
			DefaultProfile: "personal",
			Profiles: []appconfig.CMProfile{
				{Name: "work", Model: "gpt-4.1-mini", BaseURL: "https://work.example/v1"},
				{Name: "personal", Model: "deepseek-chat", BaseURL: "https://personal.example/v1"},
			},
		}},
		Output: output,
	})
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	if _, err := module.Run(context.Background(), Input{}); err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}
	for _, field := range []string{"work", "gpt-4.1-mini", "https://work.example/v1", "personal", "deepseek-chat", "https://personal.example/v1", "* personal"} {
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
