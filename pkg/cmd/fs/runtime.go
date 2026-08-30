package fs

import (
	"errors"
	"sync"
	"time"
)

// RuntimeOptions selects all command-owned services for one FS process.
// CLI parsing and terminal presentation are intentionally outside this type.
type RuntimeOptions struct {
	Directory          string
	BindingAddress     string
	Port               int
	ManagementEnabled  bool
	SafeHTML           bool
	Accounts           []string
	SessionDirectory   string
	SessionIdleTimeout time.Duration
	ChunkedUploads     bool
	UploadChunkSize    int64
}

// Runtime owns the FS services that share one root-confined workspace.
type Runtime struct {
	workspace         *Workspace
	authentication    *Authentication
	chunkedUploads    *ChunkedUploadManager
	downloads         *DownloadManager
	extractions       *ExtractionManager
	thumbnails        *ThumbnailService
	bindingAddress    string
	port              int
	managementEnabled bool
	safeHTML          bool

	close    sync.Once
	closeErr error
}

// NewRuntime constructs the unregistered FS service graph without opening a
// listener. The returned Runtime must be closed or passed to Start.
func NewRuntime(options RuntimeOptions) (*Runtime, error) {
	workspace, err := OpenWorkspace(options.Directory)
	if err != nil {
		return nil, err
	}
	authentication, err := NewAuthentication(options.Accounts, AuthenticationOptions{
		SessionDirectory: options.SessionDirectory,
		IdleLifetime:     options.SessionIdleTimeout,
	})
	if err != nil {
		_ = workspace.Close()
		return nil, err
	}
	chunkSize := options.UploadChunkSize
	if chunkSize == 0 {
		chunkSize = 8 * 1024 * 1024
	}
	thumbnails, err := NewThumbnailService(workspace)
	if err != nil {
		_ = authentication.Close()
		_ = workspace.Close()
		return nil, err
	}
	return &Runtime{
		workspace:         workspace,
		authentication:    authentication,
		downloads:         NewDownloadManager(workspace),
		extractions:       NewExtractionManager(workspace, ArchiveExtractionOptions{Inspector: NewSevenZipArchiveInspector(), Capacity: defaultArchiveCapacity}),
		thumbnails:        thumbnails,
		bindingAddress:    options.BindingAddress,
		port:              options.Port,
		managementEnabled: options.ManagementEnabled,
		safeHTML:          options.SafeHTML,
		chunkedUploads: func() *ChunkedUploadManager {
			if !options.ChunkedUploads {
				return nil
			}
			return NewChunkedUploadManager(workspace, chunkSize)
		}(),
	}, nil
}

// Start opens the runtime's listener. Server shutdown releases every
// Runtime-owned service exactly once.
func (runtime *Runtime) Start() (*RunningServer, error) {
	if runtime == nil {
		return nil, errors.New("FS runtime is required")
	}
	return StartServer(runtime.workspace, ServerOptions{
		BindingAddress: runtime.bindingAddress,
		Port:           runtime.port,
		ReadOnly: ReadOnlyServerOptions{
			ManagementEnabled: runtime.managementEnabled,
			SafeHTML:          runtime.safeHTML,
			Authentication:    runtime.authentication,
			ChunkedUploads:    runtime.chunkedUploads,
			Downloads:         runtime.downloads,
			Extractions:       runtime.extractions,
			Thumbnails:        runtime.thumbnails,
			BindingAddress:    runtime.bindingAddress,
		},
		Release: runtime.Close,
	})
}

// Close stops all owned services before releasing the Browse Root handle.
func (runtime *Runtime) Close() error {
	if runtime == nil {
		return nil
	}
	runtime.close.Do(func() {
		var result error
		if runtime.chunkedUploads != nil {
			result = errors.Join(result, runtime.chunkedUploads.Close())
		}
		if runtime.downloads != nil {
			runtime.downloads.Close()
		}
		if runtime.extractions != nil {
			runtime.extractions.Close()
		}
		if runtime.thumbnails != nil {
			runtime.thumbnails.Close()
		}
		if runtime.authentication != nil {
			result = errors.Join(result, runtime.authentication.Close())
		}
		if runtime.workspace != nil {
			result = errors.Join(result, runtime.workspace.Close())
		}
		runtime.closeErr = result
	})
	return runtime.closeErr
}
