package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const (
	prototypeVersion  = "0.0.0-prototype"
	internalApplyMode = "__ycy_internal_apply_update"
	defaultRemote     = "origin"
)

type streams struct {
	in          io.Reader
	out         io.Writer
	err         io.Writer
	interactive bool
}

type exitResult struct {
	code    int
	message string
}

type cliAdapter struct {
	streams     streams
	environment map[string]string
	actions     commandActions
	logLevel    string
	version     bool
	root        *cobra.Command
}

func newCLIAdapter(s streams, environment map[string]string) *cliAdapter {
	adapter := &cliAdapter{streams: s, environment: environment}
	adapter.actions = commandActions{
		out:      s.out,
		prompt:   promptAdapter{in: s.in, out: s.out, interactive: s.interactive},
		logLevel: func() string { return adapter.logLevel },
		internal: func(ctx context.Context, args []string) error {
			return adapter.actions.print(ctx, "internal updater probe", args)
		},
	}
	adapter.root = adapter.buildRoot()
	return adapter
}

func (a *cliAdapter) execute(ctx context.Context, argv []string, signalCode func() int) exitResult {
	if len(argv) > 0 && argv[0] == internalApplyMode {
		return classifyExit(a.actions.internal(ctx, slices.Clone(argv[1:])), signalCode())
	}

	normalized := slices.Clone(argv)
	if a.environment["YCY_PROTOTYPE_RAW_COBRA"] != "1" {
		var err error
		normalized, err = normalizeArgv(argv)
		if err != nil {
			return exitResult{code: 1, message: err.Error()}
		}
	}
	a.root.SetArgs(normalized)
	err := a.root.ExecuteContext(ctx)
	return classifyExit(err, signalCode())
}

func (a *cliAdapter) buildRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "ycy-prototype",
		Short:         "Throwaway ycy CLI compatibility probe",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.version {
				fmt.Fprintln(a.streams.out, prototypeVersion)
				return nil
			}
			_ = cmd.Help()
			return usageError{"a command is required"}
		},
	}
	root.SetIn(a.streams.in)
	root.SetOut(a.streams.out)
	root.SetErr(a.streams.err)
	root.PersistentFlags().String("log-level", "", "log level: debug, info, warn, or error")
	root.PersistentFlags().BoolVarP(&a.version, "version", "V", false, "output the version number")
	root.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		if a.version {
			fmt.Fprintln(a.streams.out, prototypeVersion)
			return stopError{}
		}
		level, err := resolveLogLevel(root.Flags().Lookup("log-level").Value.String(), a.environment["YCY_LOG_LEVEL"])
		if err != nil {
			return err
		}
		a.logLevel = level
		return nil
	}

	a.registerGit(root)
	a.registerDiff(root)
	a.registerFS(root)
	a.registerRun(root)
	a.registerPrompt(root)
	a.registerWait(root)
	a.registerFailure(root)
	return root
}

func (a *cliAdapter) registerGit(root *cobra.Command) {
	git := &cobra.Command{
		Use:   "git",
		Short: "Git utilities",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = cmd.Help()
			return usageError{"a git command is required"}
		},
	}
	cm := &cobra.Command{Use: "cm", Args: cobra.NoArgs}
	var push string
	var stagePush string
	var excludes []string
	cm.Flags().StringVarP(&push, "push", "p", "", "optional push remote")
	cm.Flags().StringVarP(&stagePush, "stage-push", "c", "", "optional stage-and-push remote")
	cm.Flags().StringArrayVarP(&excludes, "exclude", "x", nil, "repeatable probe value")
	cm.Flags().Lookup("push").NoOptDefVal = defaultRemote
	cm.Flags().Lookup("stage-push").NoOptDefVal = defaultRemote
	cm.RunE = func(cmd *cobra.Command, _ []string) error {
		return a.actions.gitCM(cmd.Context(), gitCMInput{
			Push:      optionalRemote{Set: cmd.Flags().Changed("push"), Remote: push},
			StagePush: optionalRemote{Set: cmd.Flags().Changed("stage-push"), Remote: stagePush},
			Excludes:  slices.Clone(excludes),
		})
	}
	git.AddCommand(cm)
	root.AddCommand(git)
}

func (a *cliAdapter) registerDiff(root *cobra.Command) {
	excludes := []string{}
	command := &cobra.Command{
		Use:  "diff <baseline> <target>",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.actions.diff(cmd.Context(), diffInput{Baseline: args[0], Target: args[1], Excludes: slices.Clone(excludes)})
		},
	}
	command.Flags().StringArrayVarP(&excludes, "exclude", "x", nil, "repeatable exclusion")
	root.AddCommand(command)
}

func (a *cliAdapter) registerFS(root *cobra.Command) {
	accounts := []string{}
	command := &cobra.Command{
		Use:  "fs [directory]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			directory := "."
			if len(args) == 1 {
				directory = args[0]
			}
			return a.actions.fs(cmd.Context(), fsInput{Directory: directory, Accounts: slices.Clone(accounts)})
		},
	}
	command.Flags().StringArrayVar(&accounts, "account", nil, "repeatable account")
	root.AddCommand(command)
}

func (a *cliAdapter) registerRun(root *cobra.Command) {
	command := &cobra.Command{
		Use: "run [path] -- [args...]",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := parseRunArgs(args, cmd.ArgsLenAtDash())
			if err != nil {
				return err
			}
			return a.actions.run(cmd.Context(), input)
		},
	}
	root.AddCommand(command)
}

func (a *cliAdapter) registerPrompt(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:  "prompt",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.actions.choose(cmd.Context())
		},
	})
}

func (a *cliAdapter) registerWait(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:  "wait",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.actions.wait(cmd.Context())
		},
	})
}

func (a *cliAdapter) registerFailure(root *cobra.Command) {
	var kind string
	var childCode int
	command := &cobra.Command{
		Use:  "fail",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if kind == "deadline" {
				ctx, cancel := newDeadlineContext(cmd.Context())
				defer cancel()
				<-ctx.Done()
			}
			return failureAction(kind, childCode)
		},
	}
	command.Flags().StringVar(&kind, "kind", "action", "action, cancel, child, or deadline")
	command.Flags().IntVar(&childCode, "child-code", 7, "child status for child failure")
	root.AddCommand(command)
}

func normalizeArgv(argv []string) ([]string, error) {
	normalized := slices.Clone(argv)
	if len(normalized) >= 2 && normalized[0] == "git" && normalized[1] == "cm" {
		normalized = normalizeOptionalRemotes(normalized)
	}
	return normalized, nil
}

func normalizeOptionalRemotes(argv []string) []string {
	result := make([]string, 0, len(argv))
	for index := 0; index < len(argv); index++ {
		token := argv[index]
		for _, candidate := range []struct {
			short string
			long  string
		}{
			{short: "-p", long: "--push"},
			{short: "-c", long: "--stage-push"},
		} {
			if !strings.HasPrefix(token, candidate.short) || len(token) == len(candidate.short) {
				continue
			}
			value := token[len(candidate.short):]
			if !strings.HasPrefix(value, "-") && !strings.HasPrefix(value, "=") {
				token = candidate.long + "=" + value
			}
			break
		}
		if (token == "--push" || token == "-p" || token == "--stage-push" || token == "-c") && index+1 < len(argv) {
			next := argv[index+1]
			if next != "--" && next != "-" && !strings.HasPrefix(next, "-") {
				name := token
				if token == "-p" {
					name = "--push"
				}
				if token == "-c" {
					name = "--stage-push"
				}
				result = append(result, name+"="+next)
				index++
				continue
			}
		}
		result = append(result, token)
	}
	return result
}

func parseRunArgs(args []string, argsBeforeDash int) (runInput, error) {
	if argsBeforeDash == -1 {
		if len(args) > 1 {
			return runInput{}, usageError{"run accepts at most one path unless passthrough starts after --"}
		}
		input := runInput{Passthrough: []string{}}
		if len(args) == 1 {
			input.Path = args[0]
		}
		return input, nil
	}
	if argsBeforeDash > 1 {
		return runInput{}, usageError{"run accepts at most one path before --"}
	}
	input := runInput{Passthrough: slices.Clone(args[argsBeforeDash:])}
	if argsBeforeDash == 1 {
		input.Path = args[0]
	}
	return input, nil
}

func resolveLogLevel(cliValue, environmentValue string) (string, error) {
	value := cliValue
	if value == "" {
		value = environmentValue
	}
	if value == "" {
		value = "info"
	}
	value = strings.ToLower(strings.TrimSpace(value))
	for _, allowed := range []string{"debug", "info", "warn", "error"} {
		if value == allowed {
			return value, nil
		}
	}
	return "", fmt.Errorf("invalid log level %q", value)
}

func classifyExit(err error, signalCode int) exitResult {
	if err == nil || errors.As(err, new(stopError)) {
		return exitResult{code: 0}
	}
	if errors.Is(err, errUserCancelled) {
		return exitResult{code: 0, message: "Cancelled."}
	}
	if errors.Is(err, context.Canceled) && signalCode != 0 {
		return exitResult{code: signalCode, message: "Interrupted."}
	}
	var child childExitError
	if errors.As(err, &child) {
		return exitResult{code: child.code, message: child.Error()}
	}
	return exitResult{code: 1, message: err.Error()}
}

type usageError struct {
	message string
}

func (e usageError) Error() string { return e.message }

type stopError struct{}

func (stopError) Error() string { return "stop after output" }

func environmentMap(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, entry := range values {
		name, value, ok := strings.Cut(entry, "=")
		if ok {
			result[name] = value
		}
	}
	return result
}
