package connect

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

func terminalTunnelConnectionSelectorFor(experience *terminalexperience.Runtime) ClientConnectionSelector {
	if experience == nil || experience.Session().Kind == terminalexperience.Automation {
		return nil
	}
	return func(ctx context.Context, connections []appconfig.TunnelConnection) (string, bool, error) {
		run := experience.Open(ctx)
		defer run.Close()
		return newTerminalTunnelConnectionAdapter(run, experience.Session()).Select(ctx, connections)
	}
}

type terminalTunnelConnectionAdapter struct {
	run     terminalexperience.ExperienceRun
	session terminalexperience.Session
}

func newTerminalTunnelConnectionAdapter(run terminalexperience.ExperienceRun, session terminalexperience.Session) *terminalTunnelConnectionAdapter {
	return &terminalTunnelConnectionAdapter{run: run, session: session}
}

func (adapter *terminalTunnelConnectionAdapter) Select(_ context.Context, connections []appconfig.TunnelConnection) (string, bool, error) {
	answer, err := adapter.run.Ask(terminalexperience.InteractionRequest{
		Kind:         terminalexperience.InteractionSelect,
		Message:      "Select a tunnel connection",
		PlainLead:    "Select a tunnel connection",
		PlainPrompt:  "> ",
		Options:      tunnelConnectionInteractionOptions(connections),
		CancelValues: []string{"", "q", "quit", "cancel"},
		ParsePlain: func(value string) (terminalexperience.InteractionAnswer, error) {
			return parseTunnelConnectionSelection(value, connections)
		},
	})
	if errors.Is(err, terminalexperience.ErrInteractionCancelled) {
		_ = adapter.run.Present(terminalTunnelConnectionDocument(adapter.session, "Cancelled"))
		return "", true, nil
	}
	if err != nil {
		return "", false, err
	}
	return answer.Value, false, nil
}

func tunnelConnectionInteractionOptions(connections []appconfig.TunnelConnection) []terminalexperience.InteractionOption {
	options := make([]terminalexperience.InteractionOption, 0, len(connections))
	for _, connection := range connections {
		options = append(options, terminalexperience.InteractionOption{
			Label: fmt.Sprintf("%s  %s", connection.Server, maskTunnelToken(connection.Token)),
			Value: connection.ID,
		})
	}
	return options
}

func parseTunnelConnectionSelection(value string, connections []appconfig.TunnelConnection) (terminalexperience.InteractionAnswer, error) {
	index, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || index < 1 || index > len(connections) {
		return terminalexperience.InteractionAnswer{}, errors.New("Invalid selection")
	}
	return terminalexperience.InteractionAnswer{Value: connections[index-1].ID}, nil
}

func terminalTunnelConnectionDocument(session terminalexperience.Session, text string) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleWarning
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: text}}}
}

func maskTunnelToken(token string) string {
	if len(token) <= 12 {
		return strings.Repeat("*", len(token))
	}
	return token[:8] + "********" + token[len(token)-4:]
}
