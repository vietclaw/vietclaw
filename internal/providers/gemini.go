package providers

import (
	"context"
	"encoding/json"
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

	var tools []*genai.Tool
	if len(req.Tools) > 0 {
		var err error
		tools, err = geminiToolsFromChat(req.Tools)
		if err != nil {
			return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: err.Error()}, err
		}
	}
	tools = append(tools, &genai.Tool{
		GoogleSearch: &genai.GoogleSearch{},
	})
	genConfig.Tools = tools

	resp, err := client.Models.GenerateContent(ctx, model, contents, genConfig)
	if err != nil {
		return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: err.Error()}, err
	}

	text := ""
	var toolCalls []ToolCall
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				if text != "" {
					text += "\n"
				}
				text += part.Text
			}
			if part.FunctionCall != nil {
				argsBytes, _ := json.Marshal(part.FunctionCall.Args)
				toolCalls = append(toolCalls, ToolCall{
					ID:   part.FunctionCall.ID,
					Type: "function",
					Function: ToolCallFunction{
						Name:      part.FunctionCall.Name,
						Arguments: string(argsBytes),
					},
				})
			}
		}
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
		ToolCalls:        toolCalls,
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

	var streamTools []*genai.Tool
	if len(req.Tools) > 0 {
		var err error
		streamTools, err = geminiToolsFromChat(req.Tools)
		if err != nil {
			return nil, err
		}
	}
	streamTools = append(streamTools, &genai.Tool{
		GoogleSearch: &genai.GoogleSearch{},
	})
	genConfig.Tools = streamTools

	ch := make(chan StreamChunk, 32)
	go func() {
		defer close(ch)
		for result, err := range client.Models.GenerateContentStream(ctx, model, contents, genConfig) {
			if err != nil {
				ch <- StreamChunk{Error: err.Error()}
				return
			}
			
			var chunkText string
			var chunkToolCalls []ToolCall
			if len(result.Candidates) > 0 && result.Candidates[0].Content != nil {
				for _, part := range result.Candidates[0].Content.Parts {
					if part.Text != "" {
						chunkText += part.Text
					}
					if part.FunctionCall != nil {
						argsBytes, _ := json.Marshal(part.FunctionCall.Args)
						chunkToolCalls = append(chunkToolCalls, ToolCall{
							ID:   part.FunctionCall.ID,
							Type: "function",
							Function: ToolCallFunction{
								Name:      part.FunctionCall.Name,
								Arguments: string(argsBytes),
							},
						})
					}
				}
			}
			if chunkText != "" {
				ch <- StreamChunk{Text: chunkText}
			}
			if len(chunkToolCalls) > 0 {
				ch <- StreamChunk{ToolCalls: chunkToolCalls}
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
		var parts []*genai.Part
		if msg.Role == "assistant" {
			role = "model"
			if msg.Content != "" {
				parts = append(parts, &genai.Part{Text: msg.Content})
			}
			for _, tc := range msg.ToolCalls {
				var args map[string]any
				if tc.Function.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
				}
				parts = append(parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						ID:   tc.ID,
						Name: tc.Function.Name,
						Args: args,
					},
				})
			}
		} else if msg.Role == "tool" {
			role = "user"
			parts = append(parts, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					ID:   msg.ToolCallID,
					Name: msg.Name,
					Response: map[string]any{
						"result": msg.Content,
					},
				},
			})
		} else {
			// user
			parts = append(parts, &genai.Part{Text: msg.Content})
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: parts,
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

func geminiToolsFromChat(tools []ToolDefinition) ([]*genai.Tool, error) {
	if len(tools) == 0 {
		return nil, nil
	}
	var decls []*genai.FunctionDeclaration
	for _, t := range tools {
		var schema *genai.Schema
		if len(t.Function.Parameters) > 0 {
			data, err := json.Marshal(t.Function.Parameters)
			if err != nil {
				return nil, err
			}
			var s genai.Schema
			if err := json.Unmarshal(data, &s); err != nil {
				return nil, err
			}
			schema = &s
		}
		decls = append(decls, &genai.FunctionDeclaration{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			Parameters:  schema,
		})
	}
	return []*genai.Tool{
		{FunctionDeclarations: decls},
	}, nil
}
