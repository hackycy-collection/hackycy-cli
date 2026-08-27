//go:build windows

package terminaltest

import (
	"errors"
	"os/exec"
)

// ErrPTYUnsupported is returned when a target cannot create a pseudoterminal.
var ErrPTYUnsupported = errors.New("controlled pseudoterminals are not supported on this target")

// StartPTY reports that the controlled Unix PTY fixture is unavailable on
// Windows. Native Windows console tests remain a separate acceptance concern.
func StartPTY(_ *exec.Cmd) (*PTYProcess, error) {
	return nil, ErrPTYUnsupported
}
