package agent

import (
	"context"
	"fmt"
	"strings"

	"vietclaw/internal/framework"
	"vietclaw/internal/i18n"
	"vietclaw/internal/providers"
	"vietclaw/internal/router"
)

func (s *Service) ChatStream(ctx context.Context, req ChatRequest) (<-chan providers.StreamChunk, error) {
	req = normalizeRequest(req, s.cfg)
	req = s.enrichRequestModel(ctx, req)
	req = s.applyAgentProfile(ctx, req)
	if strings.TrimSpace(req.Message) == "" {
		return nil, fmt.Errorf("%s", s.text(i18n.AgentMessageRequired))
	}

	if err := s.ensureSession(ctx, req); err != nil {
		return nil, err
	}
	if err := s.addMessage(ctx, req.SessionID, RoleUser, req.Message); err != nil {
		return nil, err
	}

	intent := s.router.Classify(ctx, req.Message, s.Language())
	runID := newID("run")
	if err := s.insertRun(ctx, runID, req.SessionID, "", string(intent), "", "", RunStatusRunning, ""); err != nil {
		return nil, err
	}

	switch intent {
	case router.IntentMemoryAdd:
		return s.streamRuleResponse(ctx, req, runID, intent, s.handleMemoryAdd)
	case router.IntentMemoryQuery:
		return s.streamRuleResponse(ctx, req, runID, intent, s.handleMemoryQuery)
	default:
		return s.StreamAgenticLoop(ctx, req, runID, intent)
	}
}

func (s *Service) runAgenticLoop(ctx context.Context, req ChatRequest, runID string, intent router.Intent) (ChatResponse, error) {
	embedder := s.router.SelectDefaultEmbedder()
	messages, err := s.context.Messages(ctx, req.SessionID, s.memoryScope(req), req.AgentID, req.Message, embedder)
	if err != nil {
		_ = s.finishRun(ctx, runID, RunStatusFailed, err.Error(), "", "")
		return ChatResponse{}, err
	}
	messages = s.applyProfilePersona(req, messages)

	chatReq := s.loopChatRequest(req, messages)
	selection, excludedProviders, err := s.selectLoopProvider(ctx, req, chatReq, nil)
	if err != nil {
		return s.selectionError(ctx, req, runID, intent, err), nil
	}
	chatReq.Model = selection.Model
	req.ParentProvider = selection.Provider.ID()
	req.ParentModel = selection.Model

	var finalReply string
	var finalProvider string
	var finalModel string
	var totalCost float64
	var accumulatedText string
	maxSteps := s.profileMaxSteps(req.AgentID)
	toolCtx := s.toolContext(ctx, req)
	scope := s.memoryScope(req)

	for step := 1; ; step++ {
		if maxSteps > 0 && step > maxSteps {
			finalReply = finalizeAtMaxSteps(accumulatedText, s.text(i18n.AgentMaxStepsReached))
			break
		}
		providerResp, nextSelection, nextExcluded, err := s.chatWithFallback(ctx, chatReq, selection, excludedProviders, s.profile(req.AgentID).Providers)
		if err != nil {
			_ = s.finishRun(ctx, runID, RunStatusFailed, err.Error(), selection.Provider.ID(), selection.Model)
			return ChatResponse{
				OK:        false,
				SessionID: req.SessionID,
				AgentID:   req.AgentID,
				Intent:    string(intent),
				Provider:  selection.Provider.ID(),
				Model:     selection.Model,
				Error:     err.Error(),
			}, err
		}
		selection = nextSelection
		excludedProviders = nextExcluded
		chatReq.Model = selection.Model

		totalCost += providerResp.EstimatedCostUSD
		finalProvider = providerResp.Provider
		finalModel = providerResp.Model
		if providerResp.Text != "" {
			accumulatedText += providerResp.Text
			s.publishSessionEvent(req.SessionID, SessionEvent{
				Event: "text",
				Text:  providerResp.Text,
			})
		}

		if len(providerResp.ToolCalls) > 0 {
			chatReq.Messages = append(chatReq.Messages, providers.Message{
				Role:      RoleAssistant,
				Content:   accumulatedText,
				ToolCalls: providerResp.ToolCalls,
			})
			if err := s.addMessage(ctx, req.SessionID, RoleAssistant, accumulatedText); err != nil {
				return ChatResponse{}, err
			}
			if err := s.executeToolCalls(toolCtx, req, runID, req.SessionID, providerResp.ToolCalls, &chatReq, scope, embedder); err != nil {
				return ChatResponse{}, err
			}
			continue
		}

		finalReply = accumulatedText
		break
	}

	_ = s.addMessage(ctx, req.SessionID, RoleAssistant, finalReply)
	_ = s.insertCost(ctx, providers.ChatResponse{
		Provider:         finalProvider,
		Model:            finalModel,
		EstimatedCostUSD: totalCost,
	})
	_ = s.finishRun(ctx, runID, RunStatusCompleted, finalReply, finalProvider, finalModel)

	return ChatResponse{
		OK:        true,
		SessionID: req.SessionID,
		AgentID:   req.AgentID,
		Intent:    string(intent),
		Reply:     finalReply,
		Provider:  finalProvider,
		Model:     finalModel,
		CostUSD:   totalCost,
	}, nil
}

func (s *Service) StreamAgenticLoop(ctx context.Context, req ChatRequest, runID string, intent router.Intent) (<-chan providers.StreamChunk, error) {
	ch := make(chan providers.StreamChunk)

	go func() {
		defer close(ch)

		ch <- providers.StreamChunk{Event: "session", SessionID: req.SessionID}

		embedder := s.router.SelectDefaultEmbedder()
		messages, err := s.context.Messages(ctx, req.SessionID, s.memoryScope(req), req.AgentID, req.Message, embedder)
		if err != nil {
			_ = s.finishRun(ctx, runID, RunStatusFailed, err.Error(), "", "")
			ch <- providers.StreamChunk{Error: err.Error()}
			return
		}
		messages = s.applyProfilePersona(req, messages)

		chatReq := s.loopChatRequest(req, messages)
		selection, excludedProviders, err := s.selectLoopProvider(ctx, req, chatReq, nil)
		if err != nil {
			resp := s.selectionError(ctx, req, runID, intent, err)
			ch <- providers.StreamChunk{Error: resp.Error}
			return
		}
		chatReq.Model = selection.Model
		req.ParentProvider = selection.Provider.ID()
		req.ParentModel = selection.Model

		var finalProvider string
		var finalModel string
		var totalCost float64
		var accumulatedText string
		maxSteps := s.profileMaxSteps(req.AgentID)
		spawnNotify := func(agentID, status, summary, childSessionID, parentSessionID string) {
			ch <- providers.StreamChunk{
				Event:           "spawn",
				ToolName:        agentID,
				ToolInput:       status,
				ToolResult:      summary,
				SessionID:       childSessionID,
				ParentSessionID: parentSessionID,
			}
		}
		toolCtx := withSpawnNotifier(s.toolContext(ctx, req), spawnNotify)
		scope := s.memoryScope(req)

		for step := 1; ; step++ {
			if err := ctx.Err(); err != nil {
				_ = s.finishRun(ctx, runID, RunStatusFailed, err.Error(), finalProvider, finalModel)
				ch <- providers.StreamChunk{Error: err.Error()}
				return
			}
			if maxSteps > 0 && step > maxSteps {
				before := accumulatedText
				accumulatedText = finalizeAtMaxSteps(accumulatedText, s.text(i18n.AgentMaxStepsReached))
				if len(accumulatedText) > len(before) {
					ch <- providers.StreamChunk{Event: "text", Text: accumulatedText[len(before):]}
				}
				break
			}
			attempt, nextSelection, nextExcluded, err := s.streamWithFallback(ctx, ch, chatReq, selection, excludedProviders, s.profile(req.AgentID).Providers)
			if err != nil {
				_ = s.finishRun(ctx, runID, RunStatusFailed, err.Error(), selection.Provider.ID(), selection.Model)
				ch <- providers.StreamChunk{Error: err.Error()}
				return
			}
			selection = nextSelection
			excludedProviders = nextExcluded
			chatReq.Model = selection.Model

			finalProvider = selection.Provider.ID()
			finalModel = selection.Model
			if attempt.text != "" {
				accumulatedText += attempt.text
			}

			tempReq := chatReq
			tempReq.MaxOutputTokens = providers.EstimateTokens(accumulatedText)
			totalCost += selection.Provider.EstimateCost(tempReq).EstimatedCostUSD

			if len(attempt.toolCalls) > 0 {
				chatReq.Messages = append(chatReq.Messages, providers.Message{
					Role:      RoleAssistant,
					Content:   accumulatedText,
					ToolCalls: attempt.toolCalls,
				})
				if err := s.addMessage(ctx, req.SessionID, RoleAssistant, accumulatedText); err != nil {
					ch <- providers.StreamChunk{Error: err.Error()}
					return
				}

				for _, tc := range attempt.toolCalls {
					ch <- providers.StreamChunk{
						Event:     "tool_call",
						ToolName:  tc.Function.Name,
						ToolInput: tc.Function.Arguments,
					}
					toolResult, err := s.executeToolCall(toolCtx, req, runID, req.SessionID, tc, &chatReq, scope, embedder)
					if err != nil {
						ch <- providers.StreamChunk{Error: err.Error()}
						return
					}
					ch <- providers.StreamChunk{
						Event:      "tool_result",
						ToolName:   tc.Function.Name,
						ToolResult: toolResult,
					}
				}
				continue
			}
			break
		}

		_ = s.addMessage(ctx, req.SessionID, RoleAssistant, accumulatedText)
		_ = s.insertCost(ctx, providers.ChatResponse{
			Provider:         finalProvider,
			Model:            finalModel,
			EstimatedCostUSD: totalCost,
		})
		_ = s.finishRun(ctx, runID, RunStatusCompleted, accumulatedText, finalProvider, finalModel)
		ch <- providers.StreamChunk{Done: true}
	}()

	return ch, nil
}

type ruleHandler func(context.Context, ChatRequest, string, router.Intent) (ChatResponse, error)

func (s *Service) streamRuleResponse(ctx context.Context, req ChatRequest, runID string, intent router.Intent, handle ruleHandler) (<-chan providers.StreamChunk, error) {
	ch := make(chan providers.StreamChunk, 2)
	go func() {
		defer close(ch)
		ch <- providers.StreamChunk{Event: "session", SessionID: req.SessionID}
		resp, err := handle(ctx, req, runID, intent)
		if err != nil {
			ch <- providers.StreamChunk{Error: err.Error()}
			return
		}
		const chunkSize = 64
		text := resp.Reply
		for i := 0; i < len(text); i += chunkSize {
			end := i + chunkSize
			if end > len(text) {
				end = len(text)
			}
			ch <- providers.StreamChunk{Event: "text", Text: text[i:end]}
		}
		ch <- providers.StreamChunk{Done: true}
	}()
	return ch, nil
}

func (s *Service) loopChatRequest(req ChatRequest, messages []providers.Message) providers.ChatRequest {
	includeDelegate := s.cfg.Framework.Enabled && s.cfg.Framework.DelegateEnabled
	return providers.ChatRequest{
		SessionID:       req.SessionID,
		Messages:        messages,
		Temperature:     defaultTemperature,
		MaxOutputTokens: s.maxOutputTokens(),
		Metadata: map[string]any{
			"user_id":  req.UserID,
			"channel":  req.Channel,
			"mode":     req.Mode,
			"language": s.profileLanguage(req.AgentID),
		},
		Tools: s.tools.GetDefinitionsForProfile(s.profile(req.AgentID), includeDelegate),
	}
}

func (s *Service) maxOutputTokens() int {
	return s.cfg.Agent.MaxOutputTokens
}

func (s *Service) selectionError(ctx context.Context, req ChatRequest, runID string, intent router.Intent, err error) ChatResponse {
	reply := err.Error()
	_ = s.addMessage(ctx, req.SessionID, RoleAssistant, reply)
	_ = s.finishRun(ctx, runID, RunStatusNeedsApproval, reply, "", "")
	return ChatResponse{
		OK:        false,
		SessionID: req.SessionID,
		AgentID:   req.AgentID,
		Intent:    string(intent),
		Reply:     reply,
		Error:     reply,
	}
}

func (s *Service) chatWithFallback(ctx context.Context, req providers.ChatRequest, selection router.Selection, excluded []string, allowed []string) (providers.ChatResponse, router.Selection, []string, error) {
	for {
		resp, err := selection.Provider.Chat(ctx, req)
		if err == nil {
			return resp, selection, excluded, nil
		}
		s.logf("provider %s chat error: %v", selection.Provider.ID(), err)
		excluded = append(excluded, selection.Provider.ID())
		next, fallbackErr := s.router.SelectForProfile(ctx, req, excluded, allowed)
		if fallbackErr != nil {
			return providers.ChatResponse{}, selection, excluded, err
		}
		selection = next
		req.Model = selection.Model
	}
}

type streamAttempt struct {
	text      string
	toolCalls []providers.ToolCall
}

func (s *Service) streamWithFallback(ctx context.Context, out chan<- providers.StreamChunk, req providers.ChatRequest, selection router.Selection, excluded []string, allowed []string) (streamAttempt, router.Selection, []string, error) {
	for {
		streamCh, err := selection.Provider.ChatStream(ctx, req)
		if err != nil {
			s.logf("provider %s stream init error: %v", selection.Provider.ID(), err)
			excluded = append(excluded, selection.Provider.ID())
			next, fallbackErr := s.router.SelectForProfile(ctx, req, excluded, allowed)
			if fallbackErr != nil {
				return streamAttempt{}, selection, excluded, err
			}
			selection = next
			req.Model = selection.Model
			continue
		}

		var attempt streamAttempt
		failed := ""
		for chunk := range streamCh {
			if chunk.Error != "" {
				failed = chunk.Error
				break
			}
			if chunk.Text != "" {
				attempt.text += chunk.Text
				out <- providers.StreamChunk{Event: "text", Text: chunk.Text}
			}
			if len(chunk.ToolCalls) > 0 {
				attempt.toolCalls = append(attempt.toolCalls, chunk.ToolCalls...)
			}
		}
		if failed == "" {
			return attempt, selection, excluded, nil
		}

		s.logf("provider %s stream generation error: %s", selection.Provider.ID(), failed)
		excluded = append(excluded, selection.Provider.ID())
		next, fallbackErr := s.router.SelectForProfile(ctx, req, excluded, allowed)
		if fallbackErr != nil {
			return streamAttempt{}, selection, excluded, fmt.Errorf("%s", failed)
		}
		selection = next
		req.Model = selection.Model
	}
}

func (s *Service) executeToolCalls(ctx context.Context, req ChatRequest, runID, sessionID string, calls []providers.ToolCall, chatReq *providers.ChatRequest, scope string, embedder providers.Provider) error {
	for _, tc := range calls {
		if _, err := s.executeToolCall(ctx, req, runID, sessionID, tc, chatReq, scope, embedder); err != nil {
			return err
		}
	}
	return nil
}

func finalizeAtMaxSteps(accumulated, limitMsg string) string {
	if strings.TrimSpace(accumulated) == "" {
		return limitMsg
	}
	return accumulated + "\n\n" + limitMsg
}

func (s *Service) executeToolCall(ctx context.Context, req ChatRequest, runID, sessionID string, tc providers.ToolCall, chatReq *providers.ChatRequest, scope string, embedder providers.Provider) (string, error) {
	var toolResult string
	var execErr error
	if s.isFrameworkTool(tc.Function.Name) {
		toolResult, execErr = s.handleFrameworkTool(ctx, req, runID, tc.Function.Name, tc.Function.Arguments, spawnNotifierFromContext(ctx))
	} else if s.framework != nil && s.framework.Config.HooksEnabled {
		_ = s.framework.Hooks.Emit(ctx, framework.EventBeforeTool, framework.HookContext{
			SessionID: sessionID,
			AgentID:   req.AgentID,
			ToolName:  tc.Function.Name,
			ToolInput: tc.Function.Arguments,
		})
		toolResult, execErr = s.tools.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
	} else {
		toolResult, execErr = s.tools.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
	}
	ok := execErr == nil
	errText := ""
	if execErr != nil {
		errText = execErr.Error()
		toolResult = formatToolFailureMessage(s.profile(req.AgentID).Language, toolResult, execErr)
		s.recordToolReflection(ctx, scope, tc.Function.Name, tc.Function.Arguments, toolResult, embedder)
	}
	s.logToolEvent(ctx, sessionID, tc.Function.Name, tc.Function.Arguments, toolResult, ok, errText)
	if s.framework != nil && s.framework.Config.HooksEnabled {
		_ = s.framework.Hooks.Emit(ctx, framework.EventAfterTool, framework.HookContext{
			SessionID:  sessionID,
			AgentID:    req.AgentID,
			ToolName:   tc.Function.Name,
			ToolInput:  tc.Function.Arguments,
			ToolOutput: toolResult,
			ToolError:  errText,
		})
	}

	chatReq.Messages = append(chatReq.Messages, providers.Message{
		Role:       "tool",
		Name:       tc.Function.Name,
		ToolCallID: tc.ID,
		Content:    toolResult,
	})

	toolPersistText := fmt.Sprintf("[Tool Execution: %s]\nInput: %s\nOutput: %s", tc.Function.Name, tc.Function.Arguments, toolResult)
	if err := s.addMessage(ctx, sessionID, "system", toolPersistText); err != nil {
		return "", err
	}
	return toolResult, nil
}
