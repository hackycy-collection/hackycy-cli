package cm

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestUseModuleSelectsAndPresentsTheRequestedProfile(t *testing.T) {
	writer := &recordingCMUseWriter{}
	presenter := &recordingCMUsePresenter{}
	module, err := NewUse(UseDependencies{Writer: writer, Presenter: presenter})
	if err != nil {
		t.Fatalf("NewUse() returned an error: %v", err)
	}

	result, err := module.Run(context.Background(), UseRequest{Profile: "work"})
	if err != nil || result != (UseResult{}) {
		t.Fatalf("Run() = (%#v, %v)", result, err)
	}
	if got, want := writer.names, []string{"work"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("writer names = %#v, want %#v", got, want)
	}
	if presenter.message != "Default CM profile set to work" {
		t.Fatalf("success message = %q", presenter.message)
	}
}

func TestUseModuleReturnsSelectionFailuresWithoutPresentingSuccess(t *testing.T) {
	failure := errors.New("CM profile not found: missing")
	presenter := &recordingCMUsePresenter{}
	module, err := NewUse(UseDependencies{
		Writer:    &recordingCMUseWriter{err: failure},
		Presenter: presenter,
	})
	if err != nil {
		t.Fatalf("NewUse() returned an error: %v", err)
	}

	result, err := module.Run(context.Background(), UseRequest{Profile: "missing"})
	if !errors.Is(err, failure) || result != (UseResult{}) || presenter.message != "" {
		t.Fatalf("Run() = (%#v, %v, presenter=%#v)", result, err, presenter)
	}
}

func TestNewUseRequiresEachCommandOwnedAdapter(t *testing.T) {
	for _, dependencies := range []UseDependencies{
		{Presenter: &recordingCMUsePresenter{}},
		{Writer: &recordingCMUseWriter{}},
	} {
		if _, err := NewUse(dependencies); err == nil {
			t.Fatalf("NewUse(%#v) returned nil error", dependencies)
		}
	}
}
