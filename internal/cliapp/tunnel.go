package cliapp

import (
	"context"

	"github.com/hackycy/hackycy-cli/internal/commands/tunnel"
	"github.com/spf13/cobra"
)

// TunnelServerHandler is the fixed typed handler for tunnel server.
type TunnelServerHandler func(context.Context, tunnel.ServerConfig) error

func (app *App) registerTunnel(root *cobra.Command, configureLogging func() error) {
	address := ""
	controlPort := ""
	frpPort := ""
	httpPort := ""
	portRange := ""
	advertiseFRPAddr := ""
	dataDirectory := ""
	sessionIdleDays := ""

	tunnelCommand := &cobra.Command{
		Use:   "tunnel",
		Short: "Manage trusted tunnel clients and tunnel definitions",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := command.Help(); err != nil {
				return err
			}
			return errHelpRequested
		},
	}
	serverCommand := &cobra.Command{
		Use:   "server",
		Short: "Run the Tunnel Control Plane and supervised frps process",
		Args:  cobra.NoArgs,
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, _ []string) error {
			config, err := tunnel.ResolveServerConfig(tunnel.ServerOptionInput{
				Address:          tunnelServerOption(command, "address", address),
				ControlPort:      tunnelServerOption(command, "control-port", controlPort),
				FRPPort:          tunnelServerOption(command, "frp-port", frpPort),
				HTTPPort:         tunnelServerOption(command, "http-port", httpPort),
				PortRange:        tunnelServerOption(command, "port-range", portRange),
				AdvertiseFRPAddr: tunnelServerOption(command, "advertise-frp-addr", advertiseFRPAddr),
				DataDir:          tunnelServerOption(command, "data-dir", dataDirectory),
				SessionIdleDays:  tunnelServerOption(command, "session-idle-days", sessionIdleDays),
			}, app.environmentLookup)
			if err != nil {
				app.logging.Logger("tunnel.server").Error("Could not resolve tunnel server configuration", map[string]any{"error": err.Error()})
				return err
			}
			return app.tunnelServer(command.Context(), config)
		},
	}
	serverCommand.Flags().StringVar(&address, "address", address, "Address for all tunnel server listeners")
	serverCommand.Flags().StringVar(&controlPort, "control-port", controlPort, "Tunnel Control Plane port")
	serverCommand.Flags().StringVar(&frpPort, "frp-port", frpPort, "FRP bind port")
	serverCommand.Flags().StringVar(&httpPort, "http-port", httpPort, "FRP HTTP vhost port")
	serverCommand.Flags().StringVar(&portRange, "port-range", portRange, "Server Port Pool")
	serverCommand.Flags().StringVar(&advertiseFRPAddr, "advertise-frp-addr", advertiseFRPAddr, "FRP endpoint advertised to trusted clients")
	serverCommand.Flags().StringVar(&dataDirectory, "data-dir", dataDirectory, "Tunnel server state directory")
	serverCommand.Flags().StringVar(&sessionIdleDays, "session-idle-days", sessionIdleDays, "Session idle lifetime in days")
	tunnelCommand.AddCommand(serverCommand)
	root.AddCommand(tunnelCommand)
}

func tunnelServerOption(command *cobra.Command, name, value string) *string {
	if !command.Flags().Changed(name) {
		return nil
	}
	return &value
}
