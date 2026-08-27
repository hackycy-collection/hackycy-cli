package terminal

import (
	"bytes"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTrackedTeaModelUsesTheCurrentPhaseOnNarrowTerminals(t *testing.T) {
	cancellations := 0
	model := newTrackedTeaModel("Git Pulse", 40, false, &bytes.Buffer{}, func() {
		cancellations++
	})
	model.applyPhase(OperationPhase{Name: "Scanning repositories", Detail: "workspace/project", State: PhaseCompleted})
	model.applyPhase(OperationPhase{Name: "Fetching commits", Detail: "workspace/project", State: PhaseActive})

	view := model.View()
	if !strings.Contains(view, "Git Pulse") || !strings.Contains(view, "Fetching commits") || !strings.Contains(view, "workspace/project") {
		t.Fatalf("narrow view = %q", view)
	}
	if strings.Contains(view, "Scanning repositories") {
		t.Fatalf("narrow view retained completed phase: %q", view)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cancellations != 0 || !strings.Contains(model.View(), "Press Esc again to cancel") {
		t.Fatalf("first Esc = cancellations=%d, view=%q", cancellations, model.View())
	}
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cancellations != 1 || !strings.Contains(model.View(), "Cancelling...") {
		t.Fatalf("second Esc = cancellations=%d, view=%q", cancellations, model.View())
	}
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cancellations != 1 {
		t.Fatalf("Ctrl-C requested cancellation more than once: %d", cancellations)
	}
}
