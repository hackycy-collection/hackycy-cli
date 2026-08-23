package main

import (
	"context"
	"io"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
)

func newConfigForkListHandler(output io.Writer) cliapp.ConfigForkListHandler {
	return func(context context.Context, input configfork.Input) (configfork.Result, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configfork.Result{}, err
		}
		module, err := configfork.New(configfork.Dependencies{Reader: store, Output: output})
		if err != nil {
			return configfork.Result{}, err
		}
		return module.Run(context, input)
	}
}
