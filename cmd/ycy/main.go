package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	"github.com/hackycy/hackycy-cli/internal/logging"
	"github.com/hackycy/hackycy-cli/web"
)

var version = "0.0.0-dev"

func main() {
	if err := webassets.Validate(); err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	ctx, stop := newYcySignalContext(context.Background())
	defer stop()
	runtime := logging.NewRuntime(logging.Options{Writer: os.Stderr, Color: terminal(os.Stderr)})
	exportEnv, err := newExportEnvModule(os.Stdin, os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	rmModule, err := newRMModule(os.Stdin, os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	runModule, err := newRunModule(os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	gitHeatModule, err := newGitHeatModule(os.Stdout, terminal(os.Stdout) && os.Getenv("NO_COLOR") == "")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	gitPulseModule, err := newGitPulseModule(os.Stdin, os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	app, err := cliapp.New(cliapp.BuildInfo{Version: version}, cliapp.Dependencies{
		Logging:          runtime,
		ExportEnv:        exportEnv.Run,
		ConfigForkList:   newConfigForkListHandler(os.Stdout),
		ConfigForkAdd:    newConfigForkAddHandler(os.Stdin, os.Stdout),
		ConfigForkRemove: newConfigForkRemoveHandler(os.Stdin, os.Stdout),
		ConfigCMList:     newConfigCMListHandler(os.Stdout),
		ConfigCMAdd:      newConfigCMAddHandler(os.Stdin, os.Stdout),
		ConfigCMUse:      newConfigCMUseHandler(os.Stdout),
		ConfigCMSet:      newConfigCMSetHandler(os.Stdout),
		ConfigCMRemove:   newConfigCMRemoveHandler(os.Stdin, os.Stdout),
		ConfigCMTest:     newConfigCMTestHandler(os.Stdout),
		RM:               rmModule.Run,
		Run:              runModule.Run,
		GitHeat:          gitHeatModule.Run,
		GitPulse:         gitPulseModule.Run,
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	result := app.Execute(ctx, os.Args[1:])
	os.Exit(result.Code)
}

func terminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
