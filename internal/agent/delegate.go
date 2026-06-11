package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"vietclaw/internal/framework"
	"vietclaw/internal/router"
	"vietclaw/internal/tools"
)

func (s *Service) handleDelegate(ctx context.Context, parentReq ChatRequest, parentRunID string, argsJSON string) (string, error) {
	var args struct {
		AgentID string `json:"agent_id"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid delegate args: %w", err)
	}
	agentID := strings.TrimSpace(args.AgentID)
	message := strings.TrimSpace(args.Message)
	if agentID == "" || message == "" {
		return "", fmt.Errorf("agent_id and message are required")
	}
	found := false
	for _, p := range s.cfg.Agents {
		if p.ID == agentID {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("agent profile not found: %s", agentID)
	}
	resp, err := s.Delegate(ctx, parentReq, parentRunID, agentID, message)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Delegated to %s: %s", agentID, resp.Reply), nil
}

func (s *Service) Delegate(ctx context.Context, parentReq ChatRequest, parentRunID, agentID, message string) (ChatResponse, error) {
	childReq := parentReq
	childReq.AgentID = agentID
	childReq.Message = message
	childReq.SessionID = parentReq.SessionID + ":delegate:" + agentID

	if err := s.ensureSession(ctx, childReq); err != nil {
		return ChatResponse{}, err
	}
	if err := s.addMessage(ctx, childReq.SessionID, RoleUser, message); err != nil {
		return ChatResponse{}, err
	}

	intent := s.router.Classify(ctx, message, s.profileLanguage(agentID))
	runID := newID("run")
	if err := s.insertRun(ctx, runID, childReq.SessionID, parentRunID, string(intent), "", "", RunStatusRunning, ""); err != nil {
		return ChatResponse{}, err
	}

	if s.framework != nil && s.framework.Hooks != nil {
		_ = s.framework.Hooks.Emit(ctx, framework.EventRunStart, framework.HookContext{
			RunID:       runID,
			ParentRunID: parentRunID,
			SessionID:   childReq.SessionID,
			AgentID:     agentID,
			Message:     message,
		})
	}

	switch intent {
	case router.IntentMemoryAdd:
		return s.handleMemoryAdd(ctx, childReq, runID, intent)
	case router.IntentMemoryQuery:
		return s.handleMemoryQuery(ctx, childReq, runID, intent)
	default:
		return s.runAgenticLoop(ctx, childReq, runID, intent)
	}
}

func (s *Service) isFrameworkTool(name string) bool {
	return strings.TrimSpace(name) == tools.ToolAgentDelegate
}
