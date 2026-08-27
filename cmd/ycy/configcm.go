package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	configcm "github.com/hackycy/hackycy-cli/internal/commands/config/cm"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
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

func newConfigCMListHandler(experience *terminalexperience.Runtime) cliapp.ConfigCMListHandler {
	return func(context context.Context, input configcm.Input) (configcm.Result, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.Result{}, err
		}
		module, err := configcm.New(configcm.Dependencies{Reader: store})
		if err != nil {
			return configcm.Result{}, err
		}
		result, err := module.Run(context, input)
		if err != nil {
			return configcm.Result{}, err
		}
		run := experience.Open(context)
		defer run.Close()
		if err := run.Present(terminalCMListDocument(experience.Session(), result)); err != nil {
			return configcm.Result{}, err
		}
		return result, nil
	}
}

const terminalCMListEmptyMessage = "No CM profiles configured. Run \"ycy config cm add\" to add one."

func terminalCMListDocument(session terminalexperience.Session, result configcm.Result) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRolePlain,
			Text: terminalCMListPlainText(result.Profiles),
		}}}
	}

	blocks := []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleTitle,
		Text: "Commit message profiles",
	}}
	if len(result.Profiles) == 0 {
		return terminalexperience.PresentationDocument{Blocks: append(blocks,
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: "No CM profiles configured."},
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Run \"ycy config cm add\" to add one."},
		)}
	}
	blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "PROFILE  MODEL  BASE URL"})
	for _, profile := range result.Profiles {
		role := terminalexperience.VisualRolePlain
		if profile.Default {
			role = terminalexperience.VisualRoleSuccess
		}
		blocks = append(blocks, terminalexperience.PresentationBlock{Role: role, Text: terminalCMListRow(profile)})
	}
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func terminalCMListPlainText(profiles []configcm.Profile) string {
	if len(profiles) == 0 {
		return terminalCMListEmptyMessage + "\n"
	}

	var output bytes.Buffer
	for _, profile := range profiles {
		_, _ = fmt.Fprintln(&output, terminalCMListRow(profile))
	}
	return output.String()
}

func terminalCMListRow(profile configcm.Profile) string {
	marker := " "
	if profile.Default {
		marker = "*"
	}
	return fmt.Sprintf("%s %s %s %s", marker, profile.Name, profile.Model, profile.BaseURL)
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

func newConfigCMUseHandler(experience *terminalexperience.Runtime) cliapp.ConfigCMUseHandler {
	return func(context context.Context, request configcm.UseRequest) (configcm.UseResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.UseResult{}, err
		}
		module, err := configcm.NewUse(configcm.UseDependencies{
			Writer: store,
		})
		if err != nil {
			return configcm.UseResult{}, err
		}
		result, err := module.Run(context, request)
		if err != nil {
			return configcm.UseResult{}, err
		}
		run := experience.Open(context)
		defer run.Close()
		if err := run.Present(terminalCMUseDocument(experience.Session(), result)); err != nil {
			return configcm.UseResult{}, err
		}
		return result, nil
	}
}

func terminalCMUseDocument(session terminalexperience.Session, result configcm.UseResult) terminalexperience.PresentationDocument {
	message := fmt.Sprintf("Default CM profile set to %s", result.Profile)
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleSuccess
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: message}}}
}

func newConfigCMSetHandler(experience *terminalexperience.Runtime) cliapp.ConfigCMSetHandler {
	return func(context context.Context, request configcm.SetRequest) (configcm.SetResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configcm.SetResult{}, err
		}
		module, err := configcm.NewSet(configcm.SetDependencies{
			Writer: store,
		})
		if err != nil {
			return configcm.SetResult{}, err
		}
		result, err := module.Run(context, request)
		if err != nil {
			return configcm.SetResult{}, err
		}
		run := experience.Open(context)
		defer run.Close()
		if err := run.Present(terminalCMSetDocument(experience.Session(), result)); err != nil {
			return configcm.SetResult{}, err
		}
		return result, nil
	}
}

func terminalCMSetDocument(session terminalexperience.Session, result configcm.SetResult) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleSuccess
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
		Role: role,
		Text: fmt.Sprintf("Profile %s updated", result.Profile),
	}}}
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
