package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

func newConfigCMTestHandler(experience *terminalexperience.Runtime) cliapp.ConfigCMTestHandler {
	return func(context context.Context, request configcm.TestRequest) (configcm.TestResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.TestResult{}, err
		}
		module, err := configcm.NewTest(configcm.TestDependencies{
			Resolver:  store,
			Transport: http.DefaultClient,
		})
		if err != nil {
			return configcm.TestResult{}, err
		}
		result, runErr := module.Run(context, request)
		if result.Content != "" || result.Diagnostic != nil {
			run := experience.Open(context)
			defer run.Close()
			if err := run.Present(terminalCMTestDocument(experience.Session(), result)); err != nil {
				return configcm.TestResult{}, err
			}
		}
		if runErr != nil {
			return result, runErr
		}
		return result, nil
	}
}

func terminalCMTestDocument(session terminalexperience.Session, result configcm.TestResult) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		if result.Diagnostic != nil {
			return terminalCMTestFailureDocument(*result.Diagnostic, terminalexperience.VisualRolePlain)
		}
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRolePlain,
			Text: fmt.Sprintf("Response: %s\nDone", result.Content),
		}}}
	}

	if result.Diagnostic != nil {
		document := terminalCMTestFailureDocument(*result.Diagnostic, terminalexperience.VisualRoleMuted)
		document.Blocks = append([]terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRoleTitle,
			Text: "Commit message provider test",
		}, {
			Role: terminalexperience.VisualRoleWarning,
			Text: "Provider request failed",
		}}, document.Blocks...)
		return document
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleTitle,
		Text: "Commit message provider test",
	}, {
		Role: terminalexperience.VisualRolePlain,
		Text: "Response:\n" + result.Content,
	}, {
		Role: terminalexperience.VisualRoleSuccess,
		Text: "Done",
	}}}
}

func terminalCMTestFailureDocument(diagnostic configcm.TestDiagnostic, role terminalexperience.VisualRole) terminalexperience.PresentationDocument {
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
		Role: role,
		Text: fmt.Sprintf("Provider: %s\nBase URL: %s\nModel: %s", diagnostic.Provider, diagnostic.BaseURL, diagnostic.Model),
	}}}
}
