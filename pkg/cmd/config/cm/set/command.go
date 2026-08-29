package set

import (
	"context"
	"errors"
	"fmt"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type StoreProvider func() (SetWriter, error)

type Options struct {
	Context  context.Context
	Profile  string
	Key      string
	Value    string
	Store    StoreProvider
	Terminal *terminalexperience.Runtime
}

func NewCmdSet(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runSet
	}
	return &cobra.Command{Use: "set <profile> <key> <value>", Short: "Set an optional commit message provider profile value", Args: cobra.ExactArgs(3), RunE: func(command *cobra.Command, arguments []string) error {
		if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
			return errors.New("config cm set Factory is incomplete")
		}
		return runF(&Options{Context: command.Context(), Profile: arguments[0], Key: arguments[1], Value: arguments[2], Store: func() (SetWriter, error) {
			store, err := factory.ConfigStore()
			if err != nil {
				return nil, err
			}
			return store, nil
		}, Terminal: factory.Terminal})
	}}
}

func runSet(options *Options) error {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return errors.New("config cm set options are incomplete")
	}
	writer, err := options.Store()
	if err != nil {
		return err
	}
	module, err := NewSet(SetDependencies{Writer: writer})
	if err != nil {
		return err
	}
	result, err := module.Run(options.Context, SetRequest{Profile: options.Profile, Key: options.Key, Value: options.Value})
	if err != nil {
		return err
	}
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	return run.Present(terminalCMSetDocument(options.Terminal.Session(), result))
}

func terminalCMSetDocument(session terminalexperience.Session, result SetResult) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleSuccess
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{Role: role, Text: fmt.Sprintf("Profile %s updated", result.Profile)}}}
}

var _ SetWriter = (*appconfig.Store)(nil)
