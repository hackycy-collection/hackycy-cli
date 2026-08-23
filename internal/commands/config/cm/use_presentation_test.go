package cm

import "testing"

func TestPresentUseSuccessReportsOnlyTheSelectedProfile(t *testing.T) {
	presenter := &recordingCMUsePresenter{}

	PresentUseSuccess(presenter, "work")

	if presenter.message != "Default CM profile set to work" {
		t.Fatalf("success message = %q", presenter.message)
	}
}

type recordingCMUsePresenter struct {
	message string
}

func (presenter *recordingCMUsePresenter) Success(message string) {
	presenter.message = message
}
