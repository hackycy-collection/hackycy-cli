package remove

import (
	"context"
	"errors"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type StoreProvider func() (Reader, RemoveWriter, error)

type Options struct {
	Context  context.Context
	Profile  string
	Store    StoreProvider
	Terminal *terminalexperience.Runtime
}

func NewCmdRemove(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runRemove
	}
	return &cobra.Command{Use: "remove <profile>", Short: "Remove a commit message provider profile", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, arguments []string) error {
		if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
			return errors.New("config cm remove Factory is incomplete")
		}
		return runF(&Options{Context: command.Context(), Profile: arguments[0], Store: func() (Reader, RemoveWriter, error) {
			store, err := factory.ConfigStore()
			if err != nil {
				return nil, nil, err
			}
			return store, store, nil
		}, Terminal: factory.Terminal})
	}}
}

func runRemove(options *Options) error {
	_, err := executeRemove(options)
	return err
}

func executeRemove(options *Options) (RemoveResult, error) {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return RemoveResult{}, errors.New("config cm remove options are incomplete")
	}
	reader, writer, err := options.Store()
	if err != nil {
		return RemoveResult{}, err
	}
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	adapter := newTerminalCMRemoveAdapter(run)
	module, err := NewRemove(RemoveDependencies{Reader: reader, Prompter: adapter, Writer: writer, Presenter: adapter})
	if err != nil {
		return RemoveResult{}, err
	}
	return module.Run(options.Context, RemoveRequest{Profile: options.Profile})
}

var _ Reader = (*appconfig.Store)(nil)
var _ RemoveWriter = (*appconfig.Store)(nil)
