package cliapp

import (
	"context"

	forkcommand "github.com/hackycy/hackycy-cli/internal/commands/git/fork"
	"github.com/spf13/cobra"
)

// GitForkHandler is the fixed typed handler for git fork.
type GitForkHandler func(context.Context, forkcommand.Input) (forkcommand.Result, error)

func (app *App) gitForkCommand(configureLogging func() error) *cobra.Command {
	command := &cobra.Command{
		Use:   "fork <repo> [dest]",
		Short: "Download a repo without git history (supports GitHub/GitLab, public/private)",
		Args:  cobra.RangeArgs(1, 2),
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, arguments []string) error {
			input := forkcommand.Input{Repository: arguments[0]}
			if len(arguments) == 2 {
				input.Destination = arguments[1]
			}
			_, err := app.gitFork(command.Context(), input)
			return err
		},
	}
	return command
}
