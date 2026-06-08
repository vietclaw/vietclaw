package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"vietclaw/internal/config"
)

const (
	// OpenCode Zen API — OpenAI-compatible endpoint.
	ZenBaseURL   = "https://opencode.ai/zen/v1"
	ZenModelsURL = "https://opencode.ai/zen/v1/models"
)

// OpenCodeZen is an OpenAI-compatible provider backed by OpenCode Zen.
// It reuses the OpenAICompatible provider under the hood so all tool-calling,
// streaming, and embedding logic is handled uniformly.
type OpenCodeZen struct {
	OpenAICompatible
}

func NewOpenCodeZen(cfg config.ProviderConfig, client *http.Client) *OpenCodeZen {
	cfg.Type = TypeOpenCodeZen
	if cfg.BaseURL == "" {
		cfg.BaseURL = ZenBaseURL
	}
	return &OpenCodeZen{
		OpenAICompatible: OpenAICompatible{
			providerBase: providerBase{cfg: cfg},
			client:       client,
		},
	}
}

// FetchZenModels queries the Zen /models endpoint and returns available model IDs.
// apiKey may be empty for providers that list models without auth.
func FetchZenModels(ctx context.Context, baseURL, apiKeyEnv string) ([]string, error) {
	modelsURL := baseURL + "/models"
	if baseURL == "" {
		modelsURL = ZenModelsURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build models request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if apiKeyEnv != "" {
		if key := os.Getenv(apiKeyEnv); key != "" {
			req.Header.Set("Authorization", "Bearer "+key)
		}
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("models request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models endpoint returned %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		// Some providers return a flat array.
		Models []string `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode models response: %w", err)
	}

	var ids []string
	for _, m := range result.Data {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	for _, m := range result.Models {
		if m != "" {
			ids = append(ids, m)
		}
	}
	return ids, nil
}

// Legacy OpenCodeCLI kept for backwards-compatibility but disabled by default.
// Use TypeOpenCodeZen instead.
type OpenCodeCLI struct {
	providerBase
}

func NewOpenCodeCLI(cfg config.ProviderConfig) *OpenCodeCLI {
	cfg.Type = TypeOpenCodeCLI
	return &OpenCodeCLI{providerBase: providerBase{cfg: cfg}}
}

func (p *OpenCodeCLI) Chat(_ context.Context, req ChatRequest) (ChatResponse, error) {
	return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: "opencode CLI mode is deprecated; use type opencode_zen instead"},
		fmt.Errorf("opencode CLI mode is deprecated; use type opencode_zen instead")
}

func (p *OpenCodeCLI) ChatStream(_ context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- StreamChunk{Error: "opencode CLI mode is deprecated; use type opencode_zen instead"}
	}()
	return ch, nil
}

func (p *OpenCodeCLI) Embed(_ context.Context, _ string) ([]float32, error) {
	return nil, fmt.Errorf("opencode CLI mode is deprecated")
}

func (p *OpenCodeCLI) EstimateCost(req ChatRequest) CostEstimate {
	return CostEstimate{InputTokens: EstimateMessagesTokens(req.Messages)}
}
