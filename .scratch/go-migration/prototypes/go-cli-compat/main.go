package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var interrupted atomic.Int32
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		sig := <-signals
		switch sig {
		case os.Interrupt:
			interrupted.Store(130)
		case syscall.SIGTERM:
			interrupted.Store(143)
		default:
			interrupted.Store(1)
		}
		cancel()
	}()

	stdinInfo, _ := os.Stdin.Stat()
	stdoutInfo, _ := os.Stdout.Stat()
	adapter := newCLIAdapter(streams{
		in:  os.Stdin,
		out: os.Stdout,
		err: os.Stderr,
		interactive: stdinInfo != nil && stdoutInfo != nil &&
			stdinInfo.Mode()&os.ModeCharDevice != 0 && stdoutInfo.Mode()&os.ModeCharDevice != 0,
	}, environmentMap(os.Environ()))
	result := adapter.execute(ctx, os.Args[1:], func() int { return int(interrupted.Load()) })
	if result.message != "" {
		fmt.Fprintln(os.Stderr, result.message)
	}
	os.Exit(result.code)
}
