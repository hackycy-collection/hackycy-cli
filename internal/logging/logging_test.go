package logging

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParseLevelNormalizesInput(t *testing.T) {
	level, err := ParseLevel(" WARN ")
	if err != nil {
		t.Fatalf("ParseLevel returned an error: %v", err)
	}
	if level != Warn {
		t.Fatalf("level = %v, want %v", level, Warn)
	}
	if _, err := ParseLevel("verbose"); err == nil {
		t.Fatal("ParseLevel accepted an invalid level")
	}
}

func TestRuntimeFiltersAndRedacts(t *testing.T) {
	var output bytes.Buffer
	runtime := NewRuntime(Options{
		Writer: &output,
		Now:    func() time.Time { return time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC) },
	})
	runtime.SetLevel(Warn)
	logger := runtime.Logger("cli").With(map[string]any{"request": "one", "token": "secret"})
	logger.Info("Bearer top-secret", nil)
	logger.Warn("password=not-for-logs\ncontinuing", map[string]any{"authorization": "hidden", "count": 2})

	line := output.String()
	if strings.Contains(line, "top-secret") || strings.Contains(line, "not-for-logs") || strings.Contains(line, "hidden") || strings.Contains(line, `"token":"secret"`) {
		t.Fatalf("secret leaked in %q", line)
	}
	for _, expected := range []string{"2026-08-23T00:00:00Z WARN", "cli", `password=[REDACTED]\ncontinuing`, `"authorization":"[REDACTED]"`, `"count":2`} {
		if !strings.Contains(line, expected) {
			t.Fatalf("output %q does not contain %q", line, expected)
		}
	}
}

func TestRuntimeReconfiguresExistingLoggerAndColor(t *testing.T) {
	var output bytes.Buffer
	runtime := NewRuntime(Options{Writer: &output, Color: true})
	logger := runtime.Logger("scope")
	runtime.SetLevel(Error)
	logger.Warn("not written", nil)
	logger.Error("written", nil)

	if strings.Contains(output.String(), "not written") || !strings.Contains(output.String(), "\x1b[31mERROR") {
		t.Fatalf("unexpected output %q", output.String())
	}
}
