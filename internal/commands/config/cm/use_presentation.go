package cm

import "fmt"

// UsePresenter renders the successful config cm use outcome.
type UsePresenter interface {
	Success(message string)
}

// PresentUseSuccess reports the selected default profile.
func PresentUseSuccess(presenter UsePresenter, profile string) {
	presenter.Success(fmt.Sprintf("Default CM profile set to %s", profile))
}
