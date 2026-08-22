package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

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

	context, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	runtime := logging.NewRuntime(logging.Options{Writer: os.Stderr, Color: terminal(os.Stderr)})
	app, err := cliapp.New(cliapp.BuildInfo{Version: version}, cliapp.Dependencies{Logging: runtime})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout)
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	result := app.Execute(context, os.Args[1:])
	os.Exit(result.Code)
}

func terminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
