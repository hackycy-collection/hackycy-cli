package fs

import (
	"context"
	"errors"
)

// Dependencies are the host facts required by FS.
type Dependencies struct {
	NetworkInterfaces func() ([]NetworkInterface, error)
}

// Module owns the unregistered FS command lifecycle behind its typed API.
type Module struct {
	networkInterfaces func() ([]NetworkInterface, error)
}

// New constructs an unregistered FS command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.NetworkInterfaces == nil {
		return nil, errors.New("FS network interface provider is required")
	}
	return &Module{networkInterfaces: dependencies.NetworkInterfaces}, nil
}

// BrowserOperation holds a started file browser open until its caller waits or closes it.
type BrowserOperation struct {
	Startup Startup
	server  *RunningServer
}

// Start binds the file browser and returns its safe startup facts without
// assigning terminal presentation or process exit behavior.
func (module *Module) Start(ctx context.Context, input Input) (*BrowserOperation, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return nil, nil
	}
	runtime, err := NewRuntime(RuntimeOptions{
		Directory:          input.Directory,
		BindingAddress:     input.Address,
		Port:               input.Port,
		ManagementEnabled:  input.ManagementEnabled,
		SafeHTML:           input.SafeHTML,
		Accounts:           append([]string(nil), input.Accounts...),
		SessionDirectory:   input.SessionDirectory,
		SessionIdleTimeout: input.SessionIdleTimeout,
		ChunkedUploads:     input.ChunkedUploads,
		UploadChunkSize:    input.UploadChunkSize,
	})
	if err != nil {
		return nil, err
	}
	server, err := runtime.Start()
	if err != nil {
		_ = runtime.Close()
		return nil, err
	}
	interfaces, err := module.networkInterfaces()
	if err != nil {
		_ = server.Close()
		return nil, err
	}
	startup := runtimeStartup(runtime, interfaces)
	startup.Port = server.Port()
	startup.URLs = makeFSStartupURLs(runtime.bindingAddress, startup.Port, interfaces)
	operation := &BrowserOperation{Startup: startup, server: server}
	if ctx.Err() != nil {
		if err := operation.Wait(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return operation, nil
}

// Wait keeps the foreground file browser alive until its context ends or the
// server stops on its own.
func (operation *BrowserOperation) Wait(ctx context.Context) error {
	if operation == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return operation.server.Close()
	case <-operation.server.done:
		return operation.server.Wait()
	}
}

// Close stops the foreground file browser when presentation cannot continue.
func (operation *BrowserOperation) Close() error {
	if operation == nil {
		return nil
	}
	return operation.server.Close()
}
