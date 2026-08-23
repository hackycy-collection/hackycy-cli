package cliapp

import (
	"context"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	"github.com/spf13/cobra"
)

// ConfigCMListHandler is the fixed typed handler for config cm list.
type ConfigCMListHandler func(context.Context, configcm.Input) (configcm.Result, error)

func (app *App) registerConfigCM(configCommand *cobra.Command, configureLogging func() error) {
	if app.configCMList == nil {
		return
	}
	cmCommand := &cobra.Command{
		Use:   "cm",
		Short: "Manage CM profiles",
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
		Short: "List configured CM profiles",
		Args:  cobra.NoArgs,
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, _ []string) error {
			_, err := app.configCMList(command.Context(), configcm.Input{})
			return err
		},
	}
	cmCommand.AddCommand(listCommand)
	configCommand.AddCommand(cmCommand)
}
