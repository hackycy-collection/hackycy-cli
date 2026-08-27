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
	Out               io.Writer
	Err               io.Writer
	Environment       func(string) string
	EnvironmentLookup func(string) (string, bool)
	Logging           *logging.Runtime
	Discovery         DiscoveryPresenter
	ExportEnv         ExportEnvHandler
	ConfigForkList    ConfigForkListHandler
	ConfigForkAdd     ConfigForkAddHandler
	ConfigForkRemove  ConfigForkRemoveHandler
	ConfigCMList      ConfigCMListHandler
	ConfigCMAdd       ConfigCMAddHandler
	ConfigCMUse       ConfigCMUseHandler
	ConfigCMSet       ConfigCMSetHandler
	ConfigCMRemove    ConfigCMRemoveHandler
	ConfigCMTest      ConfigCMTestHandler
	RM                RmHandler
	Run               RunHandler
	GitHeat           GitHeatHandler
	GitPulse          GitPulseHandler
	GitFork           GitForkHandler
	GitCM             GitCMHandler
	ZIP               ZipHandler
	Diff              DiffHandler
	FS                FSHandler
	TunnelServer      TunnelServerHandler
	TunnelConnect     TunnelConnectHandler
	Upgrade           UpgradeHandler
}

// Outcome leaves process exit ownership with cmd/ycy.
type Outcome struct {
	Code int
	Err  error
}

// App owns Cobra, global options, diagnostics, and fresh root-tree construction.
type App struct {
	build             BuildInfo
	out               io.Writer
	err               io.Writer
	environment       func(string) string
	environmentLookup func(string) (string, bool)
	logging           *logging.Runtime
	discovery         DiscoveryPresenter
	exportEnv         ExportEnvHandler
	configForkList    ConfigForkListHandler
	configForkAdd     ConfigForkAddHandler
	configForkRemove  ConfigForkRemoveHandler
	configCMList      ConfigCMListHandler
	configCMAdd       ConfigCMAddHandler
	configCMUse       ConfigCMUseHandler
	configCMSet       ConfigCMSetHandler
	configCMRemove    ConfigCMRemoveHandler
	configCMTest      ConfigCMTestHandler
	rm                RmHandler
	run               RunHandler
	gitHeat           GitHeatHandler
	gitPulse          GitPulseHandler
	gitFork           GitForkHandler
	gitCM             GitCMHandler
	zip               ZipHandler
	diff              DiffHandler
	fs                FSHandler
	tunnelServer      TunnelServerHandler
	tunnelConnect     TunnelConnectHandler
	upgrade           UpgradeHandler
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
	providedEnvironment := dependencies.Environment != nil
	if !providedEnvironment {
		dependencies.Environment = os.Getenv
	}
	if dependencies.EnvironmentLookup == nil {
		if !providedEnvironment {
			dependencies.EnvironmentLookup = os.LookupEnv
		} else {
			dependencies.EnvironmentLookup = func(key string) (string, bool) {
				value := dependencies.Environment(key)
				return value, value != ""
			}
		}
	}
	if dependencies.Logging == nil {
		dependencies.Logging = logging.NewRuntime(logging.Options{Writer: dependencies.Err})
	}
	return &App{
		build:             build,
		out:               dependencies.Out,
		err:               dependencies.Err,
		environment:       dependencies.Environment,
		environmentLookup: dependencies.EnvironmentLookup,
		logging:           dependencies.Logging,
		discovery:         dependencies.Discovery,
		exportEnv:         dependencies.ExportEnv,
		configForkList:    dependencies.ConfigForkList,
		configForkAdd:     dependencies.ConfigForkAdd,
		configForkRemove:  dependencies.ConfigForkRemove,
		configCMList:      dependencies.ConfigCMList,
		configCMAdd:       dependencies.ConfigCMAdd,
		configCMUse:       dependencies.ConfigCMUse,
		configCMSet:       dependencies.ConfigCMSet,
		configCMRemove:    dependencies.ConfigCMRemove,
		configCMTest:      dependencies.ConfigCMTest,
		rm:                dependencies.RM,
		run:               dependencies.Run,
		gitHeat:           dependencies.GitHeat,
		gitPulse:          dependencies.GitPulse,
		gitFork:           dependencies.GitFork,
		gitCM:             dependencies.GitCM,
		zip:               dependencies.ZIP,
		diff:              dependencies.Diff,
		fs:                dependencies.FS,
		tunnelServer:      dependencies.TunnelServer,
		tunnelConnect:     dependencies.TunnelConnect,
		upgrade:           dependencies.Upgrade,
	}, nil
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
				_, _ = fmt.Fprintln(app.out, app.build.Version)
				return nil
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
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug diagnostics")
	root.PersistentFlags().BoolVar(&quiet, "quiet", false, "Only emit error diagnostics (short form: -q)")
	root.PersistentFlags().StringVar(&logFormat, "log-format", "", "Diagnostic format: text or json")
	root.Flags().BoolVarP(&showVersion, "version", "V", false, "Print version")
	if app.exportEnv != nil {
		app.registerExportEnv(root, configureDiagnostics)
	}
	if app.configForkList != nil || app.configForkAdd != nil || app.configForkRemove != nil || app.configCMList != nil || app.configCMAdd != nil || app.configCMUse != nil || app.configCMSet != nil || app.configCMRemove != nil || app.configCMTest != nil {
		app.registerConfig(root, configureDiagnostics)
	}
	if app.rm != nil {
		app.registerRM(root, configureDiagnostics)
	}
	if app.run != nil {
		app.registerRun(root, configureDiagnostics)
	}
	if app.gitHeat != nil || app.gitPulse != nil || app.gitFork != nil || app.gitCM != nil {
		app.registerGit(root, configureDiagnostics)
	}
	if app.zip != nil {
		app.registerZIP(root, configureDiagnostics)
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
	if app.discovery != nil {
		root.SetHelpFunc(func(command *cobra.Command, _ []string) {
			app.discovery.PresentDiscovery(command.Context(), newDiscoveryDocument(command))
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
		LookupEnv:    app.environmentLookup,
	})
	if err != nil {
		return err
	}
	app.logging.SetLevel(configuration.Level)
	app.logging.SetFormat(configuration.Format)
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

var errHelpRequested = errors.New("help requested")
