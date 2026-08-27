package diff

import (
	"context"
	"errors"
)

// Dependencies are the external host facts required by Diff.
type Dependencies struct {
	NetworkInterfaces func() ([]NetworkInterface, error)
}

// Module owns Diff command lifecycle behind its typed command interface.
type Module struct {
	networkInterfaces func() ([]NetworkInterface, error)
}

// New constructs an unregistered Diff command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.NetworkInterfaces == nil {
		return nil, errors.New("diff network interface provider is required")
	}
	return &Module{
		networkInterfaces: dependencies.NetworkInterfaces,
	}, nil
}

// Operation holds a started comparison open until its caller waits or closes it.
type Operation struct {
	Startup Startup
	session *comparisonSession
}

// Start binds a fixed comparison and returns its safe startup facts without
// assigning terminal presentation or process exit behavior.
func (module *Module) Start(ctx context.Context, input Input) (*Operation, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return nil, nil
	}
	session, err := startComparison(input)
	if err != nil {
		return nil, err
	}
	interfaces, err := module.networkInterfaces()
	if err != nil {
		_ = session.server.Close()
		return nil, err
	}
	operation := &Operation{session: session, Startup: session.startupPresentation(interfaces)}
	if ctx.Err() != nil {
		if err := operation.Wait(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return operation, nil
}

// Wait keeps the foreground server alive until its context ends or the server
// stops on its own.
func (operation *Operation) Wait(ctx context.Context) error {
	if operation == nil {
		return nil
	}
	return operation.session.wait(ctx)
}

// Close stops the foreground server when presentation cannot continue.
func (operation *Operation) Close() error {
	if operation == nil {
		return nil
	}
	return operation.session.server.Close()
}
