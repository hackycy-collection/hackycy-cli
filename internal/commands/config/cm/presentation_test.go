package cm

import (
	"strings"
	"testing"
)

func TestRenderEmptyListShowsTheAddHint(t *testing.T) {
	got := Render(nil)
	want := "No CM profiles configured. Run \"ycy config cm add\" to add one.\n"
	if got != want {
		t.Fatalf("Render(nil) = %q, want %q", got, want)
	}
}

func TestRenderListsProfilesInStoredOrderWithTheDefaultMarkerAndNoSecrets(t *testing.T) {
	const plaintext = "api-key-that-must-not-escape"
	const ciphertext = "MDEyMzQ1Njc4OWFiY2RlZg==:tag:ciphertext"
	output := Render([]Profile{
		{Name: "work", Model: "gpt-4.1-mini", BaseURL: "https://work.example/v1"},
		{Name: "personal", Model: "deepseek-chat", BaseURL: "https://personal.example/v1", Default: true},
	})

	for _, field := range []string{
		"work", "gpt-4.1-mini", "https://work.example/v1",
		"personal", "deepseek-chat", "https://personal.example/v1",
	} {
		if !strings.Contains(output, field) {
			t.Fatalf("Render() output missing %q: %q", field, output)
		}
	}
	if !strings.Contains(output, "* personal") {
		t.Fatalf("Render() did not mark the default profile: %q", output)
	}
	if strings.Contains(output, "* work") {
		t.Fatalf("Render() marked a non-default profile: %q", output)
	}
	if strings.Index(output, "work") > strings.Index(output, "personal") {
		t.Fatalf("Render() did not preserve stored order: %q", output)
	}
	if strings.Contains(output, plaintext) || strings.Contains(output, ciphertext) {
		t.Fatalf("Render() exposed secret material: %q", output)
	}
}
