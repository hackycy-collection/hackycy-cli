package fs

import (
	"context"
	"errors"
	"net/netip"
	"strconv"
	"sync"
	"testing"
)

func TestModuleRunsUntilItsContextStopsAndPresentsBoundFacts(t *testing.T) {
	context, cancel := context.WithCancel(context.Background())
	defer cancel()
	presenter := &testPresenter{onPresent: cancel}
	module, err := New(Dependencies{
		NetworkInterfaces: func() ([]NetworkInterface, error) {
			return []NetworkInterface{
				{Internal: true, Addresses: []netip.Addr{netip.MustParseAddr("127.0.0.1")}},
				{Addresses: []netip.Addr{netip.MustParseAddr("192.0.2.44"), netip.MustParseAddr("2001:db8::44")}},
			}, nil
		},
		Presenter: presenter,
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	result, err := module.Run(context, Input{
		Directory:         t.TempDir(),
		Address:           "0.0.0.0",
		Port:              0,
		ManagementEnabled: true,
		ChunkedUploads:    true,
		UploadChunkSize:   4 * 1024 * 1024,
	})
	if err != nil || result != (Result{}) {
		t.Fatalf("Run() = (%#v, %v)", result, err)
	}
	startup, stops := presenter.snapshot()
	if stops != 1 || startup.Port == 0 || startup.Directory == "" || !startup.ManagementEnabled || !startup.ChunkedUploads || startup.UploadChunkSize != 4*1024*1024 {
		t.Fatalf("presenter = startup=%#v stops=%d", startup, stops)
	}
	if len(startup.URLs) != 2 || startup.URLs[0] != (StartupURL{Label: "Local", URL: "http://localhost:" + strconv.Itoa(startup.Port)}) || startup.URLs[1] != (StartupURL{Label: "Network", URL: "http://192.0.2.44:" + strconv.Itoa(startup.Port)}) {
		t.Fatalf("startup URLs = %#v", startup.URLs)
	}
}

func TestModuleClosesItsRuntimeWhenStartupPresentationFails(t *testing.T) {
	presentationErr := errors.New("presenter failed")
	module, err := New(Dependencies{
		NetworkInterfaces: func() ([]NetworkInterface, error) { return nil, nil },
		Presenter:         &testPresenter{presentErr: presentationErr},
	})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	_, err = module.Run(context.Background(), Input{Directory: t.TempDir(), Address: "127.0.0.1", Port: 0})
	if !errors.Is(err, presentationErr) {
		t.Fatalf("Run error = %v, want presentation error", err)
	}
}

func TestModuleRequiresItsHostDependencies(t *testing.T) {
	if module, err := New(Dependencies{}); err == nil || module != nil {
		t.Fatalf("New(empty) = (%v, %v), want nil module and error", module, err)
	}
	if module, err := New(Dependencies{NetworkInterfaces: func() ([]NetworkInterface, error) { return nil, nil }}); err == nil || module != nil {
		t.Fatalf("New(no presenter) = (%v, %v), want nil module and error", module, err)
	}
}

type testPresenter struct {
	mu         sync.Mutex
	startup    Startup
	presentErr error
	stops      int
	onPresent  func()
}

func (presenter *testPresenter) Present(startup Startup) error {
	presenter.mu.Lock()
	presenter.startup = startup
	presenter.mu.Unlock()
	if presenter.onPresent != nil {
		presenter.onPresent()
	}
	return presenter.presentErr
}

func (presenter *testPresenter) Stopped() error {
	presenter.mu.Lock()
	defer presenter.mu.Unlock()
	presenter.stops++
	return nil
}

func (presenter *testPresenter) snapshot() (Startup, int) {
	presenter.mu.Lock()
	defer presenter.mu.Unlock()
	return presenter.startup, presenter.stops
}
