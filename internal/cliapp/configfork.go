package cliapp

import (
	"context"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	"github.com/spf13/cobra"
)

// ConfigForkListHandler is the fixed typed handler for config fork list.
type ConfigForkListHandler func(context.Context, configfork.Input) (configfork.Result, error)

// ConfigForkAddHandler is the fixed typed handler for config fork add.
type ConfigForkAddHandler func(context.Context, configfork.AddRequest) (configfork.AddResult, error)

// ConfigForkRemoveHandler is the fixed typed handler for config fork remove.
type ConfigForkRemoveHandler func(context.Context, configfork.RemoveRequest) (configfork.RemoveResult, error)

func (app *App) registerConfigFork(configCommand *cobra.Command, configureLogging func() error) {
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
	if app.configForkList != nil {
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
	}
	if app.configForkAdd != nil {
		addCommand := &cobra.Command{
			Use:   "add",
			Short: "Add a provider instance",
			Args:  cobra.NoArgs,
			PreRunE: func(*cobra.Command, []string) error {
				return configureLogging()
			},
			RunE: func(command *cobra.Command, _ []string) error {
				_, err := app.configForkAdd(command.Context(), configfork.AddRequest{})
				return err
			},
		}
		forkCommand.AddCommand(addCommand)
	}
	if app.configForkRemove != nil {
		removeCommand := &cobra.Command{
			Use:   "remove",
			Short: "Remove a provider instance",
			Args:  cobra.NoArgs,
			PreRunE: func(*cobra.Command, []string) error {
				return configureLogging()
			},
			RunE: func(command *cobra.Command, _ []string) error {
				_, err := app.configForkRemove(command.Context(), configfork.RemoveRequest{})
				return err
			},
		}
		forkCommand.AddCommand(removeCommand)
	}
	configCommand.AddCommand(forkCommand)
}
