package cliapp

import (
	"context"

	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	"github.com/spf13/cobra"
)

// ConfigCMListHandler is the fixed typed handler for config cm list.
type ConfigCMListHandler func(context.Context, configcm.Input) (configcm.Result, error)

// ConfigCMAddHandler is the fixed typed handler for config cm add.
type ConfigCMAddHandler func(context.Context, configcm.AddRequest) (configcm.AddResult, error)

// ConfigCMUseHandler is the fixed typed handler for config cm use.
type ConfigCMUseHandler func(context.Context, configcm.UseRequest) (configcm.UseResult, error)

// ConfigCMSetHandler is the fixed typed handler for config cm set.
type ConfigCMSetHandler func(context.Context, configcm.SetRequest) (configcm.SetResult, error)

func (app *App) registerConfigCM(configCommand *cobra.Command, configureLogging func() error) {
	if app.configCMList == nil && app.configCMAdd == nil && app.configCMUse == nil && app.configCMSet == nil {
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
	if app.configCMList != nil {
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
	}
	if app.configCMAdd != nil {
		addCommand := &cobra.Command{
			Use:   "add",
			Short: "Add a CM profile",
			Args:  cobra.NoArgs,
			PreRunE: func(*cobra.Command, []string) error {
				return configureLogging()
			},
			RunE: func(command *cobra.Command, _ []string) error {
				_, err := app.configCMAdd(command.Context(), configcm.AddRequest{})
				return err
			},
		}
		cmCommand.AddCommand(addCommand)
	}
	if app.configCMUse != nil {
		useCommand := &cobra.Command{
			Use:   "use <profile>",
			Short: "Set the default commit message provider profile",
			Args:  cobra.ExactArgs(1),
			PreRunE: func(*cobra.Command, []string) error {
				return configureLogging()
			},
			RunE: func(command *cobra.Command, arguments []string) error {
				_, err := app.configCMUse(command.Context(), configcm.UseRequest{Profile: arguments[0]})
				return err
			},
		}
		cmCommand.AddCommand(useCommand)
	}
	if app.configCMSet != nil {
		setCommand := &cobra.Command{
			Use:   "set <profile> <key> <value>",
			Short: "Set an optional commit message provider profile value",
			Args:  cobra.ExactArgs(3),
			PreRunE: func(*cobra.Command, []string) error {
				return configureLogging()
			},
			RunE: func(command *cobra.Command, arguments []string) error {
				_, err := app.configCMSet(command.Context(), configcm.SetRequest{
					Profile: arguments[0],
					Key:     arguments[1],
					Value:   arguments[2],
				})
				return err
			},
		}
		cmCommand.AddCommand(setCommand)
	}
	configCommand.AddCommand(cmCommand)
}
