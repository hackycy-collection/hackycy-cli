package list

import (
	"context"
	"errors"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// StoreProvider resolves the appconfig boundary only when the command runs.
// Keeping it as a function preserves the Factory's lazy configuration contract.
type StoreProvider func() (Reader, error)

// Options contains the parsed list request and leaf-owned capabilities.
type Options struct {
	Context  context.Context
	Store    StoreProvider
	Terminal *terminal.Runtime
}

// NewCmdList creates the config fork list command with an optional test runner.
func NewCmdList(factory *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	if runF == nil {
		runF = runList
	}
	command := &cobra.Command{
		Use:   "list",
		Short: "List configured provider instances",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if factory == nil || factory.ConfigStore == nil || factory.Terminal == nil {
				return errors.New("config fork list Factory is incomplete")
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
	return command
}

func runList(options *Options) error {
	if options == nil || options.Store == nil || options.Terminal == nil {
		return errors.New("config fork list options are incomplete")
	}
	store, err := options.Store()
	if err != nil {
		return err
	}
	module, err := New(Dependencies{Reader: store})
	if err != nil {
		return err
	}
	result, err := module.Run(options.Context, Input{})
	if err != nil {
		return err
	}
	run := options.Terminal.Open(options.Context)
	defer run.Close()
	return run.Result(terminalForkListDocument(result))
}

// Ensure the Factory's concrete Store remains the intended implementation of
// the narrow list Reader boundary.
var _ Reader = (*appconfig.Store)(nil)
