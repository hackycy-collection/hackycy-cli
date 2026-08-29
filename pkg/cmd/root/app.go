package root

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/logging"
	"github.com/hackycy/hackycy-cli/internal/terminal"
	configcommand "github.com/hackycy/hackycy-cli/pkg/cmd/config"
	exportcommand "github.com/hackycy/hackycy-cli/pkg/cmd/export"
	rmcommand "github.com/hackycy/hackycy-cli/pkg/cmd/rm"
	runcommand "github.com/hackycy/hackycy-cli/pkg/cmd/run"
	zipcommand "github.com/hackycy/hackycy-cli/pkg/cmd/zip"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Dependencies are the temporary unmigrated command handlers supplied by the
// composition root. Process facts belong exclusively to cmdutil.Factory.
type Dependencies struct {
	GitHeat       GitHeatHandler
	GitPulse      GitPulseHandler
	GitFork       GitForkHandler
	GitCM         GitCMHandler
	Diff          DiffHandler
	FS            FSHandler
	TunnelServer  TunnelServerHandler
	TunnelConnect TunnelConnectHandler
	Upgrade       UpgradeHandler
}

// Outcome leaves process exit ownership with cmd/ycy.
type Outcome struct {
	Code int
	Err  error
}

// App owns Cobra, global options, diagnostics, and fresh root-tree construction.
type App struct {
	factory       *cmdutil.Factory
	gitHeat       GitHeatHandler
	gitPulse      GitPulseHandler
	gitFork       GitForkHandler
	gitCM         GitCMHandler
	diff          DiffHandler
	fs            FSHandler
	tunnelServer  TunnelServerHandler
	tunnelConnect TunnelConnectHandler
	upgrade       UpgradeHandler
}

// New creates the lifted root with the bounded Factory and current temporary handlers.
func New(factory *cmdutil.Factory, dependencies Dependencies) (*App, error) {
	if factory == nil {
		return nil, errors.New("command Factory is required")
	}
	if strings.TrimSpace(factory.Version) == "" {
		return nil, errors.New("CLI build version is required")
	}
	if factory.IOStreams.Out == nil || factory.Terminal == nil || factory.Logging == nil || factory.Environment == nil || factory.EnvironmentLookup == nil {
		return nil, errors.New("command Factory is incomplete")
	}
	return &App{
		factory:       factory,
		gitHeat:       dependencies.GitHeat,
		gitPulse:      dependencies.GitPulse,
		gitFork:       dependencies.GitFork,
		gitCM:         dependencies.GitCM,
		diff:          dependencies.Diff,
		fs:            dependencies.FS,
		tunnelServer:  dependencies.TunnelServer,
		tunnelConnect: dependencies.TunnelConnect,
		upgrade:       dependencies.Upgrade,
	}, nil
}

func (app *App) output() io.Writer {
	return app.factory.IOStreams.Out
}

func (app *App) diagnostics() io.Writer {
	return app.factory.Terminal.DiagnosticWriter()
}

// Execute builds a fresh Cobra tree for every invocation and never exits the process itself.
func (app *App) Execute(context context.Context, arguments []string) Outcome {
	arguments = normalizeDiagnosticAliases(arguments)
	controls := collectDiagnosticControls(arguments)
	root := app.rootCommandWithDiagnosticControls(controls)
	return app.execute(func() error {
		if app.gitCM != nil {
			arguments = normalizeGitCMArguments(arguments)
		}
		root.SetArgs(arguments)
		return normalizeCobraError(root, arguments, root.ExecuteContext(context))
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
		var exitOutcome exitCodedError
		if errors.As(err, &exitOutcome) {
			return Outcome{Code: exitOutcome.ExitCode()}
		}
		if errors.Is(err, errHelpRequested) {
			return Outcome{Code: 1, Err: err}
		}
		app.reportError(err)
		return Outcome{Code: 1, Err: err}
	}
	return Outcome{}
}

type exitCodedError interface {
	error
	ExitCode() int
}

func (app *App) rootCommand() *cobra.Command {
	return app.rootCommandWithDiagnosticControls(diagnosticControls{})
}

func (app *App) rootCommandWithDiagnosticControls(controls diagnosticControls) *cobra.Command {
	var logLevel string
	var logFormat string
	var verbose bool
	var quiet bool
	var showVersion bool
	configureDiagnostics := func() error {
		return app.configureDiagnosticLogging(controls)
	}
	root := &cobra.Command{
		Use:           "ycy",
		Short:         "Ycy command line interface",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if showVersion {
				_, _ = fmt.Fprintln(app.output(), app.factory.Version)
				return nil
			}
			if err := command.Help(); err != nil {
				return err
			}
			return errHelpRequested
		},
	}
	root.SetOut(app.output())
	root.SetErr(app.diagnostics())
	root.PersistentPreRunE = func(command *cobra.Command, _ []string) error {
		switch command.CommandPath() {
		case "ycy rm", "ycy export env", "ycy run", "ycy zip",
			"ycy config fork list", "ycy config fork add", "ycy config fork remove",
			"ycy config cm list", "ycy config cm add", "ycy config cm use",
			"ycy config cm set", "ycy config cm remove", "ycy config cm test":
			return configureDiagnostics()
		}
		return nil
	}
	root.PersistentFlags().StringVar(&logLevel, "log-level", "", "Log level: debug, info, warn, or error")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug diagnostics")
	root.PersistentFlags().BoolVar(&quiet, "quiet", false, "Only emit error diagnostics (short form: -q)")
	root.PersistentFlags().StringVar(&logFormat, "log-format", "", "Diagnostic format: text or json")
	root.Flags().BoolVarP(&showVersion, "version", "V", false, "Print version")
	root.AddCommand(exportcommand.NewCmdExport(app.factory))
	configCommand := configcommand.NewCmdConfig(app.factory)
	root.AddCommand(configCommand)
	root.AddCommand(rmcommand.NewCmdRM(app.factory, nil))
	root.AddCommand(runcommand.NewCmdRun(app.factory, nil))
	root.AddCommand(zipcommand.NewCmdZIP(app.factory, nil))
	if app.gitHeat != nil || app.gitPulse != nil || app.gitFork != nil || app.gitCM != nil {
		app.registerGit(root, configureDiagnostics)
	}
	if app.diff != nil {
		app.registerDiff(root, configureDiagnostics)
	}
	if app.fs != nil {
		app.registerFS(root, configureDiagnostics)
	}
	if app.tunnelServer != nil || app.tunnelConnect != nil {
		app.registerTunnel(root, configureDiagnostics)
	}
	if app.upgrade != nil {
		app.registerUpgrade(root, configureDiagnostics)
	}
	if app.factory.Terminal.Session().Kind == terminal.RichInteractive {
		discovery := newTerminalDiscoveryAdapter(app.factory.Terminal)
		root.SetHelpFunc(func(command *cobra.Command, _ []string) {
			discovery.PresentDiscovery(command.Context(), newDiscoveryDocument(command))
		})
	}
	return root
}

func (app *App) configureDiagnosticLogging(controls diagnosticControls) error {
	configuration, err := logging.ParseConfiguration(logging.ConfigurationInput{
		LogLevels:    controls.logLevels,
		LogFormats:   controls.logFormats,
		VerboseCount: controls.verboseCount,
		QuietCount:   controls.quietCount,
		LookupEnv:    app.factory.EnvironmentLookup,
	})
	if err != nil {
		return err
	}
	app.factory.Logging.SetLevel(configuration.Level)
	app.factory.Logging.SetFormat(configuration.Format)
	return nil
}

func (app *App) reportError(err error) {
	_, _ = fmt.Fprintf(app.diagnostics(), "error: %s\n", err)
}

func (app *App) reportRuntimeError(err error) {
	_, _ = fmt.Fprintln(app.output())
	_, _ = fmt.Fprintf(app.diagnostics(), "error: %s\n", logging.Redact(err.Error()))
	if app.debugEnabled() {
		_, _ = app.diagnostics().Write(debug.Stack())
	}
}

func (app *App) debugEnabled() bool {
	return app.factory.Environment("DEBUG") != "" || strings.EqualFold(app.factory.Environment("NODE_ENV"), "development")
}

var errHelpRequested = errors.New("help requested")
