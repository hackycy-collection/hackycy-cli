package cm

import (
	"strings"
	"testing"
)

func TestDecodeCMTestProviderResponseReturnsTrimmedContentAndNormalizedUsage(t *testing.T) {
	result, err := decodeCMTestProviderResponse([]byte(`{
  "choices": [{"message": {"content": "  ok  "}}],
  "usage": {"prompt_tokens": 3, "completion_tokens": 2}
}`))
	if err != nil {
		t.Fatalf("decodeCMTestProviderResponse() returned an error: %v", err)
	}
	if result.Content != "ok" {
		t.Fatalf("result content = %q, want ok", result.Content)
	}
	assertCMTestUsage(t, result.Usage, 3, true, 2, true, 5, true)
}

func TestDecodeCMTestProviderResponsePrefersReportedUsableTotals(t *testing.T) {
	result, err := decodeCMTestProviderResponse([]byte(`{
  "choices": [{"message": {"content": "ok"}}],
  "usage": {"prompt_tokens": 3, "completion_tokens": 2, "total_tokens": 9}
}`))
	if err != nil {
		t.Fatalf("decodeCMTestProviderResponse() returned an error: %v", err)
	}
	assertCMTestUsage(t, result.Usage, 3, true, 2, true, 9, true)
}

func TestDecodeCMTestProviderResponseKeepsOnlyFiniteNumericUsageFields(t *testing.T) {
	result, err := decodeCMTestProviderResponse([]byte(`{
  "choices": [{"message": {"content": "ok"}}],
  "usage": {"prompt_tokens": 3, "completion_tokens": "2", "total_tokens": null}
}`))
	if err != nil {
		t.Fatalf("decodeCMTestProviderResponse() returned an error: %v", err)
	}
	assertCMTestUsage(t, result.Usage, 3, true, 0, false, 0, false)
}

func TestDecodeCMTestProviderResponseRejectsEmptyContentWithBoundedContext(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
		want string
	}{
		{
			name: "empty body",
			body: "",
			want: "Provider returned an empty response (finish_reason=unknown, response=<empty body>)",
		},
		{
			name: "blank choice content",
			body: `{"choices":[{"finish_reason":"length","message":{"content":"  "}}]}`,
			want: "Provider returned an empty response (finish_reason=length, response={\"choices\":[{\"finish_reason\":\"length\",\"message\":{\"content\":\"  \"}}]})",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeCMTestProviderResponse([]byte(test.body))
			if err == nil || err.Error() != test.want {
				t.Fatalf("decodeCMTestProviderResponse() error = %v, want %q", err, test.want)
			}
		})
	}

	if got, want := cmTestResponseSummary(strings.Repeat("x", 600)), strings.Repeat("x", 500)+"\u2026"; got != want {
		t.Fatalf("cmTestResponseSummary() = %q, want %q", got, want)
	}
}

func assertCMTestUsage(t *testing.T, usage *cmTestTokenUsage, prompt float64, hasPrompt bool, completion float64, hasCompletion bool, total float64, hasTotal bool) {
	t.Helper()
	if usage == nil {
		t.Fatal("usage = nil")
	}
	assertCMTestUsageField(t, "prompt", usage.PromptTokens, prompt, hasPrompt)
	assertCMTestUsageField(t, "completion", usage.CompletionTokens, completion, hasCompletion)
	assertCMTestUsageField(t, "total", usage.TotalTokens, total, hasTotal)
}

func assertCMTestUsageField(t *testing.T, name string, value *float64, want float64, present bool) {
	t.Helper()
	if !present {
		if value != nil {
			t.Fatalf("%s usage = %v, want absent", name, *value)
		}
		return
	}
	if value == nil || *value != want {
		t.Fatalf("%s usage = %v, want %v", name, value, want)
	}
}
