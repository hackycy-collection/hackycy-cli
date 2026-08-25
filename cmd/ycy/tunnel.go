package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	"github.com/hackycy/hackycy-cli/internal/commands/tunnel"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func newTunnelServerHandler(logger logging.Logger) cliapp.TunnelServerHandler {
	return func(ctx context.Context, config tunnel.ServerConfig) error {
		return tunnel.RunServer(ctx, config, tunnel.ServerRunOptions{Logger: logger})
	}
}

func newTunnelConnectHandler(input io.Reader, output io.Writer, logger logging.Logger, ycyVersion string) cliapp.TunnelConnectHandler {
	return func(ctx context.Context, optionInput tunnel.ClientOptionInput) error {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return err
		}
		resolved, err := tunnel.ResolveClientConfig(ctx, optionInput, tunnel.ClientResolutionOptions{
			Reader:           store,
			Environment:      os.LookupEnv,
			DefaultServer:    tunnel.DefaultTunnelServer,
			SelectConnection: terminalTunnelConnectionSelectorFor(input, output),
		})
		if err != nil {
			logger.Error("Could not resolve tunnel client configuration", map[string]any{"error": err.Error()})
			return err
		}
		if resolved == nil {
			return nil
		}
		runOptions := tunnel.ClientRunOptions{
			InstanceIdentity: store,
			Logger:           logger,
			YCYVersion:       ycyVersion,
		}
		if resolved.RememberOnAuthentication {
			config := resolved.Config
			runOptions.OnAuthenticated = func() error {
				return store.RememberTunnelConnection(config.Server, config.Token, time.Now())
			}
		}
		return tunnel.RunClient(ctx, resolved.Config, runOptions)
	}
}

func terminalTunnelConnectionSelectorFor(input io.Reader, output io.Writer) tunnel.ClientConnectionSelector {
	inputFile, inputIsFile := input.(*os.File)
	outputFile, outputIsFile := output.(*os.File)
	if !inputIsFile || !outputIsFile || !terminal(inputFile) || !terminal(outputFile) {
		return nil
	}
	selector := newTerminalTunnelConnectionSelector(input, output)
	return selector.Select
}

type terminalTunnelConnectionSelector struct {
	input  *bufio.Reader
	output io.Writer
}

func newTerminalTunnelConnectionSelector(input io.Reader, output io.Writer) *terminalTunnelConnectionSelector {
	return &terminalTunnelConnectionSelector{input: bufio.NewReader(input), output: output}
}

func (selector *terminalTunnelConnectionSelector) Select(ctx context.Context, connections []appconfig.TunnelConnection) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	_, _ = fmt.Fprintln(selector.output, "Select a tunnel connection")
	for index, connection := range connections {
		_, _ = fmt.Fprintf(selector.output, "%d) %s  %s\n", index+1, connection.Server, maskTunnelToken(connection.Token))
	}
	for {
		if err := ctx.Err(); err != nil {
			return "", false, err
		}
		_, _ = fmt.Fprint(selector.output, "> ")
		line, err := selector.input.ReadString('\n')
		value := strings.TrimSpace(line)
		if value == "" || isTunnelSelectionCancellation(value) {
			_, _ = fmt.Fprintln(selector.output, "Cancelled")
			return "", true, nil
		}
		index, parseErr := strconv.Atoi(value)
		if parseErr == nil && index >= 1 && index <= len(connections) {
			return connections[index-1].ID, false, nil
		}
		if err != nil {
			_, _ = fmt.Fprintln(selector.output, "Cancelled")
			return "", true, nil
		}
		_, _ = fmt.Fprintln(selector.output, "Invalid selection")
	}
}

func isTunnelSelectionCancellation(value string) bool {
	return strings.EqualFold(value, "q") || strings.EqualFold(value, "quit") || strings.EqualFold(value, "cancel")
}

func maskTunnelToken(token string) string {
	if len(token) <= 12 {
		return strings.Repeat("*", len(token))
	}
	return token[:8] + "********" + token[len(token)-4:]
}
