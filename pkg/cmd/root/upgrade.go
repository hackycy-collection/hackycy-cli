package root

import (
	"context"

	"github.com/spf13/cobra"
)

// UpgradeHandler is the typed public Go-to-Go update boundary.
type UpgradeHandler func(context.Context) error

func (app *App) registerUpgrade(root *cobra.Command, configureLogging func() error) {
	command := &cobra.Command{
		Use:          "upgrade",
		Short:        "Update ycy to the latest release",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, _ []string) error {
			return app.upgrade(command.Context())
		},
	}
	root.AddCommand(command)
}
