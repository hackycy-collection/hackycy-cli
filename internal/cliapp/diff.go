package cliapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/hackycy/hackycy-cli/internal/commands/diff"
	"github.com/spf13/cobra"
)

// DiffHandler is the fixed typed handler for diff.
type DiffHandler func(context.Context, diff.Input) (diff.Result, error)

func (app *App) registerDiff(root *cobra.Command, configureLogging func() error) {
	port := "1205"
	exclusions := []string{}
	var public bool
	var noGitIgnore bool
	command := &cobra.Command{
		Use:   "diff <baseline-directory> <target-directory>",
		Short: "Compare two directories in a browser",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, arguments []string) error {
			parsedPort, err := parseDiffPort(port)
			if err != nil {
				return err
			}
			_, err = app.diff(command.Context(), diff.Input{
				BaselineDirectory: arguments[0],
				TargetDirectory:   arguments[1],
				Port:              parsedPort,
				Public:            public,
				Exclusions:        append([]string{}, exclusions...),
				NoGitIgnore:       noGitIgnore,
			})
			return err
		},
	}
	command.Flags().StringVarP(&port, "port", "p", port, "Port to serve on")
	command.Flags().BoolVar(&public, "public", false, "Make the diff available on the local network")
	command.Flags().StringArrayVarP(&exclusions, "exclude", "x", exclusions, "Add an exclusion")
	command.Flags().BoolVar(&noGitIgnore, "no-gitignore", false, "Do not apply Target Directory .gitignore files")
	root.AddCommand(command)
}

func parseDiffPort(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("'%s' is not a valid port", value)
	}
	port := 0
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character < '0' || character > '9' {
			return 0, fmt.Errorf("'%s' is not a valid port", value)
		}
		port = port*10 + int(character-'0')
		if port > 65535 {
			return 0, errors.New("Port must be between 0 and 65535")
		}
	}
	return port, nil
}
