package rm

import (
	"reflect"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestRMConsoleDescriptorProvidesSafeRouteAndModeContext(t *testing.T) {
	force := true
	want := terminalexperience.ConsoleDescriptor{
		Command: "YCY / rm",
		Target:  "Remove selected files or clean project artifacts",
		Status:  "READY",
		Metadata: []terminalexperience.ConsoleMetadata{
			{Label: "scope", Value: "destructive filesystem mutation"},
			{Label: "route", Value: "explicit path removal"},
			{Label: "mode", Value: "force"},
		},
	}
	if got := terminalRMConsoleDescriptor(&Options{Paths: []string{"/private/absolute/path"}, Force: force}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Console descriptor = %#v, want %#v", got, want)
	}
	if got := terminalRMConsoleDescriptor(&Options{}); got.Metadata[1].Value != "smart cleanup" || got.Metadata[2].Value != "default-negative confirmation" {
		t.Fatalf("smart descriptor metadata = %#v", got.Metadata)
	}
	unsafe := terminalRMConsoleDescriptor(&Options{Paths: []string{"bad\npath"}})
	for _, field := range []string{unsafe.Command, unsafe.Target, unsafe.Status} {
		if terminaltest.ContainsTerminalControl([]byte(field)) {
			t.Fatalf("descriptor field contains terminal control: %q", field)
		}
	}
	for _, metadata := range unsafe.Metadata {
		if terminaltest.ContainsTerminalControl([]byte(metadata.Label)) || terminaltest.ContainsTerminalControl([]byte(metadata.Value)) {
			t.Fatalf("descriptor metadata contains terminal control: %#v", metadata)
		}
	}
}
