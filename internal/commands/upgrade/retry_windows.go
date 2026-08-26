//go:build windows

package upgrade

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func isRetryableFileError(err error) bool {
	return errors.Is(err, os.ErrPermission) || errors.Is(err, windows.ERROR_ACCESS_DENIED) || errors.Is(err, windows.ERROR_SHARING_VIOLATION)
}
