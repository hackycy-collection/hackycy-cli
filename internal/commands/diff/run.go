package diff

import (
	"context"
	"errors"
)

// Presenter renders the one startup presentation after the server binds.
type Presenter interface {
	Present(Startup) error
}

// Dependencies are the external facts and terminal Adapter required by Diff.
type Dependencies struct {
	NetworkInterfaces func() ([]NetworkInterface, error)
	Presenter         Presenter
}

// Module owns Diff command lifecycle behind its typed command interface.
type Module struct {
	networkInterfaces func() ([]NetworkInterface, error)
	presenter         Presenter
}

// New constructs an unregistered Diff command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.NetworkInterfaces == nil {
		return nil, errors.New("diff network interface provider is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("diff presenter is required")
	}
	return &Module{
		networkInterfaces: dependencies.NetworkInterfaces,
		presenter:         dependencies.Presenter,
	}, nil
}

// Run starts a fixed comparison, presents its usable URLs, and waits for the
// owning process context without assigning an exit code itself.
func (module *Module) Run(ctx context.Context, input Input) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return Result{}, nil
	}
	session, err := startComparison(input)
	if err != nil {
		return Result{}, err
	}
	interfaces, err := module.networkInterfaces()
	if err != nil {
		_ = session.server.Close()
		return Result{}, err
	}
	if ctx.Err() != nil {
		return Result{}, session.wait(ctx)
	}
	if err := module.presenter.Present(session.startupPresentation(interfaces)); err != nil {
		_ = session.server.Close()
		return Result{}, err
	}
	if err := session.wait(ctx); err != nil {
		return Result{}, err
	}
	return Result{}, nil
}
