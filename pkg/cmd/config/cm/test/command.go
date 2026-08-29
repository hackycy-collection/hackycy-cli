package test

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type StoreProvider func() (TestProfileResolver, error)

type Options struct {
	Context  context.Context
	Profile  string
	Store    StoreProvider
	HTTP     *http.Client
	Terminal *terminalexperience.Runtime
}

func NewCmdTest(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runTest
	}
	return &cobra.Command{Use: "test [profile]", Short: "Test a commit message provider profile", Args: cobra.MaximumNArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		if factory == nil || factory.ConfigStore == nil || factory.HTTPClient == nil || factory.Terminal == nil {
			return errors.New("config cm test Factory is incomplete")
		}
		profile := ""
		if len(arguments) == 1 {
			profile = arguments[0]
		}
		return runF(&Options{Context: command.Context(), Profile: profile, Store: func() (TestProfileResolver, error) {
			store, err := factory.ConfigStore()
			if err != nil {
				return nil, err
			}
			return store, nil
		}, HTTP: factory.HTTPClient, Terminal: factory.Terminal})
	}}
}

func runTest(options *Options) error {
	if options == nil || options.Store == nil || options.HTTP == nil || options.Terminal == nil {
		return errors.New("config cm test options are incomplete")
	}
	resolver, err := options.Store()
	if err != nil {
		return err
	}
	module, err := NewTest(TestDependencies{Resolver: resolver, Transport: options.HTTP})
	if err != nil {
		return err
	}
	result, runErr := module.Run(options.Context, TestRequest{Profile: options.Profile})
	if result.Content != "" || result.Diagnostic != nil {
		run := options.Terminal.Open(options.Context)
		defer run.Close()
		if err := run.Present(terminalCMTestDocument(options.Terminal.Session(), result)); err != nil {
			return err
		}
	}
	return runErr
}

func terminalCMTestDocument(session terminalexperience.Session, result TestResult) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		if result.Diagnostic != nil {
			return terminalCMTestFailureDocument(*result.Diagnostic, terminalexperience.VisualRolePlain)
		}
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRolePlain, Text: fmt.Sprintf("Response: %s\nDone", result.Content)}}}
	}
	if result.Diagnostic != nil {
		document := terminalCMTestFailureDocument(*result.Diagnostic, terminalexperience.VisualRoleMuted)
		document.Blocks = append([]terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "Commit message provider test"}, {Role: terminalexperience.VisualRoleWarning, Text: "Provider request failed"}}, document.Blocks...)
		return document
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: terminalexperience.VisualRoleTitle, Text: "Commit message provider test"}, {Role: terminalexperience.VisualRolePlain, Text: "Response:\n" + result.Content}, {Role: terminalexperience.VisualRoleSuccess, Text: "Done"}}}
}

func terminalCMTestFailureDocument(diagnostic TestDiagnostic, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: fmt.Sprintf("Provider: %s\nBase URL: %s\nModel: %s", diagnostic.Provider, diagnostic.BaseURL, diagnostic.Model)}}}
}

var _ TestProfileResolver = (*appconfig.Store)(nil)
