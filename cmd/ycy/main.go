package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hackycy/hackycy-cli/internal/fsthumbnail"
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

	if fsthumbnail.IsThumbnailWorkerInvocation(arguments) {
		if err := fsthumbnail.RunThumbnailWorker(os.Stdin, os.Stdout); err != nil {
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
	app, err := rootcommand.New(commandFactory)
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
