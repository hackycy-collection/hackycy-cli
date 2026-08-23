package cliapp

import (
	"context"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	"github.com/spf13/cobra"
)

// ConfigForkListHandler is the fixed typed handler for config fork list.
type ConfigForkListHandler func(context.Context, configfork.Input) (configfork.Result, error)

func (app *App) registerConfigForkList(root *cobra.Command, configureLogging func() error) {
	configCommand := &cobra.Command{
		Use:   "config",
		Short: "Manage ycy configuration",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := command.Help(); err != nil {
				return err
			}
			return errHelpRequested
		},
	}
	forkCommand := &cobra.Command{
		Use:   "fork",
		Short: "Manage git fork provider instances",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := command.Help(); err != nil {
				return err
			}
			return errHelpRequested
		},
	}
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List configured provider instances",
		Args:  cobra.NoArgs,
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, _ []string) error {
			_, err := app.configForkList(command.Context(), configfork.Input{})
			return err
		},
	}
	forkCommand.AddCommand(listCommand)
	configCommand.AddCommand(forkCommand)
	root.AddCommand(configCommand)
}
