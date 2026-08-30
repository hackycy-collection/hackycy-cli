package add

import (
	"context"
	"errors"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type StoreProvider func() (AddWriter, error)

type Options struct {
	Context  context.Context
	Store    StoreProvider
	Terminal *terminal.Runtime
}

func NewCmdAdd(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runAdd
	}
	return &cobra.Command{
		Use:   "add",
		Short: "Add a CM profile",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
				return errors.New("config cm add Factory is incomplete")
			}
			return runF(&Options{Context: command.Context(), Store: func() (AddWriter, error) {
				store, err := factory.ConfigStore()
				if err != nil {
					return nil, err
				}
				return store, nil
			}, Terminal: factory.Terminal})
		},
	}
}

func runAdd(options *Options) error {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return errors.New("config cm add options are incomplete")
	}
	if options.Terminal.Capabilities().Interaction == terminal.Automation {
		return errConfigCMAddRequiresInteractive
	}
	writer, err := options.Store()
	if err != nil {
		return err
	}
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	adapter := newTerminalCMAddAdapter(run)
	module, err := NewAdd(AddDependencies{Prompter: adapter, Writer: writer, Presenter: adapter})
	if err != nil {
		return err
	}
	_, err = module.Run(options.Context, AddRequest{})
	return err
}

var errConfigCMAddRequiresInteractive = errors.New("config cm add requires an interactive terminal")
var _ AddWriter = (*appconfig.Store)(nil)
