//go:build !windows

package terminaltest

import (
	"errors"
	"os/exec"

	"github.com/creack/pty"
)

// ErrPTYUnsupported is returned when a target cannot create a pseudoterminal.
var ErrPTYUnsupported = errors.New("controlled pseudoterminals are not supported on this target")

// StartPTY starts command with all standard streams attached to one controlled
// pseudoterminal.
func StartPTY(command *exec.Cmd) (*PTYProcess, error) {
	if command == nil {
		return nil, errors.New("PTY command is required")
	}
	terminal, err := pty.Start(command)
	if err != nil {
		return nil, err
	}
	return &PTYProcess{terminal: terminal, command: command}, nil
}
