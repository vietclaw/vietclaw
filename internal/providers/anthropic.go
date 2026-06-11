package providers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"vietclaw/internal/config"
)

type Anthropic struct {
	providerBase
	client *http.Client
}

func NewAnthropic(cfg config.ProviderConfig, client *http.Client) *Anthropic {
	cfg.Type = TypeAnthropic
	return &Anthropic{providerBase: providerBase{cfg: cfg}, client: client}
}

func (p *Anthropic) anthropicClient() (*anthropic.Client, error) {
	apiKey := os.Getenv(p.cfg.APIKeyEnv)
	if p.cfg.APIKeyEnv != "" && apiKey == "" {
		return nil, fmt.Errorf("missing api key env %s", p.cfg.APIKeyEnv)
	}

	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(p.client),
	}
	if p.cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(p.cfg.BaseURL))
	}
	client := anthropic.NewClient(opts...)
	return &client, nil
}

func (p *Anthropic) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	client, err := p.anthropicClient()
	if err != nil {
		return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: err.Error()}, err
	}

	model := defaultString(req.Model, p.cfg.DefaultModel)
	messages, systemPrompt := anthropicMessagesFromChat(req)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: AnthropicMaxOutputTokens(req.MaxOutputTokens),
		Messages:  messages,
	}
	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{
				Text: systemPrompt,
			},
		}
	}
	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(req.Temperature)
	}

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: err.Error()}, err
	}

	var textParts []string
	for _, block := range resp.Content {
		if block.Type == "text" && block.Text != "" {
			textParts = append(textParts, block.Text)
		}
	}
	text := strings.Join(textParts, "\n")

	inputTokens := int(resp.Usage.InputTokens)
	outputTokens := int(resp.Usage.OutputTokens)
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

func (p *Anthropic) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	client, err := p.anthropicClient()
	if err != nil {
		return nil, err
	}

	model := defaultString(req.Model, p.cfg.DefaultModel)
	messages, systemPrompt := anthropicMessagesFromChat(req)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: AnthropicMaxOutputTokens(req.MaxOutputTokens),
		Messages:  messages,
	}
	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{
				Text: systemPrompt,
			},
		}
	}
	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(req.Temperature)
	}

	stream := client.Messages.NewStreaming(ctx, params)
	ch := make(chan StreamChunk, 32)

	go func() {
		defer close(ch)
		for stream.Next() {
			event := stream.Current()
			switch eventVariant := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				if eventVariant.Delta.Type == "text_delta" && eventVariant.Delta.Text != "" {
					ch <- StreamChunk{Text: eventVariant.Delta.Text}
				}
			}
		}

		if err := stream.Err(); err != nil {
			ch <- StreamChunk{Error: err.Error()}
			return
		}
		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}

func (p *Anthropic) EstimateCost(req ChatRequest) CostEstimate {
	inTokens := EstimateMessagesTokens(req.Messages)
	outTokens := OutputTokenBudget(req.MaxOutputTokens)
	return CostEstimate{
		InputTokens:      inTokens,
		OutputTokens:     outTokens,
		EstimatedCostUSD: EstimateCostUSD(inTokens, outTokens, p.cfg),
	}
}

func (p *Anthropic) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("embeddings not supported by Anthropic provider")
}

func anthropicMessagesFromChat(req ChatRequest) ([]anthropic.MessageParam, string) {
	var system []string
	var messages []anthropic.MessageParam
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			system = append(system, msg.Content)
			continue
		}
		role := msg.Role
		if role != "assistant" {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		} else {
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}
	return messages, strings.Join(system, "\n\n")
}
