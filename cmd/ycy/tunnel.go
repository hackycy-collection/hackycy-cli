package main

import (
	"context"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	"github.com/hackycy/hackycy-cli/internal/commands/tunnel"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

func newTunnelServerHandler(logger logging.Logger) cliapp.TunnelServerHandler {
	return func(ctx context.Context, config tunnel.ServerConfig) error {
		return tunnel.RunServer(ctx, config, tunnel.ServerRunOptions{Logger: logger})
	}
}
