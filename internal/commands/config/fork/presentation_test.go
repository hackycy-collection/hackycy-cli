package fork

import (
	"strings"
	"testing"
)

func TestRenderEmptyListShowsTheAddHint(t *testing.T) {
	got := Render(nil)
	want := "No instances configured. Run \"ycy config fork add\" to add one.\n"
	if got != want {
		t.Fatalf("Render(nil) = %q, want %q", got, want)
	}
}

func TestRenderListsSafeFieldsInStoredOrderWithoutSecrets(t *testing.T) {
	const plaintext = "token-that-must-not-escape"
	const ciphertext = "MDEyMzQ1Njc4OWFiY2RlZg==:tag:ciphertext"
	output := Render([]Instance{
		{Name: "work", Type: "gitlab", Scheme: "https", Host: "gitlab.example", TokenPreview: "MDEy***"},
		{Name: "personal", Type: "github", Scheme: "http", Host: "github.example", TokenPreview: "QWVy***"},
	})

	for _, field := range []string{
		"NAME", "TYPE", "SCHEME", "HOST", "TOKEN",
		"work", "gitlab", "https", "gitlab.example", "MDEy***",
		"personal", "github", "http", "github.example", "QWVy***",
		"2 instances configured",
	} {
		if !strings.Contains(output, field) {
			t.Fatalf("Render() output missing %q: %q", field, output)
		}
	}
	if strings.Index(output, "work") > strings.Index(output, "personal") {
		t.Fatalf("Render() did not preserve order: %q", output)
	}
	if strings.Contains(output, plaintext) || strings.Contains(output, ciphertext) {
		t.Fatalf("Render() exposed a secret: %q", output)
	}
}
