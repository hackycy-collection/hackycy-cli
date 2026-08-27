package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

type terminalDiscoveryPresenter struct {
	experience terminalexperience.Experience
}

func newTerminalDiscoveryPresenter(experience terminalexperience.Experience) cliapp.DiscoveryPresenter {
	return terminalDiscoveryPresenter{experience: experience}
}

func (presenter terminalDiscoveryPresenter) PresentDiscovery(ctx context.Context, document cliapp.DiscoveryDocument) {
	run := presenter.experience.Open(ctx)
	defer run.Close()
	_ = run.Present(terminalDiscoveryDocument(document))
}

func terminalDiscoveryDocument(document cliapp.DiscoveryDocument) terminalexperience.PresentationDocument {
	blocks := []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleTitle,
		Text: document.CommandPath,
	}}
	if document.Summary != "" {
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: terminalexperience.VisualRoleMuted,
			Text: document.Summary,
		})
	}
	blocks = append(blocks, terminalexperience.PresentationBlock{
		Role: terminalexperience.VisualRolePlain,
		Text: "Usage:\n  " + document.Usage,
	})
	if len(document.Descendants) > 0 {
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: terminalexperience.VisualRoleActive,
			Text: "Commands:\n" + formatDiscoveryDescendants(document.Descendants),
		})
	}
	if len(document.Flags) > 0 {
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: terminalexperience.VisualRolePlain,
			Text: "Flags:\n" + formatDiscoveryFlags(document.Flags),
		})
	}
	if len(document.Examples) > 0 {
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: terminalexperience.VisualRolePlain,
			Text: "Examples:\n" + strings.Join(document.Examples, "\n"),
		})
	}
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func formatDiscoveryDescendants(descendants []cliapp.DiscoveryDescendant) string {
	lines := make([]string, 0, len(descendants))
	for _, descendant := range descendants {
		line := "  " + descendant.Name
		if descendant.Summary != "" {
			line += "\t" + descendant.Summary
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func formatDiscoveryFlags(flags []cliapp.DiscoveryFlag) string {
	lines := make([]string, 0, len(flags))
	for _, flag := range flags {
		name := "--" + flag.Name
		if flag.Shorthand != "" {
			name = fmt.Sprintf("-%s, %s", flag.Shorthand, name)
		}
		line := "  " + name
		if flag.Usage != "" {
			line += "\t" + flag.Usage
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
