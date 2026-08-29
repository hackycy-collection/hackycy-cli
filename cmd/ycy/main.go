package main

import (
	"context"
	"fmt"
	"os"

	fscommand "github.com/hackycy/hackycy-cli/internal/commands/fs"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	commandfactory "github.com/hackycy/hackycy-cli/pkg/cmd/factory"
	rootcommand "github.com/hackycy/hackycy-cli/pkg/cmd/root"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
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
	terminalRoot := newRootTerminal(os.Stdin, os.Stdout, os.Stderr, os.LookupEnv, terminal)
	commandFactory := commandfactory.New(commandfactory.Options{
		Version: version,
		IOStreams: cmdutil.IOStreams{
			In:     terminalRoot.input,
			Out:    terminalRoot.output,
			ErrOut: terminalRoot.diagnostics,
		},
		Session:           terminalRoot.experience.Session(),
		Environment:       os.Getenv,
		EnvironmentLookup: os.LookupEnv,
	})
	terminalRoot.experience = commandFactory.Terminal
	normalDiagnostics := commandFactory.Terminal.DiagnosticWriter()
	if err := consumeUpgradeStartup(arguments, terminalRoot.experience); err != nil {
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}

	if err := webassets.Validate(); err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}

	ctx, stop := newYcySignalContext(context.Background())
	defer stop()
	runtime := commandFactory.Logging
	diffHandler, err := newDiffHandler(terminalRoot.experience)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	fsHandler, err := newFSHandler(terminalRoot.experience)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	gitHeat, err := newGitHeatHandler(terminalRoot.experience)
	if err != nil {
		_, _ = fmt.Fprintln(terminalRoot.output)
		_, _ = fmt.Fprintf(normalDiagnostics, "error: %s\n", err)
		os.Exit(1)
	}
	app, err := rootcommand.New(commandFactory, rootcommand.Dependencies{
		ExportEnv:        newExportEnvHandler(terminalRoot.experience),
		ConfigForkList:   newConfigForkListHandler(terminalRoot.experience),
		ConfigForkAdd:    newConfigForkAddHandler(terminalRoot.experience),
		ConfigForkRemove: newConfigForkRemoveHandler(terminalRoot.experience),
		ConfigCMList:     newConfigCMListHandler(terminalRoot.experience),
		ConfigCMAdd:      newConfigCMAddHandler(terminalRoot.experience),
		ConfigCMUse:      newConfigCMUseHandler(terminalRoot.experience),
		ConfigCMSet:      newConfigCMSetHandler(terminalRoot.experience),
		ConfigCMRemove:   newConfigCMRemoveHandler(terminalRoot.experience),
		ConfigCMTest:     newConfigCMTestHandler(terminalRoot.experience),
		RM:               newRMHandler(terminalRoot.experience),
		ZIP:              newZipHandler(terminalRoot.experience),
		Run:              newRunHandler(terminalRoot.experience, terminalRoot.input, terminalRoot.output, terminalRoot.diagnostics),
		Diff:             diffHandler,
		FS:               fsHandler,
		TunnelServer:     newTunnelServerHandler(runtime.Logger("tunnel.server")),
		TunnelConnect:    newTunnelConnectHandler(terminalRoot.experience, runtime.Logger("tunnel.client"), version),
		Upgrade:          newUpgradeHandler(terminalRoot.experience, version),
		GitHeat:          gitHeat,
		GitPulse:         newGitPulseHandler(terminalRoot.experience),
		GitFork:          newGitForkHandler(terminalRoot.experience),
		GitCM:            newGitCMHandler(terminalRoot.experience),
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
