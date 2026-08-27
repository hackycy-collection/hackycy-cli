package terminal_test

import (
	"testing"

	"github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name  string
		facts terminaltest.Facts
		want  terminal.Session
	}{
		{
			name:  "recognized terminal is rich",
			facts: allTerminalFacts(map[string]string{"TERM": "xterm-256color"}),
			want:  terminal.Session{Kind: terminal.RichInteractive, Color: true},
		},
		{
			name:  "no color keeps rich session",
			facts: allTerminalFacts(map[string]string{"TERM": "screen-256color", "NO_COLOR": "1"}),
			want:  terminal.Session{Kind: terminal.RichInteractive},
		},
		{
			name:  "empty no color keeps color enabled",
			facts: allTerminalFacts(map[string]string{"TERM": "tmux-256color", "NO_COLOR": ""}),
			want:  terminal.Session{Kind: terminal.RichInteractive, Color: true},
		},
		{
			name:  "dumb terminal is plain",
			facts: allTerminalFacts(map[string]string{"TERM": "dumb"}),
			want:  terminal.Session{Kind: terminal.PlainInteractive},
		},
		{
			name:  "unknown terminal is plain",
			facts: allTerminalFacts(map[string]string{"TERM": "unrecognized-terminal"}),
			want:  terminal.Session{Kind: terminal.PlainInteractive},
		},
		{
			name:  "missing terminal is plain",
			facts: allTerminalFacts(nil),
			want:  terminal.Session{Kind: terminal.PlainInteractive},
		},
		{
			name: "redirected stdin is automation",
			facts: terminaltest.Facts{
				Stdout:      terminaltest.StreamFacts{Terminal: true},
				Stderr:      terminaltest.StreamFacts{Terminal: true},
				Environment: map[string]string{"TERM": "xterm-256color"},
			},
			want: terminal.Session{Kind: terminal.Automation},
		},
		{
			name: "redirected stdout is automation",
			facts: terminaltest.Facts{
				Stdin:       terminaltest.StreamFacts{Terminal: true},
				Stderr:      terminaltest.StreamFacts{Terminal: true},
				Environment: map[string]string{"TERM": "xterm-256color"},
			},
			want: terminal.Session{Kind: terminal.Automation},
		},
		{
			name: "redirected stderr is automation",
			facts: terminaltest.Facts{
				Stdin:       terminaltest.StreamFacts{Terminal: true},
				Stdout:      terminaltest.StreamFacts{Terminal: true},
				Environment: map[string]string{"TERM": "xterm-256color"},
			},
			want: terminal.Session{Kind: terminal.Automation},
		},
		{
			name:  "set empty CI is automation",
			facts: allTerminalFacts(map[string]string{"TERM": "xterm-256color", "CI": ""}),
			want:  terminal.Session{Kind: terminal.Automation},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := terminal.Classify(factsFrom(test.facts)); got != test.want {
				t.Fatalf("terminal.Classify() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func allTerminalFacts(environment map[string]string) terminaltest.Facts {
	return terminaltest.Facts{
		Stdin:       terminaltest.StreamFacts{Terminal: true},
		Stdout:      terminaltest.StreamFacts{Terminal: true},
		Stderr:      terminaltest.StreamFacts{Terminal: true},
		Environment: environment,
	}
}

func factsFrom(facts terminaltest.Facts) terminal.Facts {
	return terminal.Facts{
		Stdin:     terminal.StreamFacts{Terminal: facts.Stdin.Terminal},
		Stdout:    terminal.StreamFacts{Terminal: facts.Stdout.Terminal},
		Stderr:    terminal.StreamFacts{Terminal: facts.Stderr.Terminal},
		LookupEnv: facts.LookupEnv,
	}
}
