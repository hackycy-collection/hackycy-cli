//go:build !windows

package main

import (
	"os"
	"syscall"
)

func handledYcySignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
