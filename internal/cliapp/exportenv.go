package cliapp

import (
	"context"

	"github.com/hackycy/hackycy-cli/internal/commands/exportenv"
	"github.com/spf13/cobra"
)

// ExportEnvHandler is the fixed typed handler for the export env leaf.
type ExportEnvHandler func(context.Context, exportenv.Input) (exportenv.Result, error)

func (app *App) registerExportEnv(root *cobra.Command, configureLogging func() error) {
	var environment string
	var merge bool
	var output string

	exportCommand := &cobra.Command{
		Use:   "export",
		Short: "Export command output",
	}
	envCommand := &cobra.Command{
		Use:   "env [dir]",
		Short: "Export .env file contents as JSON",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, arguments []string) error {
			var directory string
			if len(arguments) == 1 {
				directory = arguments[0]
			}
			_, err := app.exportEnv(command.Context(), exportenv.Input{
				Directory:   directory,
				Environment: environment,
				Merge:       merge,
				Output:      output,
			})
			return err
		},
	}
	envCommand.Flags().StringVarP(&environment, "env", "e", "", "Environment name, skip interactive selection (e.g. local, prod)")
	envCommand.Flags().BoolVar(&merge, "merge", false, "Merge .env with the selected environment file")
	envCommand.Flags().StringVarP(&output, "out", "o", "", "Write output to file instead of stdout")
	exportCommand.AddCommand(envCommand)
	root.AddCommand(exportCommand)
}
