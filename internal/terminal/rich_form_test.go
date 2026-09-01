package terminal

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestRichListFormKeepsTheEndOfALongListVisible(t *testing.T) {
	options := make([]InteractionOption, 200)
	for index := range options {
		value := fmt.Sprintf("item-%03d", index)
		options[index] = InteractionOption{Label: value, Value: value}
	}
	form := newRichListForm(InteractionRequest{
		Kind:    InteractionSelect,
		Message: "Choose one",
		Options: options,
	}, 1)
	form.configure(80, 24, true)

	updated, _ := form.Update(tea.KeyPressMsg{Code: 'g', Mod: tea.ModShift, Text: "G"})
	form = updated.(*richListForm)
	view := form.View().Content
	for _, want := range []string{"item-180", "item-199"} {
		if !strings.Contains(view, want) {
			t.Fatalf("end-of-list view missing %q: %q", want, view)
		}
	}
}
