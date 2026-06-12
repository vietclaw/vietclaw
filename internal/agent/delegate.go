package agent

import (
	"context"

	"vietclaw/internal/framework"
	"vietclaw/internal/router"
	"vietclaw/internal/tools"
)

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
	s.publishSessionEvent(childReq.SessionID, SessionEvent{
		Event:  "run_status",
		Status: RunStatusRunning,
	})

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
	return tools.IsFrameworkTool(name)
}
