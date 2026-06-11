package providers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"vietclaw/internal/config"
)

type OpenAICompatible struct {
	providerBase
	client *http.Client
}

func NewOpenAICompatible(cfg config.ProviderConfig, client *http.Client) *OpenAICompatible {
	return &OpenAICompatible{providerBase: providerBase{cfg: cfg}, client: client}
}

func (p *OpenAICompatible) openaiClient() (*openai.Client, error) {
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
	client := openai.NewClient(opts...)
	return &client, nil
}

func (p *OpenAICompatible) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	client, err := p.openaiClient()
	if err != nil {
		return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: err.Error()}, err
	}

	model := defaultString(req.Model, p.cfg.DefaultModel)
	openaiMessages := openaiMessagesFromChat(req)
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(model),
		Messages: openaiMessages,
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}
	if req.MaxOutputTokens > 0 {
		params.MaxTokens = openai.Int(int64(req.MaxOutputTokens))
	}
	if len(req.Tools) > 0 {
		params.Tools = openaiToolsFromChat(req.Tools)
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return ChatResponse{Provider: p.ID(), Model: req.Model, RawError: err.Error()}, err
	}

	text := ""
	if len(resp.Choices) > 0 {
		text = resp.Choices[0].Message.Content
	}

	inputTokens := int(resp.Usage.PromptTokens)
	outputTokens := int(resp.Usage.CompletionTokens)
	if inputTokens == 0 {
		inputTokens = EstimateMessagesTokens(req.Messages)
	}
	if outputTokens == 0 {
		outputTokens = EstimateTokens(text)
	}

	var toolCalls []ToolCall
	if len(resp.Choices) > 0 {
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			toolCalls = append(toolCalls, ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
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

func (p *OpenAICompatible) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	client, err := p.openaiClient()
	if err != nil {
		return nil, err
	}

	model := defaultString(req.Model, p.cfg.DefaultModel)
	openaiMessages := openaiMessagesFromChat(req)
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(model),
		Messages: openaiMessages,
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}
	if req.MaxOutputTokens > 0 {
		params.MaxTokens = openai.Int(int64(req.MaxOutputTokens))
	}
	if len(req.Tools) > 0 {
		params.Tools = openaiToolsFromChat(req.Tools)
	}

	stream := client.Chat.Completions.NewStreaming(ctx, params)
	ch := make(chan StreamChunk, 32)

	go func() {
		defer close(ch)
		acc := openai.ChatCompletionAccumulator{}
		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					ch <- StreamChunk{Text: delta.Content}
				}
			}
		}

		if err := stream.Err(); err != nil {
			ch <- StreamChunk{Error: err.Error()}
			return
		}

		if len(acc.Choices) > 0 {
			msg := acc.Choices[0].Message
			if len(msg.ToolCalls) > 0 {
				var tcList []ToolCall
				for _, tc := range msg.ToolCalls {
					tcList = append(tcList, ToolCall{
						ID:   tc.ID,
						Type: string(tc.Type),
						Function: ToolCallFunction{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					})
				}
				ch <- StreamChunk{ToolCalls: tcList}
			}
		}
		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}

func (p *OpenAICompatible) Embed(ctx context.Context, text string) ([]float32, error) {
	client, err := p.openaiClient()
	if err != nil {
		return nil, err
	}

	model := defaultString(p.cfg.EmbedModel, config.DefaultEmbedModel)
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model: openai.EmbeddingModel(model),
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding data returned")
	}

	values := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		values[i] = float32(v)
	}

	return values, nil
}

func (p *OpenAICompatible) EstimateCost(req ChatRequest) CostEstimate {
	inTokens := EstimateMessagesTokens(req.Messages)
	outTokens := OutputTokenBudget(req.MaxOutputTokens)
	return CostEstimate{
		InputTokens:      inTokens,
		OutputTokens:     outTokens,
		EstimatedCostUSD: EstimateCostUSD(inTokens, outTokens, p.cfg),
	}
}

func openaiMessagesFromChat(req ChatRequest) []openai.ChatCompletionMessageParamUnion {
	var out []openai.ChatCompletionMessageParamUnion
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			out = append(out, openai.ChatCompletionMessageParamUnion{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: openai.String(msg.Content),
					},
				},
			})
		case "user":
			out = append(out, openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(msg.Content),
					},
				},
			})
		case "assistant":
			var assistantMsg openai.ChatCompletionAssistantMessageParam
			if msg.Content != "" {
				assistantMsg.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.String(msg.Content),
				}
			}
			if len(msg.ToolCalls) > 0 {
				var tcParams []openai.ChatCompletionMessageToolCallParam
				for _, tc := range msg.ToolCalls {
					tcParams = append(tcParams, openai.ChatCompletionMessageToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					})
				}
				assistantMsg.ToolCalls = tcParams
			}
			out = append(out, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &assistantMsg,
			})
		case "tool":
			out = append(out, openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					ToolCallID: msg.ToolCallID,
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(msg.Content),
					},
				},
			})
		}
	}
	return out
}

func openaiToolsFromChat(tools []ToolDefinition) []openai.ChatCompletionToolParam {
	var out []openai.ChatCompletionToolParam
	for _, tool := range tools {
		out = append(out, openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Function.Name,
				Description: openai.String(tool.Function.Description),
				Parameters:  shared.FunctionParameters(tool.Function.Parameters),
			},
		})
	}
	return out
}
