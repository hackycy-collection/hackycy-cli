package upgrade

import (
	"context"
	"errors"

	"github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/updater"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Options contains the parsed Upgrade request and leaf-owned presentation.
type Options struct {
	Context        context.Context
	Terminal       *terminal.Runtime
	CurrentVersion string
}

// NewCmdUpgrade creates the Upgrade leaf with an optional test runner.
func NewCmdUpgrade(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runUpgrade
	}
	return &cobra.Command{
		Use:          "upgrade",
		Short:        "Update ycy to the latest release",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(command *cobra.Command, _ []string) error {
			if factory == nil || factory.Terminal == nil || factory.Version == "" {
				return errors.New("upgrade Factory is incomplete")
			}
			return runF(&Options{
				Context:        command.Context(),
				Terminal:       factory.Terminal,
				CurrentVersion: factory.Version,
			})
		},
	}
}

func runUpgrade(options *Options) error {
	if options == nil || options.Terminal == nil {
		return errors.New("upgrade options are incomplete")
	}
	result, err := updater.RunUpgrade(options.Context, updater.UpgradeOptions{
		Resolver: updater.ReleaseResolverOptions{CurrentVersion: options.CurrentVersion},
	})
	return PresentResult(options.Context, options.Terminal, result, err)
}
