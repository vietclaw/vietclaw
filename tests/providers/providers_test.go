package providers_test

import (
	"context"
	"strings"
	"testing"

	"vietclaw/internal/config"
	"vietclaw/internal/providers"
)

func TestMockProviderDeterministic(t *testing.T) {
	p := providers.New(config.ProviderConfig{ID: "mock", Type: "mock", Enabled: true, DefaultModel: "mock-small"})
	resp, err := p.Chat(context.Background(), providers.ChatRequest{
		Messages: []providers.Message{{Role: "user", Content: "mày là gì"}},
		Model:    "mock-small",
		Metadata: map[string]any{"language": "vi"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Provider != "mock" || resp.Model != "mock-small" || resp.Text == "" || resp.EstimatedCostUSD != 0 {
		t.Fatalf("unexpected mock response: %#v", resp)
	}
	if !strings.Contains(resp.Text, "VietClaw") {
		t.Fatalf("unexpected localized mock response: %q", resp.Text)
	}
}

func TestMockProviderEnglish(t *testing.T) {
	p := providers.New(config.ProviderConfig{ID: "mock", Type: "mock", Enabled: true, DefaultModel: "mock-small"})
	resp, err := p.Chat(context.Background(), providers.ChatRequest{
		Messages: []providers.Message{{Role: "user", Content: "what are you"}},
		Model:    "mock-small",
		Metadata: map[string]any{"language": "en"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Text, "lightweight agent runtime") {
		t.Fatalf("unexpected english mock response: %q", resp.Text)
	}
}
