//go:build !windows

package upgrade

import (
	"errors"
	"os"
	"strings"
)

func isRetryableFileError(err error) bool {
	return errors.Is(err, os.ErrPermission) || strings.Contains(strings.ToLower(err.Error()), "busy") || strings.Contains(strings.ToLower(err.Error()), "sharing")
}
