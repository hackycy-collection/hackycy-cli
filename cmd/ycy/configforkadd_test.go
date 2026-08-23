package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	configfork "github.com/hackycy/hackycy-cli/internal/commands/config/fork"
)

func TestTerminalForkAddPrompterValidatesTextWithoutTrimmingAcceptedInput(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalForkAddPrompter(strings.NewReader("\n host-with-spaces \n"), output)
	question := configfork.TextPrompt{
		Message:     "Host",
		Placeholder: "e.g. gitlab.company.com, github.com",
		Validate: func(value string) error {
			if strings.TrimSpace(value) == "" {
				return errors.New("Host is required")
			}
			return nil
		},
	}

	value, cancelled := prompter.Text(question)

	if cancelled || value != " host-with-spaces " {
		t.Fatalf("Text() = (%q, %t)", value, cancelled)
	}
	if !strings.Contains(output.String(), "Host") || !strings.Contains(output.String(), "e.g. gitlab.company.com, github.com") || !strings.Contains(output.String(), "Host is required") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestTerminalForkAddPrompterSelectsNumberedChoiceAndDefaultsToFirst(t *testing.T) {
	choices := []configfork.Choice{{Value: "gitlab", Label: "GitLab"}, {Value: "github", Label: "GitHub"}}
	output := &bytes.Buffer{}
	prompter := newTerminalForkAddPrompter(strings.NewReader("3\n2\n"), output)

	value, cancelled := prompter.Select(configfork.SelectPrompt{Message: "Provider type", Choices: choices})
	if cancelled || value != "github" {
		t.Fatalf("Select() = (%q, %t)", value, cancelled)
	}
	if !strings.Contains(output.String(), "GitLab") || !strings.Contains(output.String(), "GitHub") || !strings.Contains(output.String(), "Invalid selection") {
		t.Fatalf("output = %q", output.String())
	}

	defaultPrompter := newTerminalForkAddPrompter(strings.NewReader("\n"), &bytes.Buffer{})
	value, cancelled = defaultPrompter.Select(configfork.SelectPrompt{Message: "Provider type", Choices: choices})
	if cancelled || value != "gitlab" {
		t.Fatalf("Select() default = (%q, %t)", value, cancelled)
	}
}

func TestTerminalForkAddPrompterReadsPipedPasswordWithoutWritingIt(t *testing.T) {
	output := &bytes.Buffer{}
	prompter := newTerminalForkAddPrompter(strings.NewReader("secret-token\n"), output)

	value, cancelled := prompter.Password(configfork.TextPrompt{Message: "Access token", Validate: func(string) error { return nil }})

	if cancelled || value != "secret-token" {
		t.Fatalf("Password() = (%q, %t)", value, cancelled)
	}
	if strings.Contains(output.String(), "secret-token") {
		t.Fatalf("password prompt exposed input: %q", output.String())
	}
}

func TestTerminalForkAddPrompterTreatsEOFAsCancellation(t *testing.T) {
	prompter := newTerminalForkAddPrompter(strings.NewReader(""), &bytes.Buffer{})
	value, cancelled := prompter.Text(configfork.TextPrompt{Message: "Host", Validate: func(string) error { return nil }})
	if !cancelled || value != "" {
		t.Fatalf("Text() = (%q, %t)", value, cancelled)
	}
}

func TestTerminalForkAddPresenterWritesOutcomeMessages(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalForkAddPresenter{output: output}

	presenter.Cancel("Cancelled")
	presenter.Success("Instance work (gitlab.example) added successfully")

	if got, want := output.String(), "Cancelled\nInstance work (gitlab.example) added successfully\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
