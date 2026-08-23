//go:build windows

package main

import "os"

func handledYcySignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
