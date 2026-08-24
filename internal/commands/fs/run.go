package fs

import (
	"context"
	"errors"
)

// Presenter renders the server facts and normal-stop outcome for one FS run.
type Presenter interface {
	Present(Startup) error
	Stopped() error
}

// Dependencies are the host facts and terminal adapter required by FS.
type Dependencies struct {
	NetworkInterfaces func() ([]NetworkInterface, error)
	Presenter         Presenter
}

// Module owns the unregistered FS command lifecycle behind its typed API.
type Module struct {
	networkInterfaces func() ([]NetworkInterface, error)
	presenter         Presenter
}

// New constructs an unregistered FS command module.
func New(dependencies Dependencies) (*Module, error) {
	if dependencies.NetworkInterfaces == nil {
		return nil, errors.New("FS network interface provider is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("FS presenter is required")
	}
	return &Module{networkInterfaces: dependencies.NetworkInterfaces, presenter: dependencies.Presenter}, nil
}

// Run creates the FS runtime, presents its concrete URLs, and waits for the
// owning process context without assigning an exit code itself.
func (module *Module) Run(ctx context.Context, input Input) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return Result{}, nil
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
		return Result{}, err
	}
	server, err := runtime.Start()
	if err != nil {
		_ = runtime.Close()
		return Result{}, err
	}
	interfaces, err := module.networkInterfaces()
	if err != nil {
		_ = server.Close()
		return Result{}, err
	}
	startup := runtimeStartup(runtime, interfaces)
	startup.Port = server.Port()
	startup.URLs = makeFSStartupURLs(runtime.bindingAddress, startup.Port, interfaces)
	if err := module.presenter.Present(startup); err != nil {
		_ = server.Close()
		return Result{}, err
	}
	select {
	case <-ctx.Done():
		if err := server.Close(); err != nil {
			return Result{}, err
		}
		if err := module.presenter.Stopped(); err != nil {
			return Result{}, err
		}
		return Result{}, nil
	case <-server.done:
		if err := server.Wait(); err != nil {
			return Result{}, err
		}
		if err := module.presenter.Stopped(); err != nil {
			return Result{}, err
		}
		return Result{}, nil
	}
}
