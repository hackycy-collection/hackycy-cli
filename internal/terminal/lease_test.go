package terminal_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/terminal"
)

func TestLeaseAwareDiagnosticWriterDefersAndFlushesInOrder(t *testing.T) {
	var diagnostics bytes.Buffer
	writer := terminal.NewLeaseAwareDiagnosticWriter(&diagnostics)
	lease := writer.AcquireRendererLease()

	if _, err := io.WriteString(lease.Writer(), "view\n"); err != nil {
		t.Fatalf("renderer write error = %v", err)
	}
	if _, err := io.WriteString(writer, "first diagnostic\n"); err != nil {
		t.Fatalf("first diagnostic write error = %v", err)
	}
	if _, err := io.WriteString(writer, "second diagnostic\n"); err != nil {
		t.Fatalf("second diagnostic write error = %v", err)
	}
	if got, want := diagnostics.String(), "view\n"; got != want {
		t.Fatalf("diagnostics during lease = %q, want %q", got, want)
	}

	if err := lease.Close(); err != nil {
		t.Fatalf("lease.Close() error = %v", err)
	}
	if got, want := diagnostics.String(), "view\nfirst diagnostic\nsecond diagnostic\n"; got != want {
		t.Fatalf("diagnostics after lease = %q, want %q", got, want)
	}
	if _, err := io.WriteString(lease.Writer(), "late renderer output"); err == nil {
		t.Fatal("closed lease accepted renderer output")
	}
}

func TestRendererLeaseSerializesOwnership(t *testing.T) {
	writer := terminal.NewLeaseAwareDiagnosticWriter(io.Discard)
	first := writer.AcquireRendererLease()
	secondReady := make(chan struct{})
	secondClosed := make(chan struct{})
	go func() {
		second := writer.AcquireRendererLease()
		close(secondReady)
		_ = second.Close()
		close(secondClosed)
	}()

	select {
	case <-secondReady:
		t.Fatal("second lease acquired before first lease released")
	case <-time.After(50 * time.Millisecond):
	}
	if err := first.Close(); err != nil {
		t.Fatalf("first.Close() error = %v", err)
	}
	select {
	case <-secondClosed:
	case <-time.After(time.Second):
		t.Fatal("second lease did not acquire after first lease released")
	}
}
