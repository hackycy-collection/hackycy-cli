package main

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"testing"

	fscommand "github.com/hackycy/hackycy-cli/internal/commands/fs"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalFSPresentationPreservesPlainAndAutomationLifecycle(t *testing.T) {
	startup := fscommand.Startup{
		URLs: []fscommand.StartupURL{
			{Label: "Local", URL: "http://localhost:43123"},
			{Label: "Network", URL: "http://192.168.1.50:43123"},
		},
		Directory:         "/workspace",
		BindingAddress:    "0.0.0.0",
		Port:              43123,
		ManagementEnabled: true,
		ChunkedUploads:    true,
		SafeHTML:          false,
		Authentication:    true,
		SessionDirectory:  "/workspace/.ycy/sessions",
	}
	want := "File Browser\n" +
		"Local: http://localhost:43123\n" +
		"Network: http://192.168.1.50:43123\n" +
		"Directory: /workspace\n" +
		"Bind: 0.0.0.0:43123\n" +
		"Management: true\n" +
		"Chunked uploads: true\n" +
		"HTML execution: true\n" +
		"Authentication: true\n" +
		"Session storage: /workspace/.ycy/sessions\n" +
		"File Browser stopped.\n"
	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.PlainInteractive},
		{Kind: terminalexperience.Automation},
	} {
		var output, diagnostics bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output, Diagnostics: &diagnostics})
		run := experience.Open(context.Background())
		if err := run.Present(terminalFSStartupDocument(session, startup)); err != nil {
			t.Fatalf("%v startup Present() error = %v", session.Kind, err)
		}
		if err := run.Present(terminalFSStoppedDocument(session)); err != nil {
			t.Fatalf("%v stopped Present() error = %v", session.Kind, err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("%v Close() error = %v", session.Kind, err)
		}
		if output.String() != want {
			t.Fatalf("%v presentation = %q, want %q", session.Kind, output.String(), want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("%v lifecycle contains terminal control: %q", session.Kind, output.String())
		}
		if diagnostics.Len() != 0 {
			t.Fatalf("%v lifecycle wrote stderr: %q", session.Kind, diagnostics.String())
		}
	}
}

func TestTerminalFSPresentationUsesRichSemanticRoles(t *testing.T) {
	startup := fscommand.Startup{
		URLs: []fscommand.StartupURL{
			{Label: "Local", URL: "http://localhost:43123"},
			{Label: "Network", URL: "http://192.168.1.50:43123"},
		},
		Directory:         "/workspace",
		BindingAddress:    "0.0.0.0",
		Port:              43123,
		ManagementEnabled: true,
		ChunkedUploads:    true,
		Authentication:    true,
		SessionDirectory:  "/workspace/.ycy/sessions",
	}
	want := []terminalexperience.VisualRole{
		terminalexperience.VisualRoleActive,
		terminalexperience.VisualRoleActive,
		terminalexperience.VisualRoleActive,
		terminalexperience.VisualRoleMuted,
		terminalexperience.VisualRoleMuted,
		terminalexperience.VisualRoleMuted,
		terminalexperience.VisualRoleMuted,
		terminalexperience.VisualRoleMuted,
		terminalexperience.VisualRoleMuted,
		terminalexperience.VisualRoleMuted,
	}
	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.RichInteractive, Color: true},
		{Kind: terminalexperience.RichInteractive},
	} {
		document := terminalFSStartupDocument(session, startup)
		if document.ClearBefore || len(document.Blocks) != len(want) {
			t.Fatalf("Rich document = %#v", document)
		}
		for index, role := range want {
			if document.Blocks[index].Role != role {
				t.Fatalf("Rich block %d role = %v, want %v", index, document.Blocks[index].Role, role)
			}
		}
		stopped := terminalFSStoppedDocument(session)
		if stopped.ClearBefore || len(stopped.Blocks) != 1 || stopped.Blocks[0].Role != terminalexperience.VisualRoleSuccess {
			t.Fatalf("Rich stopped document = %#v", stopped)
		}
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Present(document); err != nil {
			t.Fatalf("startup Present() error = %v", err)
		}
		if err := run.Present(stopped); err != nil {
			t.Fatalf("stopped Present() error = %v", err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if !session.Color && strings.Contains(output.String(), "\x1b[") {
			t.Fatalf("NO_COLOR Rich output contains style bytes: %q", output.String())
		}
	}
}

func TestFSHandlerClosesTheOperationWhenStartupPresentationFails(t *testing.T) {
	output := &failingFSWriter{}
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session: terminalexperience.Session{Kind: terminalexperience.Automation},
		Output:  output,
	})
	handler, err := newFSHandler(experience)
	if err != nil {
		t.Fatalf("newFSHandler() error = %v", err)
	}
	_, err = handler(context.Background(), fscommand.Input{Directory: t.TempDir(), Address: "127.0.0.1", Port: 0})
	if !errors.Is(err, errFSStartupPresentation) {
		t.Fatalf("handler error = %v, want startup presentation error", err)
	}
	if strings.Contains(output.String(), "File Browser stopped.") {
		t.Fatalf("failed startup wrote a stopped result: %q", output.String())
	}
	endpoint := ""
	for _, line := range strings.Split(output.String(), "\n") {
		if strings.HasPrefix(line, "Local: ") {
			endpoint = strings.TrimPrefix(line, "Local: ")
			break
		}
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" {
		t.Fatalf("startup endpoint = %q, parse error = %v", endpoint, err)
	}
	listener, err := net.Listen("tcp", parsed.Host)
	if err != nil {
		t.Fatalf("startup listener remains bound at %s: %v", parsed.Host, err)
	}
	_ = listener.Close()
}

var errFSStartupPresentation = errors.New("FS startup presentation failed")

type failingFSWriter struct {
	output bytes.Buffer
}

func (writer *failingFSWriter) Write(value []byte) (int, error) {
	_, _ = writer.output.Write(value)
	return 0, errFSStartupPresentation
}

func (writer *failingFSWriter) String() string {
	return writer.output.String()
}
