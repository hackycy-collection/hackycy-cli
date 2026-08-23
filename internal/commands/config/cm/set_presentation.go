package cm

import "fmt"

// SetPresenter renders the successful config cm set outcome.
type SetPresenter interface {
	Success(message string)
}

// PresentSetSuccess reports only the updated profile identity.
func PresentSetSuccess(presenter SetPresenter, request SetRequest) {
	presenter.Success(fmt.Sprintf("Profile %s updated", request.Profile))
}
