package use

import (
	"context"
	"errors"
	"fmt"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type StoreProvider func() (UseWriter, error)

type Options struct {
	Context  context.Context
	Profile  string
	Store    StoreProvider
	Terminal *terminalexperience.Runtime
}

func NewCmdUse(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runUse
	}
	return &cobra.Command{Use: "use <profile>", Short: "Set the default commit message provider profile", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
			return errors.New("config cm use Factory is incomplete")
		}
		return runF(&Options{Context: command.Context(), Profile: arguments[0], Store: func() (UseWriter, error) {
			store, err := factory.ConfigStore()
			if err != nil {
				return nil, err
			}
			return store, nil
		}, Terminal: factory.Terminal})
	}}
}

// profile is held only by the constructor closure; Options remains the leaf's semantic boundary.
// The command runner receives it through the private field below.
func runUse(options *Options) error { return executeUse(options) }

func executeUse(options *Options) error {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return errors.New("config cm use options are incomplete")
	}
	writer, err := options.Store()
	if err != nil {
		return err
	}
	module, err := NewUse(UseDependencies{Writer: writer})
	if err != nil {
		return err
	}
	result, err := module.Run(options.Context, UseRequest{Profile: options.Profile})
	if err != nil {
		return err
	}
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	return run.Present(terminalCMUseDocument(options.Terminal.Session(), result))
}

func terminalCMUseDocument(session terminalexperience.Session, result UseResult) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleSuccess
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: fmt.Sprintf("Default CM profile set to %s", result.Profile)}}}
}

var _ UseWriter = (*appconfig.Store)(nil)
