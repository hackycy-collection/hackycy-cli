package cliapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/logging"
	"github.com/spf13/cobra"
)

// BuildInfo is injected by the product build and presented by the global CLI.
type BuildInfo struct {
	Version string
}

// Dependencies are process facts supplied by the composition root.
type Dependencies struct {
	Out              io.Writer
	Err              io.Writer
	Environment      func(string) string
	Logging          *logging.Runtime
	ExportEnv        ExportEnvHandler
	ConfigForkList   ConfigForkListHandler
	ConfigForkAdd    ConfigForkAddHandler
	ConfigForkRemove ConfigForkRemoveHandler
	ConfigCMList     ConfigCMListHandler
	ConfigCMAdd      ConfigCMAddHandler
	ConfigCMUse      ConfigCMUseHandler
}

// Outcome leaves process exit ownership with cmd/ycy.
type Outcome struct {
	Code int
	Err  error
}

// App owns Cobra, global options, diagnostics, and fresh root-tree construction.
type App struct {
	build            BuildInfo
	out              io.Writer
	err              io.Writer
	environment      func(string) string
	logging          *logging.Runtime
	exportEnv        ExportEnvHandler
	configForkList   ConfigForkListHandler
	configForkAdd    ConfigForkAddHandler
	configForkRemove ConfigForkRemoveHandler
	configCMList     ConfigCMListHandler
	configCMAdd      ConfigCMAddHandler
	configCMUse      ConfigCMUseHandler
}

// New creates the current foundation command tree. Business leaves are added only with their own units.
func New(build BuildInfo, dependencies Dependencies) (*App, error) {
	if strings.TrimSpace(build.Version) == "" {
		return nil, errors.New("CLI build version is required")
	}
	if dependencies.Out == nil {
		dependencies.Out = os.Stdout
	}
	if dependencies.Err == nil {
		dependencies.Err = os.Stderr
	}
	if dependencies.Environment == nil {
		dependencies.Environment = os.Getenv
	}
	if dependencies.Logging == nil {
		dependencies.Logging = logging.NewRuntime(logging.Options{Writer: dependencies.Err})
	}
	return &App{
		build:            build,
		out:              dependencies.Out,
		err:              dependencies.Err,
		environment:      dependencies.Environment,
		logging:          dependencies.Logging,
		exportEnv:        dependencies.ExportEnv,
		configForkList:   dependencies.ConfigForkList,
		configForkAdd:    dependencies.ConfigForkAdd,
		configForkRemove: dependencies.ConfigForkRemove,
		configCMList:     dependencies.ConfigCMList,
		configCMAdd:      dependencies.ConfigCMAdd,
		configCMUse:      dependencies.ConfigCMUse,
	}, nil
}

// Execute builds a fresh Cobra tree for every invocation and never exits the process itself.
func (app *App) Execute(context context.Context, arguments []string) Outcome {
	return app.execute(func() error {
		root := app.rootCommand()
		root.SetArgs(arguments)
		return root.ExecuteContext(context)
	})
}

func (app *App) execute(invoke func() error) (outcome Outcome) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err := fmt.Errorf("%v", recovered)
			app.reportRuntimeError(err)
			outcome = Outcome{Code: 1, Err: err}
		}
	}()

	if err := invoke(); err != nil {
		if errors.Is(err, errHelpRequested) {
			return Outcome{Code: 1, Err: err}
		}
		err = normalizeCobraError(err)
		app.reportError(err)
		return Outcome{Code: 1, Err: err}
	}
	return Outcome{}
}

func (app *App) rootCommand() *cobra.Command {
	var logLevel string
	var showVersion bool
	root := &cobra.Command{
		Use:           "ycy",
		Short:         "Ycy command line interface",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if showVersion {
				_, _ = fmt.Fprintln(app.out, app.build.Version)
				return nil
			}
			if err := app.configureLogging(logLevel); err != nil {
				return err
			}
			if err := command.Help(); err != nil {
				return err
			}
			return errHelpRequested
		},
	}
	root.SetOut(app.out)
	root.SetErr(app.err)
	root.PersistentFlags().StringVar(&logLevel, "log-level", "", "Log level: debug, info, warn, or error")
	root.Flags().BoolVarP(&showVersion, "version", "V", false, "Print version")
	if app.exportEnv != nil {
		app.registerExportEnv(root, func() error {
			return app.configureLogging(logLevel)
		})
	}
	if app.configForkList != nil || app.configForkAdd != nil || app.configForkRemove != nil || app.configCMList != nil || app.configCMAdd != nil || app.configCMUse != nil {
		app.registerConfig(root, func() error {
			return app.configureLogging(logLevel)
		})
	}
	return root
}

func (app *App) configureLogging(value string) error {
	if value == "" {
		value = app.environment("YCY_LOG_LEVEL")
	}
	level, err := logging.ParseLevel(value)
	if err != nil {
		return err
	}
	app.logging.SetLevel(level)
	return nil
}

func (app *App) reportError(err error) {
	_, _ = fmt.Fprintf(app.err, "error: %s\n", err)
}

func (app *App) reportRuntimeError(err error) {
	_, _ = fmt.Fprintln(app.out)
	_, _ = fmt.Fprintf(app.err, "error: %s\n", logging.Redact(err.Error()))
	if app.debugEnabled() {
		_, _ = app.err.Write(debug.Stack())
	}
}

func (app *App) debugEnabled() bool {
	return app.environment("DEBUG") != "" || strings.EqualFold(app.environment("NODE_ENV"), "development")
}

func normalizeCobraError(err error) error {
	message := err.Error()
	if strings.HasPrefix(message, "unknown command ") {
		if start := strings.Index(message, `"`); start >= 0 {
			rest := message[start+1:]
			if end := strings.Index(rest, `"`); end >= 0 {
				return fmt.Errorf("unknown command '%s'", rest[:end])
			}
		}
	}
	return err
}

var errHelpRequested = errors.New("help requested")
