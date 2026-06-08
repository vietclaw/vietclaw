package providers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"google.golang.org/genai"
	"vietclaw/internal/config"
)

type Gemini struct {
	providerBase
	client *http.Client
}

func NewGemini(cfg config.ProviderConfig, client *http.Client) *Gemini {
	cfg.Type = TypeGemini
	return &Gemini{providerBase: providerBase{cfg: cfg}, client: client}
}

func (p *Gemini) genaiClient(ctx context.Context) (*genai.Client, error) {
	apiKey := os.Getenv(p.cfg.APIKeyEnv)
	if p.cfg.APIKeyEnv != "" && apiKey == "" {
		return nil, fmt.Errorf("missing api key env %s", p.cfg.APIKeyEnv)
	}

	clientConfig := &genai.ClientConfig{
		APIKey:     apiKey,
		HTTPClient: p.client,
		Backend:    genai.BackendGeminiAPI,
	}

	if p.cfg.BaseURL != "" {
		clientConfig.HTTPOptions = genai.HTTPOptions{
			BaseURL: p.cfg.BaseURL,
		}
	}

	return genai.NewClient(ctx, clientConfig)
}

func (p *Gemini) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	client, err := p.genaiClient(ctx)
	if err != nil {
		return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: err.Error()}, err
	}

	model := strings.TrimPrefix(defaultString(req.Model, p.cfg.DefaultModel), "models/")
	if model == "" {
		return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: "missing model"}, fmt.Errorf("missing model")
	}

	contents := geminiContentsFromChat(req)
	temp := float32(req.Temperature)
	genConfig := &genai.GenerateContentConfig{
		Temperature: &temp,
	}

	if req.MaxOutputTokens > 0 {
		genConfig.MaxOutputTokens = int32(req.MaxOutputTokens)
	}

	systemInstruction := systemInstructionFromChat(req)
	if systemInstruction != nil {
		genConfig.SystemInstruction = systemInstruction
	}

	resp, err := client.Models.GenerateContent(ctx, model, contents, genConfig)
	if err != nil {
		return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: err.Error()}, err
	}

	text := ""
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		text = resp.Candidates[0].Content.Parts[0].Text
	}

	inputTokens := int(resp.UsageMetadata.PromptTokenCount)
	outputTokens := int(resp.UsageMetadata.CandidatesTokenCount)
	if inputTokens == 0 {
		inputTokens = EstimateMessagesTokens(req.Messages)
	}
	if outputTokens == 0 {
		outputTokens = EstimateTokens(text)
	}

	return ChatResponse{
		Text:             text,
		Provider:         p.ID(),
		Model:            defaultString(req.Model, p.cfg.DefaultModel),
		InputTokens:      inputTokens,
		OutputTokens:     outputTokens,
		EstimatedCostUSD: EstimateCostUSD(inputTokens, outputTokens, p.cfg),
	}, nil
}

func (p *Gemini) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	client, err := p.genaiClient(ctx)
	if err != nil {
		return nil, err
	}

	model := strings.TrimPrefix(defaultString(req.Model, p.cfg.DefaultModel), "models/")
	if model == "" {
		return nil, fmt.Errorf("missing model")
	}

	contents := geminiContentsFromChat(req)
	temp := float32(req.Temperature)
	genConfig := &genai.GenerateContentConfig{
		Temperature: &temp,
	}

	if req.MaxOutputTokens > 0 {
		genConfig.MaxOutputTokens = int32(req.MaxOutputTokens)
	}

	systemInstruction := systemInstructionFromChat(req)
	if systemInstruction != nil {
		genConfig.SystemInstruction = systemInstruction
	}

	ch := make(chan StreamChunk, 32)
	go func() {
		defer close(ch)
		for result, err := range client.Models.GenerateContentStream(ctx, model, contents, genConfig) {
			if err != nil {
				ch <- StreamChunk{Error: err.Error()}
				return
			}
			text := result.Text()
			if text == "" && len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
				text = result.Candidates[0].Content.Parts[0].Text
			}
			if text != "" {
				ch <- StreamChunk{Text: text}
			}
		}
		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}

func (p *Gemini) Embed(ctx context.Context, text string) ([]float32, error) {
	client, err := p.genaiClient(ctx)
	if err != nil {
		return nil, err
	}

	model := strings.TrimPrefix(defaultString(p.cfg.EmbedModel, config.DefaultEmbedModel), "models/")
	if model == "" {
		return nil, fmt.Errorf("missing embed model")
	}

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: text},
			},
		},
	}

	resp, err := client.Models.EmbedContent(ctx, model, contents, nil)
	if err != nil {
		return nil, err
	}

	if len(resp.Embeddings) == 0 || len(resp.Embeddings[0].Values) == 0 {
		return nil, fmt.Errorf("empty embedding data returned")
	}
	return resp.Embeddings[0].Values, nil
}

func (p *Gemini) EstimateCost(req ChatRequest) CostEstimate {
	inTokens := EstimateMessagesTokens(req.Messages)
	outTokens := defaultOutputTokens(req.MaxOutputTokens)
	return CostEstimate{
		InputTokens:      inTokens,
		OutputTokens:     outTokens,
		EstimatedCostUSD: EstimateCostUSD(inTokens, outTokens, p.cfg),
	}
}

func geminiContentsFromChat(req ChatRequest) []*genai.Content {
	var contents []*genai.Content
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			continue
		}
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role: role,
			Parts: []*genai.Part{
				{Text: msg.Content},
			},
		})
	}
	return contents
}

func systemInstructionFromChat(req ChatRequest) *genai.Content {
	var systemParts []*genai.Part
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemParts = append(systemParts, &genai.Part{Text: msg.Content})
		}
	}
	if len(systemParts) == 0 {
		return nil
	}
	return &genai.Content{
		Parts: systemParts,
	}
}
