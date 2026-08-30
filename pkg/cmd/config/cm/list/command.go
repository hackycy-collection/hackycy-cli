package list

import (
	"context"
	"errors"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type StoreProvider func() (Reader, error)

type Options struct {
	Context  context.Context
	Store    StoreProvider
	Terminal *terminal.Runtime
}

func NewCmdList(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runList
	}
	return &cobra.Command{
		Use:   "list",
		Short: "List configured CM profiles",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
				return errors.New("config cm list Factory is incomplete")
			}
			return runF(&Options{
				Context: command.Context(),
				Store: func() (Reader, error) {
					store, err := factory.ConfigStore()
					if err != nil {
						return nil, err
					}
					return store, nil
				},
				Terminal: factory.Terminal,
			})
		},
	}
}

func runList(options *Options) error {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return errors.New("config cm list options are incomplete")
	}
	reader, err := options.Store()
	if err != nil {
		return err
	}
	module, err := New(Dependencies{Reader: reader})
	if err != nil {
		return err
	}
	result, err := module.Run(options.Context, Input{})
	if err != nil {
		return err
	}
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	return run.Result(terminalCMListDocument(result))
}

var _ Reader = (*appconfig.Store)(nil)
