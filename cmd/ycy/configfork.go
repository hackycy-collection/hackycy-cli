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

func newConfigForkAddHandler(input io.Reader, output io.Writer) cliapp.ConfigForkAddHandler {
	return func(context context.Context, request configfork.AddRequest) (configfork.AddResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configfork.AddResult{}, err
		}
		module, err := configfork.NewAdd(configfork.AddDependencies{
			Prompter:  newTerminalForkAddPrompter(input, output),
			Writer:    store,
			Presenter: terminalForkAddPresenter{output: output},
		})
		if err != nil {
			return configfork.AddResult{}, err
		}
		return module.Run(context, request)
	}
}

func newConfigForkRemoveHandler(input io.Reader, output io.Writer) cliapp.ConfigForkRemoveHandler {
	return func(context context.Context, request configfork.RemoveRequest) (configfork.RemoveResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configfork.RemoveResult{}, err
		}
		module, err := configfork.NewRemove(configfork.RemoveDependencies{
			Reader:    store,
			Prompter:  newTerminalForkRemovePrompter(input, output),
			Writer:    store,
			Presenter: terminalForkRemovePresenter{output: output},
		})
		if err != nil {
			return configfork.RemoveResult{}, err
		}
		return module.Run(context, request)
	}
}
