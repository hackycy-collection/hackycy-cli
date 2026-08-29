package root

import (
	"context"
	"fmt"
	"strings"

	runcommand "github.com/hackycy/hackycy-cli/internal/commands/run"
	"github.com/spf13/cobra"
)

// RunHandler is the fixed typed handler for run.
type RunHandler func(context.Context, runcommand.Input) (runcommand.Result, error)

type runChildOutcome struct {
	code int
}

func (outcome *runChildOutcome) Error() string {
	return "run child outcome"
}

func (outcome *runChildOutcome) ExitCode() int {
	return outcome.code
}

func (app *App) registerRun(root *cobra.Command, configureDiagnostics func() error) {
	root.AddCommand(app.runCommand(app.run, configureDiagnostics))
}

func (app *App) runCommand(handler RunHandler, configureDiagnostics func() error) *cobra.Command {
	return &cobra.Command{
		Use:                "run [path]",
		Short:              "Run package.json scripts",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE: func(command *cobra.Command, arguments []string) error {
			parsed, err := parseRunArguments(arguments)
			if err != nil {
				return err
			}
			if parsed.help {
				return command.Help()
			}
			if err := configureDiagnostics(); err != nil {
				return err
			}
			result, err := handler(command.Context(), runcommand.Input{Directory: parsed.directory})
			if err != nil {
				return err
			}
			if result.ExitCode != 0 {
				return &runChildOutcome{code: result.ExitCode}
			}
			return nil
		},
	}
}

type parsedRunArguments struct {
	directory string
	help      bool
}

func parseRunArguments(arguments []string) (parsedRunArguments, error) {
	operands := make([]string, 0, len(arguments))
	parsed := parsedRunArguments{}
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--" {
			operands = append(operands, arguments[index:]...)
			break
		}
		switch {
		case argument == "--log-level", argument == "--log-format":
			if index+1 == len(arguments) {
				return parsedRunArguments{}, fmt.Errorf("flag needs an argument: %s", argument)
			}
			index++
		case strings.HasPrefix(argument, "--log-level="), strings.HasPrefix(argument, "--log-format="), argument == "--verbose", strings.HasPrefix(argument, "--verbose="), argument == "--quiet", strings.HasPrefix(argument, "--quiet="), argument == "-v":
		default:
			operands = append(operands, argument)
		}
	}
	if len(operands) > 1 {
		return parsedRunArguments{}, fmt.Errorf("accepts at most 1 arg(s), received %d", len(operands))
	}
	if len(operands) == 1 {
		if operands[0] == "--help" || operands[0] == "-h" {
			parsed.help = true
			return parsed, nil
		}
		parsed.directory = operands[0]
	}
	return parsed, nil
}
