package agent

import (
	"context"
	"time"

	"vietclaw/internal/providers"
)

func (s *Service) insertCost(ctx context.Context, resp providers.ChatResponse) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO cost_events (provider, model, input_tokens, output_tokens, cost_usd, created_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		resp.Provider, resp.Model, resp.InputTokens, resp.OutputTokens, resp.EstimatedCostUSD, now)
	return err
}

func (s *Service) insertRun(ctx context.Context, id, sessionID, parentRunID, intent, provider, model, status, summary string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO agent_runs (id, session_id, parent_run_id, intent, provider, model, status, summary, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, sessionID, nullable(parentRunID), intent, nullable(provider), nullable(model), status, summary, now, now)
	return err
}

func (s *Service) logToolEvent(ctx context.Context, sessionID, toolName, input, output string, ok bool, errText string) {
	now := time.Now().UTC().Format(time.RFC3339)
	okVal := 0
	if ok {
		okVal = 1
	}
	result, err := s.db.ExecContext(ctx, `
INSERT INTO tool_events (session_id, tool_name, input, output, ok, error, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		nullable(sessionID), toolName, input, output, okVal, nullable(errText), now)
	if err != nil {
		s.logf("tool event log error: %v", err)
		return
	}
	eventID, _ := result.LastInsertId()
	s.publishSessionEvent(sessionID, SessionEvent{
		Event:     "tool_call",
		ToolName:  toolName,
		ToolInput: input,
	})
	resultEvent := SessionEvent{
		Event:       "tool_result",
		ToolName:    toolName,
		ToolResult:  output,
		ToolEventID: eventID,
	}
	if !ok && errText != "" {
		resultEvent.Error = errText
	}
	s.publishSessionEvent(sessionID, resultEvent)
}

func (s *Service) finishRun(ctx context.Context, id, status, summary, provider, model string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
UPDATE agent_runs SET status = ?, summary = ?, provider = ?, model = ?, updated_at = ? WHERE id = ?`,
		status, summary, nullable(provider), nullable(model), now, id)
	if err != nil {
		return err
	}
	s.publishRunLifecycleByRunID(id, status, summary)
	return nil
}

func (s *Service) publishRunLifecycleByRunID(runID, status, summary string) {
	var sessionID string
	if err := s.db.QueryRow(`SELECT session_id FROM agent_runs WHERE id = ?`, runID).Scan(&sessionID); err != nil || sessionID == "" {
		return
	}
	s.publishRunLifecycle(sessionID, status, summary)
}

func (s *Service) publishRunLifecycle(sessionID, status, summary string) {
	s.publishSessionEvent(sessionID, SessionEvent{
		Event:   "run_status",
		Status:  status,
		Summary: summary,
	})
	if status == RunStatusCompleted || status == RunStatusFailed || status == RunStatusBlocked || status == RunStatusNeedsApproval {
		s.publishSessionEvent(sessionID, SessionEvent{Event: "done"})
	}
}
