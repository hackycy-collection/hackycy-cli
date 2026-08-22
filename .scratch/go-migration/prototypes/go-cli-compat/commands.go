package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var (
	errNonInteractive = errors.New("interactive input requires a TTY")
	errUserCancelled  = errors.New("user cancelled")
)

type gitCMInput struct {
	Push      optionalRemote `json:"push"`
	StagePush optionalRemote `json:"stagePush"`
	Excludes  []string       `json:"excludes,omitempty"`
}

type optionalRemote struct {
	Set    bool   `json:"set"`
	Remote string `json:"remote,omitempty"`
}

type diffInput struct {
	Baseline string   `json:"baseline"`
	Target   string   `json:"target"`
	Excludes []string `json:"excludes"`
}

type fsInput struct {
	Directory string   `json:"directory"`
	Accounts  []string `json:"accounts"`
}

type runInput struct {
	Path        string   `json:"path,omitempty"`
	Passthrough []string `json:"passthrough"`
}

type commandActions struct {
	out      io.Writer
	prompt   promptAdapter
	logLevel func() string
	internal func(context.Context, []string) error
}

func (a commandActions) gitCM(ctx context.Context, input gitCMInput) error {
	return a.print(ctx, "git cm", input)
}

func (a commandActions) diff(ctx context.Context, input diffInput) error {
	return a.print(ctx, "diff", input)
}

func (a commandActions) fs(ctx context.Context, input fsInput) error {
	return a.print(ctx, "fs", input)
}

func (a commandActions) run(ctx context.Context, input runInput) error {
	return a.print(ctx, "run", input)
}

func (a commandActions) choose(ctx context.Context) error {
	choice, err := a.prompt.selectOne(ctx, "Choose a probe result", []string{"alpha", "beta"})
	if err != nil {
		return err
	}
	return a.print(ctx, "prompt", struct {
		Choice string `json:"choice"`
	}{Choice: choice})
}

func (a commandActions) wait(ctx context.Context) error {
	fmt.Fprintln(a.out, "waiting; send SIGINT or SIGTERM")
	<-ctx.Done()
	return ctx.Err()
}

func (a commandActions) print(ctx context.Context, command string, input any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payload := struct {
		Command  string `json:"command"`
		LogLevel string `json:"logLevel"`
		Input    any    `json:"input"`
	}{Command: command, LogLevel: a.logLevel(), Input: input}
	encoder := json.NewEncoder(a.out)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(payload)
}

type promptAdapter struct {
	in          io.Reader
	out         io.Writer
	interactive bool
}

func (p promptAdapter) selectOne(ctx context.Context, label string, options []string) (string, error) {
	if !p.interactive {
		return "", errNonInteractive
	}

	fmt.Fprintf(p.out, "%s [1-%d, q cancels]: ", label, len(options))
	answer := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(p.in)
		if scanner.Scan() {
			answer <- strings.TrimSpace(scanner.Text())
			return
		}
		answer <- "q"
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case value := <-answer:
		if value == "q" {
			return "", errUserCancelled
		}
		for index, option := range options {
			if value == fmt.Sprint(index+1) {
				return option, nil
			}
		}
		return "", fmt.Errorf("invalid selection %q", value)
	}
}

type childExitError struct {
	code int
}

func (e childExitError) Error() string {
	return fmt.Sprintf("child exited with status %d", e.code)
}

func failureAction(kind string, childCode int) error {
	switch kind {
	case "action":
		return errors.New("prototype action failed")
	case "cancel":
		return errUserCancelled
	case "child":
		return childExitError{code: childCode}
	case "deadline":
		return context.DeadlineExceeded
	default:
		return fmt.Errorf("unknown failure kind %q", kind)
	}
}

func newDeadlineContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, time.Millisecond)
}
