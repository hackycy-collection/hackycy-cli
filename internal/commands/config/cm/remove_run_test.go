package cm

import (
	"context"
	"errors"
	"testing"
)

func TestRemoveModuleRetainsNonMutatingOutcomes(t *testing.T) {
	tests := []struct {
		name          string
		confirmed     bool
		confirmCancel bool
		want          RemoveResult
	}{
		{name: "confirmation cancellation", confirmCancel: true, want: RemoveResult{Cancelled: true}},
		{name: "negative confirmation", want: RemoveResult{Declined: true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := &recordingCMRemoveStore{removed: true}
			presenter := &recordingCMRemovePresenter{}
			module := newCMRemoveModule(t, &scriptedCMRemoveRunPrompter{confirmed: test.confirmed, cancelled: test.confirmCancel}, writer, presenter)

			result, err := module.Run(context.Background(), RemoveRequest{Profile: "work"})

			if err != nil || result != test.want {
				t.Fatalf("Run() = (%#v, %v), want (%#v, nil)", result, err, test.want)
			}
			if len(writer.names) != 0 {
				t.Fatalf("Run() wrote on a non-mutating branch: %#v", writer.names)
			}
			if presenter.cancellation != "Cancelled" || presenter.success != "" {
				t.Fatalf("presenter = %#v", presenter)
			}
		})
	}
}

func TestRemoveModuleRemovesAndPresentsTheConfirmedProfile(t *testing.T) {
	writer := &recordingCMRemoveStore{removed: true}
	presenter := &recordingCMRemovePresenter{}
	module := newCMRemoveModule(t, &scriptedCMRemoveRunPrompter{confirmed: true}, writer, presenter)

	result, err := module.Run(context.Background(), RemoveRequest{Profile: "work"})

	if err != nil || result != (RemoveResult{}) {
		t.Fatalf("Run() = (%#v, %v)", result, err)
	}
	if got, want := writer.names, []string{"work"}; !sameCMRemoveStrings(got, want) {
		t.Fatalf("remove names = %#v, want %#v", got, want)
	}
	if presenter.success != "Profile work removed" || presenter.cancellation != "" {
		t.Fatalf("presenter = %#v", presenter)
	}
}

func TestRemoveModuleReturnsMissingAndWriteFailuresWithoutSuccessPresentation(t *testing.T) {
	tests := []struct {
		name      string
		writer    *recordingCMRemoveStore
		wantError string
	}{
		{name: "missing", writer: &recordingCMRemoveStore{}, wantError: "CM profile not found: missing"},
		{name: "write", writer: &recordingCMRemoveStore{err: errors.New("publish configuration")}, wantError: "publish configuration"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			presenter := &recordingCMRemovePresenter{}
			module := newCMRemoveModule(t, &scriptedCMRemoveRunPrompter{confirmed: true}, test.writer, presenter)

			result, err := module.Run(context.Background(), RemoveRequest{Profile: "missing"})

			if err == nil || err.Error() != test.wantError || result != (RemoveResult{}) {
				t.Fatalf("Run() = (%#v, %v), want error %q", result, err, test.wantError)
			}
			if presenter.cancellation != "" || presenter.success != "" {
				t.Fatalf("failure presented an outcome: %#v", presenter)
			}
		})
	}
}

func TestNewRemoveRequiresEveryCommandOwnedAdapter(t *testing.T) {
	prompter := &scriptedCMRemoveRunPrompter{}
	writer := &recordingCMRemoveStore{}
	presenter := &recordingCMRemovePresenter{}

	for _, dependencies := range []RemoveDependencies{
		{Writer: writer, Presenter: presenter},
		{Prompter: prompter, Presenter: presenter},
		{Prompter: prompter, Writer: writer},
	} {
		if _, err := NewRemove(dependencies); err == nil {
			t.Fatalf("NewRemove(%#v) returned nil error", dependencies)
		}
	}
}

func newCMRemoveModule(t *testing.T, prompter RemoveConfirmationPrompter, writer RemoveWriter, presenter RemovePresenter) *RemoveModule {
	t.Helper()
	module, err := NewRemove(RemoveDependencies{Prompter: prompter, Writer: writer, Presenter: presenter})
	if err != nil {
		t.Fatalf("NewRemove() returned an error: %v", err)
	}
	return module
}

type scriptedCMRemoveRunPrompter struct {
	confirmed bool
	cancelled bool
}

func (prompter *scriptedCMRemoveRunPrompter) Confirm(RemoveConfirmPrompt) (bool, bool) {
	return prompter.confirmed, prompter.cancelled
}
