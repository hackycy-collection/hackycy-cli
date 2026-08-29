package test

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type cmTestProviderResult struct {
	Content string
	Usage   *cmTestTokenUsage
}

type cmTestTokenUsage struct {
	PromptTokens     *float64
	CompletionTokens *float64
	TotalTokens      *float64
}

func decodeCMTestProviderResponse(body []byte) (cmTestProviderResult, error) {
	raw := string(body)
	var decoded any = map[string]any{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &decoded); err != nil {
			return cmTestProviderResult{}, fmt.Errorf("decode CM test provider response: %w", err)
		}
	}

	payload, _ := decoded.(map[string]any)
	choices, _ := payload["choices"].([]any)
	var choice map[string]any
	if len(choices) > 0 {
		choice, _ = choices[0].(map[string]any)
	}
	message, _ := choice["message"].(map[string]any)
	content, _ := message["content"].(string)
	content = strings.TrimSpace(content)
	if content == "" {
		return cmTestProviderResult{}, fmt.Errorf("Provider returned an empty response (finish_reason=%s, response=%s)", cmTestFinishReason(choice), cmTestResponseSummary(raw))
	}
	return cmTestProviderResult{Content: content, Usage: normalizeCMTestUsage(payload["usage"])}, nil
}

func normalizeCMTestUsage(value any) *cmTestTokenUsage {
	usage, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	promptTokens, hasPromptTokens := cmTestFiniteNumber(usage["prompt_tokens"])
	completionTokens, hasCompletionTokens := cmTestFiniteNumber(usage["completion_tokens"])
	totalTokens, hasTotalTokens := cmTestFiniteNumber(usage["total_tokens"])
	if !hasTotalTokens && hasPromptTokens && hasCompletionTokens {
		totalTokens = promptTokens + completionTokens
		hasTotalTokens = true
	}
	if !hasPromptTokens && !hasCompletionTokens && !hasTotalTokens {
		return nil
	}
	result := cmTestTokenUsage{}
	if hasPromptTokens {
		result.PromptTokens = &promptTokens
	}
	if hasCompletionTokens {
		result.CompletionTokens = &completionTokens
	}
	if hasTotalTokens {
		result.TotalTokens = &totalTokens
	}
	return &result
}

func cmTestFiniteNumber(value any) (float64, bool) {
	number, ok := value.(float64)
	return number, ok && !math.IsInf(number, 0) && !math.IsNaN(number)
}

func cmTestFinishReason(choice map[string]any) string {
	if reason, found := choice["finish_reason"]; found && reason != nil {
		return fmt.Sprint(reason)
	}
	return "unknown"
}

func cmTestResponseSummary(value string) string {
	const limit = 500
	if value == "" {
		return "<empty body>"
	}
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= limit {
		return trimmed
	}
	return trimmed[:limit] + "\u2026"
}
