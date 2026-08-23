package cm

import "testing"

func TestPresentSetSuccessReportsOnlyTheProfileIdentity(t *testing.T) {
	presenter := &recordingCMSetPresenter{}

	PresentSetSuccess(presenter, SetRequest{Profile: "work", Key: "apiKey", Value: "must-not-appear"})

	if presenter.message != "Profile work updated" {
		t.Fatalf("success message = %q", presenter.message)
	}
}

type recordingCMSetPresenter struct {
	message string
}

func (presenter *recordingCMSetPresenter) Success(message string) {
	presenter.message = message
}
