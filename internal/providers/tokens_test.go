package providers_test

import (
	"testing"

	"vietclaw/internal/config"
	"vietclaw/internal/providers"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty", "", 0},
		{"short string", "abc", 1},
		{"exact multiple", "abcd", 1},
		{"long string", "abcdefgh", 2},
		{"multi-byte characters", "xin chào", 2}, // 8 runes -> 8 / 4 = 2
		{"less than one token multi-byte", "xin", 1}, // 3 runes -> 1
		{"longer multi-byte", "xin chào các bạn", 4}, // 16 runes -> 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := providers.EstimateTokens(tt.text)
			if result != tt.expected {
				t.Errorf("EstimateTokens(%q) = %d, expected %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestEstimateMessagesTokens(t *testing.T) {
	tests := []struct {
		name     string
		messages []providers.Message
		expected int
	}{
		{"empty slice", []providers.Message{}, 0},
		{
			"single message",
			[]providers.Message{{Content: "abcdefgh"}}, // 8 runes -> 2
			2,
		},
		{
			"multiple messages",
			[]providers.Message{
				{Content: "abcd"}, // 1
				{Content: "efg"},  // 1
				{Content: "hijklmno"}, // 2
			},
			4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := providers.EstimateMessagesTokens(tt.messages)
			if result != tt.expected {
				t.Errorf("EstimateMessagesTokens() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestEstimateCostUSD(t *testing.T) {
	cfg := config.ProviderConfig{
		CostPer1KIn:  0.01,
		CostPer1KOut: 0.02,
	}

	tests := []struct {
		name      string
		inTokens  int
		outTokens int
		expected  float64
	}{
		{"zero tokens", 0, 0, 0.0},
		{"only input", 1000, 0, 0.01},
		{"only output", 0, 1000, 0.02},
		{"mixed", 1500, 2500, 0.065}, // 1.5 * 0.01 + 2.5 * 0.02 = 0.015 + 0.050 = 0.065
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := providers.EstimateCostUSD(tt.inTokens, tt.outTokens, cfg)
			// Using small epsilon for float comparison
			if result < tt.expected-1e-9 || result > tt.expected+1e-9 {
				t.Errorf("EstimateCostUSD(%d, %d) = %f, expected %f", tt.inTokens, tt.outTokens, result, tt.expected)
			}
		})
	}
}

func TestOutputTokenBudget(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"zero", 0, 1024}, // unlimitedOutputTokenEstimate
		{"negative", -5, 1024},
		{"positive", 2048, 2048},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := providers.OutputTokenBudget(tt.value)
			if result != tt.expected {
				t.Errorf("OutputTokenBudget(%d) = %d, expected %d", tt.value, result, tt.expected)
			}
		})
	}
}

func TestAnthropicMaxOutputTokens(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int64
	}{
		{"zero", 0, 8192}, // anthropicUnlimitedCeiling
		{"negative", -10, 8192},
		{"positive", 4096, 4096},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := providers.AnthropicMaxOutputTokens(tt.value)
			if result != tt.expected {
				t.Errorf("AnthropicMaxOutputTokens(%d) = %d, expected %d", tt.value, result, tt.expected)
			}
		})
	}
}
