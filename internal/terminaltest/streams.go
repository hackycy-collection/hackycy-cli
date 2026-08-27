package terminaltest

import (
	"bytes"
	"strings"
)

// RedirectedStreams supplies explicitly redirected standard streams for a
// command or adapter test.
type RedirectedStreams struct {
	Stdin  *strings.Reader
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
}

// NewRedirectedStreams creates isolated stream buffers with the given input.
func NewRedirectedStreams(input string) *RedirectedStreams {
	return &RedirectedStreams{
		Stdin:  strings.NewReader(input),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
}

// ContainsTerminalControl reports escape or C1 CSI bytes that are forbidden
// in Automation and Plain Interactive output. Ordinary line separators remain
// allowed.
func ContainsTerminalControl(output []byte) bool {
	for _, value := range output {
		if value == '\x1b' || value == '\x9b' {
			return true
		}
	}
	return false
}
