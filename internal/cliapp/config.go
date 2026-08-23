package cliapp

import "github.com/spf13/cobra"

func (app *App) registerConfig(root *cobra.Command, configureLogging func() error) {
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
	app.registerConfigFork(configCommand, configureLogging)
	app.registerConfigCM(configCommand, configureLogging)
	root.AddCommand(configCommand)
}
