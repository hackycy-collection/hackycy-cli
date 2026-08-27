// Package terminal owns terminal capability and rendering behavior.
package terminal

import "strings"

// SessionKind describes the terminal capability selected for one invocation.
type SessionKind uint8

const (
	// Automation applies when ycy cannot safely own every inherited standard stream.
	Automation SessionKind = iota
	// PlainInteractive supports line-oriented interaction without terminal control.
	PlainInteractive
	// RichInteractive supports approved rich terminal controls.
	RichInteractive
)

// Session is the immutable terminal capability selected for one invocation.
type Session struct {
	Kind  SessionKind
	Color bool
}

// StreamFacts are the observable capability facts for one inherited stream.
type StreamFacts struct {
	Terminal bool
}

// LookupEnv looks up one environment variable while preserving whether it is set.
type LookupEnv func(string) (string, bool)

// Facts provides all capability inputs without consulting the process terminal.
type Facts struct {
	Stdin     StreamFacts
	Stdout    StreamFacts
	Stderr    StreamFacts
	LookupEnv LookupEnv
}

// Classify selects ycy's terminal behavior from injected stream and environment facts.
func Classify(facts Facts) Session {
	if !facts.Stdin.Terminal || !facts.Stdout.Terminal || !facts.Stderr.Terminal || facts.isSet("CI") {
		return Session{Kind: Automation}
	}

	term, _ := facts.lookup("TERM")
	if !supportsRichControls(term) {
		return Session{Kind: PlainInteractive}
	}

	noColor, noColorSet := facts.lookup("NO_COLOR")
	return Session{Kind: RichInteractive, Color: !noColorSet || noColor == ""}
}

func (facts Facts) isSet(key string) bool {
	_, set := facts.lookup(key)
	return set
}

func (facts Facts) lookup(key string) (string, bool) {
	if facts.LookupEnv == nil {
		return "", false
	}
	return facts.LookupEnv(key)
}

func supportsRichControls(term string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	for _, family := range []string{
		"alacritty",
		"foot",
		"iterm",
		"kitty",
		"konsole",
		"rxvt",
		"screen",
		"st",
		"tmux",
		"vte",
		"wezterm",
		"xterm",
	} {
		if term == family || strings.HasPrefix(term, family+"-") {
			return true
		}
	}
	return false
}
