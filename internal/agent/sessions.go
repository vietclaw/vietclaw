package agent

import (
	"context"
	"database/sql"
	"time"
)

func (s *Service) Sessions(ctx context.Context) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, channel, user_id, title, summary, created_at, updated_at FROM sessions ORDER BY updated_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []Session{}
	for rows.Next() {
		var item Session
		if err := rows.Scan(&item.ID, &item.Channel, &item.UserID, &item.Title, &item.Summary, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, item)
	}
	return sessions, rows.Err()
}

func (s *Service) SessionChildren(ctx context.Context, parentID string) ([]ChildSession, error) {
	pattern := parentID + ":spawn:%"
	rows, err := s.db.QueryContext(ctx, `
SELECT
  s.id,
  s.created_at,
  s.updated_at,
  COALESCE((
    SELECT content FROM messages
    WHERE session_id = s.id AND role = 'user'
    ORDER BY id ASC LIMIT 1
  ), '') AS task_preview,
  COALESCE((
    SELECT status FROM agent_runs
    WHERE session_id = s.id
    ORDER BY created_at DESC LIMIT 1
  ), 'running') AS run_status,
  EXISTS(
    SELECT 1 FROM messages
    WHERE session_id = s.id AND role = 'assistant' AND TRIM(content) != ''
  ) AS has_reply
FROM sessions s
WHERE s.id LIKE ? ESCAPE '\'
ORDER BY s.created_at ASC`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	children := []ChildSession{}
	for rows.Next() {
		var item ChildSession
		var hasReply int
		if err := rows.Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt, &item.TaskPreview, &item.RunStatus, &hasReply); err != nil {
			return nil, err
		}
		item.HasReply = hasReply == 1
		item.AgentID = SpawnAgentID(item.ID)
		children = append(children, item)
	}
	return children, rows.Err()
}

func (s *Service) SessionMessages(ctx context.Context, id string) (SessionDetail, error) {
	var detail SessionDetail
	err := s.db.QueryRowContext(ctx, `SELECT id, channel, user_id, title, summary, created_at, updated_at FROM sessions WHERE id = ?`, id).
		Scan(&detail.Session.ID, &detail.Session.Channel, &detail.Session.UserID, &detail.Session.Title, &detail.Session.Summary, &detail.Session.CreatedAt, &detail.Session.UpdatedAt)
	if err != nil {
		return detail, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, session_id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY id ASC`, id)
	if err != nil {
		return detail, err
	}
	defer rows.Close()
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return detail, err
		}
		detail.Messages = append(detail.Messages, msg)
	}
	if err := rows.Err(); err != nil {
		return detail, err
	}

	toolEvents, err := s.sessionToolEvents(ctx, id)
	if err != nil {
		return detail, err
	}
	detail.ToolEvents = toolEvents
	detail.RunStatus, detail.RunSummary = s.sessionRunStatus(ctx, id)
	return detail, nil
}

func (s *Service) SessionToolEventsAfter(ctx context.Context, sessionID string, afterID int64) ([]ToolEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, session_id, tool_name, input, output, ok, COALESCE(error, ''), created_at
FROM tool_events
WHERE session_id = ? AND id > ?
ORDER BY id ASC`, sessionID, afterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []ToolEvent{}
	for rows.Next() {
		var item ToolEvent
		var ok int
		if err := rows.Scan(&item.ID, &item.SessionID, &item.ToolName, &item.Input, &item.Output, &ok, &item.Error, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.OK = ok == 1
		events = append(events, item)
	}
	return events, rows.Err()
}

func (s *Service) sessionToolEvents(ctx context.Context, sessionID string) ([]ToolEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, session_id, tool_name, input, output, ok, COALESCE(error, ''), created_at
FROM tool_events
WHERE session_id = ?
ORDER BY id ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []ToolEvent{}
	for rows.Next() {
		var item ToolEvent
		var ok int
		if err := rows.Scan(&item.ID, &item.SessionID, &item.ToolName, &item.Input, &item.Output, &ok, &item.Error, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.OK = ok == 1
		events = append(events, item)
	}
	return events, rows.Err()
}

func (s *Service) SessionRunStatus(ctx context.Context, sessionID string) (status, summary string) {
	return s.sessionRunStatus(ctx, sessionID)
}

func (s *Service) sessionRunStatus(ctx context.Context, sessionID string) (status, summary string) {
	err := s.db.QueryRowContext(ctx, `
SELECT status, COALESCE(summary, '')
FROM agent_runs
WHERE session_id = ?
ORDER BY created_at DESC
LIMIT 1`, sessionID).Scan(&status, &summary)
	if err != nil {
		return "", ""
	}
	return status, summary
}

func (s *Service) ensureSession(ctx context.Context, req ChatRequest) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO sessions (id, channel, user_id, title, summary, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET updated_at = excluded.updated_at`,
		req.SessionID, req.Channel, req.UserID, sql.NullString{}, sql.NullString{}, now, now)
	return err
}

func (s *Service) addMessage(ctx context.Context, sessionID, role, content string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`, sessionID, role, content, now)
	if err != nil {
		return err
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE sessions SET updated_at = ? WHERE id = ?`, now, sessionID)
	return nil
}
