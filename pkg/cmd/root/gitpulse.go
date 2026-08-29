package root

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pulsecommand "github.com/hackycy/hackycy-cli/internal/commands/git/pulse"
	"github.com/spf13/cobra"
)

// GitPulseHandler is the fixed typed handler for git pulse.
type GitPulseHandler func(context.Context, pulsecommand.Input) (pulsecommand.Result, error)

func (app *App) gitPulseCommand(configureLogging func() error) *cobra.Command {
	var days string
	command := &cobra.Command{
		Use:   "pulse [directory]",
		Short: "Show recent Git activity across repositories",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, arguments []string) error {
			input := pulsecommand.Input{}
			if len(arguments) == 1 {
				input.Directory = arguments[0]
			}
			if command.Flags().Changed("days") {
				parsed, err := parsePulseInteger(days)
				if err != nil {
					return err
				}
				input.Days = &parsed
			}
			_, err := app.gitPulse(command.Context(), input)
			return err
		},
	}
	command.Flags().StringVar(&days, "days", "", "Number of days to search")
	return command
}

func parsePulseInteger(value string) (int, error) {
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
