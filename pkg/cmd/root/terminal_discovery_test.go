package root

import (
	"context"
	"reflect"
	"testing"

	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalDiscoveryPresenterTranslatesAndClosesOneExperienceRun(t *testing.T) {
	experience := terminaltest.NewRecordingExperience()
	presenter := newTerminalDiscoveryAdapter(experience)
	err := presenter.PresentDiscovery(context.Background(), DiscoveryDocument{
		CommandPath: "ycy config",
		Summary:     "Manage ycy configuration",
		Usage:       "ycy config [flags]",
		Descendants: []DiscoveryDescendant{{Name: "cm", Summary: "Manage CM profiles"}},
		Flags:       []DiscoveryFlag{{Name: "log-level", Usage: "Log level"}},
		Examples:    []string{"ycy config --help"},
	})
	if err != nil {
		t.Fatalf("PresentDiscovery() error = %v", err)
	}

	operations := experience.Run.Operations()
	if len(operations) != 2 || operations[0].Kind != terminaltest.FinishOperation || operations[1].Kind != terminaltest.CloseOperation {
		t.Fatalf("operations = %#v", operations)
	}
	finish, ok := operations[0].Value.(terminaltest.Finish)
	if !ok {
		t.Fatalf("presented value = %#v", operations[0].Value)
	}
	if finish.Outcome != terminalexperience.Succeeded || finish.Document == nil {
		t.Fatalf("finish = %#v", finish)
	}
	document := *finish.Document
	want := terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{
		{Role: terminalexperience.VisualRoleTitle, Text: "ycy config"},
		{Role: terminalexperience.VisualRoleMuted, Text: "Manage ycy configuration"},
		{Role: terminalexperience.VisualRolePlain, Text: "Usage:\n  ycy config [flags]"},
		{Role: terminalexperience.VisualRoleActive, Text: "Commands:\n  cm\tManage CM profiles"},
		{Role: terminalexperience.VisualRolePlain, Text: "Flags:\n  --log-level\tLog level"},
		{Role: terminalexperience.VisualRolePlain, Text: "Examples:\nycy config --help"},
	}}
	if !reflect.DeepEqual(document, want) {
		t.Fatalf("document = %#v, want %#v", document, want)
	}
}
