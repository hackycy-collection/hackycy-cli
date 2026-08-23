package cm

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestSetModuleUpdatesAndPresentsOnlyTheProfile(t *testing.T) {
	writer := &recordingCMSetWriter{}
	presenter := &recordingCMSetPresenter{}
	module, err := NewSet(SetDependencies{Writer: writer, Presenter: presenter})
	if err != nil {
		t.Fatalf("NewSet() returned an error: %v", err)
	}
	request := SetRequest{Profile: "work", Key: "apiKey", Value: "must-not-appear"}

	result, err := module.Run(context.Background(), request)
	if err != nil || result != (SetResult{}) {
		t.Fatalf("Run() = (%#v, %v)", result, err)
	}
	if got, want := writer.requests, []SetRequest{request}; !reflect.DeepEqual(got, want) {
		t.Fatalf("writer requests = %#v, want %#v", got, want)
	}
	if presenter.message != "Profile work updated" {
		t.Fatalf("success message = %q", presenter.message)
	}
}

func TestSetModuleReturnsFailuresWithoutPresentingSuccess(t *testing.T) {
	failure := errors.New("Unsupported key. Use baseURL, model, apiKey, temperature, timeoutMs, or maxOutputTokens.")
	presenter := &recordingCMSetPresenter{}
	module, err := NewSet(SetDependencies{
		Writer:    &recordingCMSetWriter{err: failure},
		Presenter: presenter,
	})
	if err != nil {
		t.Fatalf("NewSet() returned an error: %v", err)
	}

	result, err := module.Run(context.Background(), SetRequest{Profile: "work", Key: "unsupported", Value: "value"})
	if !errors.Is(err, failure) || result != (SetResult{}) || presenter.message != "" {
		t.Fatalf("Run() = (%#v, %v, presenter=%#v)", result, err, presenter)
	}
}

func TestNewSetRequiresEachCommandOwnedAdapter(t *testing.T) {
	for _, dependencies := range []SetDependencies{
		{Presenter: &recordingCMSetPresenter{}},
		{Writer: &recordingCMSetWriter{}},
	} {
		if _, err := NewSet(dependencies); err == nil {
			t.Fatalf("NewSet(%#v) returned nil error", dependencies)
		}
	}
}
