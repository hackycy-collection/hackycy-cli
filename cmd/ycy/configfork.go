package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	rootcommand "github.com/hackycy/hackycy-cli/pkg/cmd/root"
)

func newConfigForkListHandler(experience *terminalexperience.Runtime) rootcommand.ConfigForkListHandler {
	return func(context context.Context, input configfork.Input) (configfork.Result, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configfork.Result{}, err
		}
		module, err := configfork.New(configfork.Dependencies{Reader: store})
		if err != nil {
			return configfork.Result{}, err
		}
		result, err := module.Run(context, input)
		if err != nil {
			return configfork.Result{}, err
		}
		run := experience.Open(context)
		defer run.Close()
		if err := run.Present(terminalForkListDocument(experience.Session(), result)); err != nil {
			return configfork.Result{}, err
		}
		return result, nil
	}
}

const terminalForkListEmptyMessage = "No instances configured. Run \"ycy config fork add\" to add one."

func terminalForkListDocument(session terminalexperience.Session, result configfork.Result) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRolePlain,
			Text: terminalForkListPlainText(result.Instances),
		}}}
	}

	blocks := []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleTitle,
		Text: "Fork provider instances",
	}}
	if len(result.Instances) == 0 {
		return terminalexperience.PresentationDocument{Blocks: append(blocks,
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleWarning, Text: "No instances configured."},
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Run \"ycy config fork add\" to add one."},
		)}
	}
	blocks = append(blocks,
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "NAME  TYPE  SCHEME  HOST  TOKEN"},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRolePlain, Text: terminalForkListRows(result.Instances)},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleSuccess, Text: terminalForkListCount(len(result.Instances))},
	)
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func terminalForkListPlainText(instances []configfork.Instance) string {
	if len(instances) == 0 {
		return terminalForkListEmptyMessage + "\n"
	}

	var output bytes.Buffer
	table := tabwriter.NewWriter(&output, 0, 8, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, "NAME\tTYPE\tSCHEME\tHOST\tTOKEN")
	for _, instance := range instances {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", instance.Name, instance.Type, instance.Scheme, instance.Host, instance.TokenPreview)
	}
	_ = table.Flush()
	_, _ = fmt.Fprintln(&output, terminalForkListCount(len(instances)))
	return output.String()
}

func terminalForkListRows(instances []configfork.Instance) string {
	lines := make([]string, 0, len(instances))
	for _, instance := range instances {
		lines = append(lines, fmt.Sprintf("%s  %s  %s  %s  %s", instance.Name, instance.Type, instance.Scheme, instance.Host, instance.TokenPreview))
	}
	return strings.Join(lines, "\n")
}

func terminalForkListCount(count int) string {
	label := "instances"
	if count == 1 {
		label = "instance"
	}
	return fmt.Sprintf("%d %s configured", count, label)
}

func newConfigForkAddHandler(experience *terminalexperience.Runtime) rootcommand.ConfigForkAddHandler {
	return func(context context.Context, request configfork.AddRequest) (configfork.AddResult, error) {
		if experience.Session().Kind == terminalexperience.Automation {
			return configfork.AddResult{}, errConfigForkAddRequiresInteractive
		}
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configfork.AddResult{}, err
		}
		run := experience.Open(context)
		defer run.Close()
		adapter := newTerminalForkAddAdapter(run, experience.Session())
		module, err := configfork.NewAdd(configfork.AddDependencies{
			Prompter:  adapter,
			Writer:    store,
			Presenter: adapter,
		})
		if err != nil {
			return configfork.AddResult{}, err
		}
		return module.Run(context, request)
	}
}

func newConfigForkRemoveHandler(experience *terminalexperience.Runtime) rootcommand.ConfigForkRemoveHandler {
	return func(context context.Context, request configfork.RemoveRequest) (configfork.RemoveResult, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return configfork.RemoveResult{}, err
		}
		if experience.Session().Kind == terminalexperience.Automation {
			instances, err := store.ListForkInstances()
			if err != nil {
				return configfork.RemoveResult{}, err
			}
			if len(instances) > 0 {
				return configfork.RemoveResult{}, errConfigForkRemoveRequiresInteractive
			}
		}
		run := experience.Open(context)
		defer run.Close()
		adapter := newTerminalForkRemoveAdapter(run, experience.Session())
		module, err := configfork.NewRemove(configfork.RemoveDependencies{
			Reader:    store,
			Prompter:  adapter,
			Writer:    store,
			Presenter: adapter,
		})
		if err != nil {
			return configfork.RemoveResult{}, err
		}
		return module.Run(context, request)
	}
}
