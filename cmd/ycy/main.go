package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	fscommand "github.com/hackycy/hackycy-cli/internal/commands/fs"
	zipcommand "github.com/hackycy/hackycy-cli/internal/commands/zip"
	"github.com/hackycy/hackycy-cli/internal/logging"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/web"
)

var version = "0.0.0-dev"

func main() {
	arguments := os.Args[1:]
	if handled, err := runHiddenUpgrade(arguments); handled {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	if fscommand.IsThumbnailWorkerInvocation(arguments) {
		if err := fscommand.RunThumbnailWorker(os.Stdin, os.Stdout); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "thumbnail worker error: %s\n", err)
			os.Exit(1)
		}
		return
	}
	if err := consumeUpgradeStartup(arguments, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := webassets.Validate(); err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	ctx, stop := newYcySignalContext(context.Background())
	defer stop()
	terminalRoot := newRootTerminal(os.Stdin, os.Stdout, os.Stderr, os.LookupEnv, terminal)
	normalDiagnostics := terminalRoot.experience.DiagnosticWriter()
	runtime := newRootLoggingRuntime(terminalRoot)
	var discovery cliapp.DiscoveryPresenter
	if terminalRoot.experience.Session().Kind == terminalexperience.RichInteractive {
		discovery = newTerminalDiscoveryPresenter(terminalRoot.experience)
	}
	exportEnv, err := newExportEnvModule(terminalRoot.input, terminalRoot.output)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	rmModule, err := newRMModule(terminalRoot.input, terminalRoot.output)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	zipModule, err := newZipModule(terminalRoot.input, terminalRoot.output)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	runModule, err := newRunModule(terminalRoot.input, terminalRoot.output, terminalRoot.diagnostics)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	diffModule, err := newDiffModule(terminalRoot.output)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	fsModule, err := newFSModule(terminalRoot.output)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	gitHeatModule, err := newGitHeatModule(terminalRoot.output, terminal(terminalRoot.output) && os.Getenv("NO_COLOR") == "")
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	gitPulseModule, err := newGitPulseModule(terminalRoot.input, terminalRoot.output)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	app, err := cliapp.New(cliapp.BuildInfo{Version: version}, cliapp.Dependencies{
		Out:              terminalRoot.output,
		Err:              normalDiagnostics,
		Logging:          runtime,
		Discovery:        discovery,
		ExportEnv:        exportEnv.Run,
		ConfigForkList:   newConfigForkListHandler(terminalRoot.output),
		ConfigForkAdd:    newConfigForkAddHandler(terminalRoot.input, terminalRoot.output),
		ConfigForkRemove: newConfigForkRemoveHandler(terminalRoot.input, terminalRoot.output),
		ConfigCMList:     newConfigCMListHandler(terminalRoot.output),
		ConfigCMAdd:      newConfigCMAddHandler(terminalRoot.input, terminalRoot.output),
		ConfigCMUse:      newConfigCMUseHandler(terminalRoot.output),
		ConfigCMSet:      newConfigCMSetHandler(terminalRoot.output),
		ConfigCMRemove:   newConfigCMRemoveHandler(terminalRoot.input, terminalRoot.output),
		ConfigCMTest:     newConfigCMTestHandler(terminalRoot.output),
		RM:               rmModule.Run,
		ZIP: func(_ context.Context, input zipcommand.Input) (zipcommand.Result, error) {
			return zipModule.Run(input)
		},
		Run:           runModule.Run,
		Diff:          diffModule.Run,
		FS:            fsModule.Run,
		TunnelServer:  newTunnelServerHandler(runtime.Logger("tunnel.server")),
		TunnelConnect: newTunnelConnectHandler(terminalRoot.input, terminalRoot.output, runtime.Logger("tunnel.client"), version),
		Upgrade:       newUpgradeHandler(terminalRoot.output, normalDiagnostics, version),
		GitHeat:       gitHeatModule.Run,
		GitPulse:      gitPulseModule.Run,
		GitFork:       newGitForkHandler(terminalRoot.input, terminalRoot.output),
		GitCM:         newGitCMHandler(terminalRoot.input, terminalRoot.output),
	})
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	result := app.Execute(ctx, arguments)
	os.Exit(result.Code)
}

func terminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

type rootTerminal struct {
	input       *os.File
	output      *os.File
	diagnostics *os.File
	experience  *terminalexperience.Runtime
}

func newRootTerminal(input, output, diagnostics *os.File, lookup terminalexperience.LookupEnv, isTerminal func(*os.File) bool) rootTerminal {
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Session: terminalexperience.Classify(terminalexperience.Facts{
			Stdin:     terminalexperience.StreamFacts{Terminal: isTerminal(input)},
			Stdout:    terminalexperience.StreamFacts{Terminal: isTerminal(output)},
			Stderr:    terminalexperience.StreamFacts{Terminal: isTerminal(diagnostics)},
			LookupEnv: lookup,
		}),
		Input:       input,
		Output:      output,
		Diagnostics: diagnostics,
	})
	return rootTerminal{
		input:       input,
		output:      output,
		diagnostics: diagnostics,
		experience:  experience,
	}
}

func newRootLoggingRuntime(root rootTerminal) *logging.Runtime {
	session := root.experience.Session()
	return logging.NewRuntime(logging.Options{
		Writer: root.experience.DiagnosticWriter(),
		Color:  session.Kind == terminalexperience.RichInteractive && session.Color,
	})
}
