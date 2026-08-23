package cliapp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	heatcommand "github.com/hackycy/hackycy-cli/internal/commands/git/heat"
	"github.com/spf13/cobra"
)

// GitHeatHandler is the fixed typed handler for git heat.
type GitHeatHandler func(context.Context, heatcommand.Input) (heatcommand.Result, error)

func (app *App) registerGit(root *cobra.Command, configureLogging func() error) {
	git := &cobra.Command{
		Use:   "git",
		Short: "Git utilities",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := command.Help(); err != nil {
				return err
			}
			return errHelpRequested
		},
	}
	git.AddCommand(app.gitHeatCommand(configureLogging))
	root.AddCommand(git)
}

func (app *App) gitHeatCommand(configureLogging func() error) *cobra.Command {
	var limit string
	var days string
	var target string
	var sort string
	var relativeTime bool
	var query string
	command := &cobra.Command{
		Use:   "heat",
		Short: "Show frequently changed files and directories in recent commits",
		Args:  cobra.NoArgs,
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, _ []string) error {
			input := heatcommand.Input{
				Target:       heatcommand.TargetDirectories,
				Sort:         heatcommand.SortPath,
				RelativeTime: relativeTime,
				Query:        query,
			}
			if command.Flags().Changed("limit") {
				parsed, err := parseHeatInteger(limit)
				if err != nil {
					return err
				}
				input.Limit = &parsed
			}
			if command.Flags().Changed("days") {
				parsed, err := parseHeatInteger(days)
				if err != nil {
					return err
				}
				input.Days = &parsed
			}
			if command.Flags().Changed("type") {
				parsed, err := parseHeatTarget(target)
				if err != nil {
					return err
				}
				input.Target = parsed
			}
			if command.Flags().Changed("sort") {
				parsed, err := parseHeatSort(sort)
				if err != nil {
					return err
				}
				input.Sort = parsed
			}
			_, err := app.gitHeat(command.Context(), input)
			return err
		},
	}
	command.Flags().StringVarP(&limit, "limit", "n", "", "Number of recent commits to inspect")
	command.Flags().StringVarP(&days, "days", "d", "", "Number of recent days to inspect")
	command.Flags().StringVarP(&target, "type", "t", "", "Report type: files or directories")
	command.Flags().StringVarP(&sort, "sort", "s", "", "Sort by count or path")
	command.Flags().BoolVarP(&relativeTime, "relative-time", "r", false, "Show Changed at as relative time")
	command.Flags().StringVarP(&query, "query", "q", "", "Highlight files or directories that contain text")
	return command
}

func parseHeatInteger(value string) (int, error) {
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

func parseHeatTarget(value string) (heatcommand.Target, error) {
	if value == string(heatcommand.TargetFiles) {
		return heatcommand.TargetFiles, nil
	}
	if value == string(heatcommand.TargetDirectories) || value == "dirs" {
		return heatcommand.TargetDirectories, nil
	}
	return "", fmt.Errorf("'%s' is not a valid report type. Use files or directories.", value)
}

func parseHeatSort(value string) (heatcommand.Sort, error) {
	if value == string(heatcommand.SortCount) {
		return heatcommand.SortCount, nil
	}
	if value == string(heatcommand.SortPath) {
		return heatcommand.SortPath, nil
	}
	return "", fmt.Errorf("'%s' is not a valid sort. Use count or path.", value)
}
