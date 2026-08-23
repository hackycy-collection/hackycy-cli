package diff

import (
	"context"
	"net/netip"
	"path/filepath"
	"strconv"
	"testing"
)

func TestModulePresentsBoundComparisonThenTreatsContextCancellationAsSuccess(t *testing.T) {
	baseline, target := comparisonRoots(t)
	resolvedBaseline, err := filepath.EvalSymlinks(baseline)
	if err != nil {
		t.Fatalf("resolve baseline: %v", err)
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatalf("resolve target: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	presenter := &recordingDiffPresenter{afterPresent: cancel}
	module, err := New(Dependencies{
		NetworkInterfaces: func() ([]NetworkInterface, error) {
			return []NetworkInterface{{Addresses: []netip.Addr{netip.MustParseAddr("192.168.1.50")}}}, nil
		},
		Presenter: presenter,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := module.Run(ctx, Input{BaselineDirectory: baseline, TargetDirectory: target, Port: 0, Public: true})
	if err != nil || result != (Result{}) {
		t.Fatalf("Run() result = %#v, error = %v", result, err)
	}
	if len(presenter.starts) != 1 {
		t.Fatalf("presentations = %#v", presenter.starts)
	}
	start := presenter.starts[0]
	if start.LocalURL == "" || start.Port == 0 || start.BaselineDirectory != resolvedBaseline || start.TargetDirectory != resolvedTarget || len(start.NetworkURLs) != 1 || start.NetworkURLs[0] != "http://192.168.1.50:"+strconv.Itoa(start.Port) {
		t.Fatalf("startup = %#v", start)
	}
}

func TestModuleDoesNothingForAnAlreadyCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	presenter := &recordingDiffPresenter{}
	module, err := New(Dependencies{
		NetworkInterfaces: func() ([]NetworkInterface, error) {
			t.Fatal("NetworkInterfaces was called")
			return nil, nil
		},
		Presenter: presenter,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	result, err := module.Run(ctx, Input{})
	if err != nil || result != (Result{}) || len(presenter.starts) != 0 {
		t.Fatalf("Run() result = %#v, error = %v, presentations = %#v", result, err, presenter.starts)
	}
}

func TestDiffModuleRequiresItsTwoExternalAdapters(t *testing.T) {
	if _, err := New(Dependencies{}); err == nil || err.Error() != "diff network interface provider is required" {
		t.Fatalf("New() missing network adapter error = %v", err)
	}
	if _, err := New(Dependencies{NetworkInterfaces: func() ([]NetworkInterface, error) { return nil, nil }}); err == nil || err.Error() != "diff presenter is required" {
		t.Fatalf("New() missing presenter error = %v", err)
	}
}

type recordingDiffPresenter struct {
	starts       []Startup
	afterPresent func()
}

func (presenter *recordingDiffPresenter) Present(start Startup) error {
	presenter.starts = append(presenter.starts, start)
	if presenter.afterPresent != nil {
		presenter.afterPresent()
	}
	return nil
}
