package pulse

import (
	"reflect"
	"testing"
)

func TestSelectAuthorsSkipsThePromptForZeroOrOneAuthor(t *testing.T) {
	noCommits, cancelled := SelectAuthors(nil, panicPulseAuthorPrompter{})
	if cancelled || len(noCommits) != 0 {
		t.Fatalf("zero-author selection = (%#v, %t)", noCommits, cancelled)
	}

	commits := []Commit{{Author: "Ada", Subject: "one"}, {Author: "Ada", Subject: "two"}}
	selected, cancelled := SelectAuthors(commits, panicPulseAuthorPrompter{})
	if cancelled || !reflect.DeepEqual(selected, commits) {
		t.Fatalf("one-author selection = (%#v, %t), want %#v, false", selected, cancelled, commits)
	}
}

func TestSelectAuthorsUsesLegacyDefaultsAndFiltersSelection(t *testing.T) {
	testCases := []struct {
		name         string
		commits      []Commit
		selected     []string
		wantInitial  []string
		wantFiltered []Commit
	}{
		{
			name:         "two authors begin selected",
			commits:      []Commit{{Author: "Zed", Subject: "first"}, {Author: "Ada", Subject: "second"}},
			selected:     []string{"Ada"},
			wantInitial:  []string{"Ada", "Zed"},
			wantFiltered: []Commit{{Author: "Ada", Subject: "second"}},
		},
		{
			name:         "three authors begin selected",
			commits:      []Commit{{Author: "Cara"}, {Author: "Ada"}, {Author: "Ben"}},
			selected:     []string{"Ben"},
			wantInitial:  []string{"Ada", "Ben", "Cara"},
			wantFiltered: []Commit{{Author: "Ben"}},
		},
		{
			name:         "more than three authors begin unselected",
			commits:      []Commit{{Author: "Dora"}, {Author: "Cara"}, {Author: "Ada"}, {Author: "Ben"}},
			selected:     []string{"Dora"},
			wantInitial:  nil,
			wantFiltered: []Commit{{Author: "Dora"}},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			prompter := &scriptedPulseAuthorPrompter{selected: testCase.selected}
			got, cancelled := SelectAuthors(testCase.commits, prompter)
			if cancelled {
				t.Fatal("SelectAuthors() unexpectedly cancelled")
			}
			if !reflect.DeepEqual(got, testCase.wantFiltered) {
				t.Fatalf("filtered commits = %#v, want %#v", got, testCase.wantFiltered)
			}
			if got, want := prompter.prompt.Message, "Filter by authors:"; got != want {
				t.Fatalf("prompt message = %q, want %q", got, want)
			}
			if !prompter.prompt.Required {
				t.Fatal("author prompt is not required")
			}
			if !reflect.DeepEqual(prompter.prompt.InitialValues, testCase.wantInitial) {
				t.Fatalf("initial authors = %#v, want %#v", prompter.prompt.InitialValues, testCase.wantInitial)
			}
		})
	}
}

func TestSelectAuthorsReturnsCancellationWithoutFiltering(t *testing.T) {
	selected, cancelled := SelectAuthors([]Commit{{Author: "Ada"}, {Author: "Ben"}}, &scriptedPulseAuthorPrompter{cancelled: true})
	if !cancelled || selected != nil {
		t.Fatalf("SelectAuthors() = (%#v, %t), want nil, true", selected, cancelled)
	}
}

type scriptedPulseAuthorPrompter struct {
	prompt    AuthorPrompt
	selected  []string
	cancelled bool
}

func (prompter *scriptedPulseAuthorPrompter) SelectAuthors(prompt AuthorPrompt) ([]string, bool) {
	prompter.prompt = prompt
	return prompter.selected, prompter.cancelled
}

type panicPulseAuthorPrompter struct{}

func (panicPulseAuthorPrompter) SelectAuthors(AuthorPrompt) ([]string, bool) {
	panic("author prompt must not be called")
}
