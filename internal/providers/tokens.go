package providers

import "vietclaw/internal/config"

const (
	charsPerEstimatedToken       = 4
	unlimitedOutputTokenEstimate = 1024 // conservative budget estimate when no cap is set
	anthropicUnlimitedCeiling    = 8192 // Anthropic requires max_tokens; use model-typical ceiling
)

func EstimateMessagesTokens(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += EstimateTokens(msg.Content)
	}
	return total
}

func EstimateTokens(text string) int {
	n := len([]rune(text)) / charsPerEstimatedToken
	if n < 1 && text != "" {
		return 1
	}
	return n
}

func EstimateCostUSD(inTokens, outTokens int, cfg config.ProviderConfig) float64 {
	return (float64(inTokens)/1000)*cfg.CostPer1KIn + (float64(outTokens)/1000)*cfg.CostPer1KOut
}

// OutputTokenBudget returns the configured output cap, or a conservative estimate for cost/budget when unlimited.
func OutputTokenBudget(value int) int {
	if value > 0 {
		return value
	}
	return unlimitedOutputTokenEstimate
}

// AnthropicMaxOutputTokens returns max_tokens for Anthropic APIs (required param).
func AnthropicMaxOutputTokens(value int) int64 {
	if value > 0 {
		return int64(value)
	}
	return int64(anthropicUnlimitedCeiling)
}
