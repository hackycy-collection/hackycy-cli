package pulse

import (
	"reflect"
	"testing"
)

func TestSelectDaysUsesExplicitLegacyValuesWithoutPrompting(t *testing.T) {
	days := -1
	got, cancelled := selectDays(Input{Days: &days}, panicPulsePrompter{})
	if cancelled || got != -1 {
		t.Fatalf("selectDays() = (%d, %t), want -1, false", got, cancelled)
	}
}

func TestSelectDaysProvidesTheLegacyPromptChoices(t *testing.T) {
	prompter := &scriptedPulsePrompter{days: 7}
	got, cancelled := selectDays(Input{}, prompter)
	if cancelled || got != 7 {
		t.Fatalf("selectDays() = (%d, %t), want 7, false", got, cancelled)
	}
	want := DayPrompt{
		Message: "Select date range:",
		Options: []DayChoice{
			{Value: 1, Label: "Today"},
			{Value: 2, Label: "Yesterday"},
			{Value: 3, Label: "Last 3 days"},
			{Value: 7, Label: "Last 7 days"},
			{Value: 30, Label: "Last 30 days"},
		},
	}
	if !reflect.DeepEqual(prompter.dayPrompt, want) {
		t.Fatalf("day prompt = %#v, want %#v", prompter.dayPrompt, want)
	}
}

func TestSelectDaysReturnsPromptCancellation(t *testing.T) {
	_, cancelled := selectDays(Input{}, &scriptedPulsePrompter{daysCancelled: true})
	if !cancelled {
		t.Fatal("selectDays() did not return cancellation")
	}
}

type panicPulsePrompter struct{}

func (panicPulsePrompter) SelectDays(DayPrompt) (int, bool) {
	panic("day prompt must not be called")
}

func (panicPulsePrompter) SelectAuthors(AuthorPrompt) ([]string, bool) {
	panic("author prompt must not be called")
}
