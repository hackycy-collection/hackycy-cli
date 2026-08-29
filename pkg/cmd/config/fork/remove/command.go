package remove

import (
	"context"
	"errors"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// StoreProvider resolves the shared appconfig store only when removal runs.
type StoreProvider func() (RemoveReader, RemoveWriter, error)

// Options contains the parsed remove request and leaf-owned capabilities.
type Options struct {
	Context  context.Context
	Store    StoreProvider
	Terminal *terminal.Runtime
}

// NewCmdRemove creates the config fork remove command with an optional test runner.
func NewCmdRemove(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runRemove
	}
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove a provider instance",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
				return errors.New("config fork remove Factory is incomplete")
			}
			return runF(&Options{
				Context: command.Context(),
				Store: func() (RemoveReader, RemoveWriter, error) {
					store, err := factory.ConfigStore()
					if err != nil {
						return nil, nil, err
					}
					return store, store, nil
				},
				Terminal: factory.Terminal,
			})
		},
	}
}

func runRemove(options *Options) error {
	_, err := executeRemove(options)
	return err
}

func executeRemove(options *Options) (RemoveResult, error) {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return RemoveResult{}, errors.New("config fork remove options are incomplete")
	}
	reader, writer, err := options.Store()
	if err != nil {
		return RemoveResult{}, err
	}
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	adapter := newTerminalForkRemoveAdapter(run, options.Terminal.Session())
	module, err := NewRemove(RemoveDependencies{
		Reader:    reader,
		Prompter:  adapter,
		Writer:    writer,
		Presenter: adapter,
	})
	if err != nil {
		return RemoveResult{}, err
	}
	return module.Run(options.Context, RemoveRequest{})
}

var _ RemoveReader = (*appconfig.Store)(nil)
var _ RemoveWriter = (*appconfig.Store)(nil)
