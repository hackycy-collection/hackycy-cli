package cliapp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/commands/rm"
	"github.com/spf13/cobra"
)

// RmHandler is the fixed typed handler for rm.
type RmHandler func(context.Context, rm.Input) (rm.Result, error)

func (app *App) registerRM(root *cobra.Command, configureLogging func() error) {
	var force bool
	var depth string
	rmCommand := &cobra.Command{
		Use:   "rm [paths...]",
		Short: "Remove files/dirs, or smartly clean project artifacts when no path given",
		Args:  cobra.ArbitraryArgs,
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, arguments []string) error {
			input := rm.Input{Paths: arguments, Force: force}
			if command.Flags().Changed("depth") {
				parsed, err := parseRMDepth(depth)
				if err != nil {
					return err
				}
				input.Depth = &parsed
			}
			_, err := app.rm(command.Context(), input)
			return err
		},
	}
	rmCommand.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	rmCommand.Flags().StringVarP(&depth, "depth", "d", "", "Smart scan depth (default: 5)")
	root.AddCommand(rmCommand)
}

func parseRMDepth(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("'%s' is not a valid integer", value)
	}
	end := 0
	if trimmed[0] == '+' || trimmed[0] == '-' {
		end++
	}
	for end < len(trimmed) && trimmed[end] >= '0' && trimmed[end] <= '9' {
		end++
	}
	if end == 0 || (end == 1 && (trimmed[0] == '+' || trimmed[0] == '-')) {
		return 0, fmt.Errorf("'%s' is not a valid integer", value)
	}
	parsed, err := strconv.ParseInt(trimmed[:end], 10, 0)
	if err != nil {
		return 0, fmt.Errorf("'%s' is not a valid integer", value)
	}
	return int(parsed), nil
}
