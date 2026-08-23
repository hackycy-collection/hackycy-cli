package main

import (
	"context"
	"io"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
)

func newConfigCMListHandler(output io.Writer) cliapp.ConfigCMListHandler {
	return func(context context.Context, input configcm.Input) (configcm.Result, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.Result{}, err
		}
		module, err := configcm.New(configcm.Dependencies{Reader: store, Output: output})
		if err != nil {
			return configcm.Result{}, err
		}
		return module.Run(context, input)
	}
}

func newConfigCMAddHandler(input io.Reader, output io.Writer) cliapp.ConfigCMAddHandler {
	return func(context context.Context, request configcm.AddRequest) (configcm.AddResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.AddResult{}, err
		}
		module, err := configcm.NewAdd(configcm.AddDependencies{
			Prompter:  newTerminalCMAddPrompter(input, output),
			Writer:    store,
			Presenter: terminalCMAddPresenter{output: output},
		})
		if err != nil {
			return configcm.AddResult{}, err
		}
		return module.Run(context, request)
	}
}
