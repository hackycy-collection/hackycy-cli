package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
)

type terminalCMTestPresenter struct {
	output io.Writer
}

func (presenter terminalCMTestPresenter) Response(content string) {
	_, _ = fmt.Fprintf(presenter.output, "Response: %s\nDone\n", content)
}

func (presenter terminalCMTestPresenter) Failure(diagnostic configcm.TestDiagnostic) {
	_, _ = fmt.Fprintf(presenter.output, "Provider: %s\nBase URL: %s\nModel: %s\n", diagnostic.Provider, diagnostic.BaseURL, diagnostic.Model)
}

func newConfigCMTestHandler(output io.Writer) cliapp.ConfigCMTestHandler {
	return func(context context.Context, request configcm.TestRequest) (configcm.TestResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.TestResult{}, err
		}
		module, err := configcm.NewTest(configcm.TestDependencies{
			Resolver:  store,
			Transport: http.DefaultClient,
			Presenter: terminalCMTestPresenter{output: output},
		})
		if err != nil {
			return configcm.TestResult{}, err
		}
		return module.Run(context, request)
	}
}
