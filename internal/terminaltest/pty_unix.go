//go:build !windows

package terminaltest

import (
	"errors"
	"fmt"
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

// Resize updates the controlled pseudoterminal's visible size.
func (process *PTYProcess) Resize(width, height uint16) error {
	if process == nil || process.terminal == nil {
		return errors.New("PTY process is required")
	}
	if width == 0 || height == 0 {
		return fmt.Errorf("PTY size must be positive: %dx%d", width, height)
	}
	return pty.Setsize(process.terminal, &pty.Winsize{Cols: width, Rows: height})
}
