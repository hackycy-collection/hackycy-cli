package main

import (
	"context"
	"fmt"
	"io"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
)

type terminalCMPresenter struct {
	output io.Writer
}

func (presenter terminalCMPresenter) Cancel(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

func (presenter terminalCMPresenter) Success(message string) {
	_, _ = fmt.Fprintln(presenter.output, message)
}

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
			Presenter: terminalCMPresenter{output: output},
		})
		if err != nil {
			return configcm.AddResult{}, err
		}
		return module.Run(context, request)
	}
}

func newConfigCMUseHandler(output io.Writer) cliapp.ConfigCMUseHandler {
	return func(context context.Context, request configcm.UseRequest) (configcm.UseResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.UseResult{}, err
		}
		module, err := configcm.NewUse(configcm.UseDependencies{
			Writer:    store,
			Presenter: terminalCMPresenter{output: output},
		})
		if err != nil {
			return configcm.UseResult{}, err
		}
		return module.Run(context, request)
	}
}

func newConfigCMSetHandler(output io.Writer) cliapp.ConfigCMSetHandler {
	return func(context context.Context, request configcm.SetRequest) (configcm.SetResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.SetResult{}, err
		}
		module, err := configcm.NewSet(configcm.SetDependencies{
			Writer:    store,
			Presenter: terminalCMPresenter{output: output},
		})
		if err != nil {
			return configcm.SetResult{}, err
		}
		return module.Run(context, request)
	}
}

func newConfigCMRemoveHandler(input io.Reader, output io.Writer) cliapp.ConfigCMRemoveHandler {
	return func(context context.Context, request configcm.RemoveRequest) (configcm.RemoveResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.RemoveResult{}, err
		}
		module, err := configcm.NewRemove(configcm.RemoveDependencies{
			Prompter:  newTerminalCMRemovePrompter(input, output),
			Writer:    store,
			Presenter: terminalCMPresenter{output: output},
		})
		if err != nil {
			return configcm.RemoveResult{}, err
		}
		return module.Run(context, request)
	}
}
