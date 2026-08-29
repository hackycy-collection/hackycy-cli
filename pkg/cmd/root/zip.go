package root

import (
	"context"

	zipcommand "github.com/hackycy/hackycy-cli/internal/commands/zip"
	"github.com/spf13/cobra"
)

// ZipHandler is the fixed typed handler for zip.
type ZipHandler func(context.Context, zipcommand.Input) (zipcommand.Result, error)

func (app *App) registerZIP(root *cobra.Command, configureLogging func() error) {
	var withoutOpen bool
	var withDir string
	command := &cobra.Command{
		Use:   "zip [directory]",
		Short: "Zip a directory into a zip file",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, arguments []string) error {
			directory := ""
			if len(arguments) == 1 {
				directory = arguments[0]
			}
			_, err := app.zip(command.Context(), zipcommand.Input{
				Directory: directory,
				Open:      !withoutOpen,
				WithDir:   withDir,
			})
			return err
		},
	}
	command.Flags().BoolVarP(&withoutOpen, "without-open", "w", false, "Do not open the zip file after creation")
	command.Flags().StringVarP(&withDir, "with-dir", "d", "", "Include the directory name as a top-level folder in the zip")
	root.AddCommand(command)
}
