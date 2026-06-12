package agent

import "database/sql"

const (
	DefaultUserID  = "local"
	DefaultChannel = "web"

	ProviderLocal = "local"
	ModelRule     = "rule"

	RoleUser      = "user"
	RoleAssistant = "assistant"

	RunStatusRunning       = "running"
	RunStatusCompleted     = "completed"
	RunStatusFailed        = "failed"
	RunStatusBlocked       = "blocked"
	RunStatusNeedsApproval = "needs_approval"
)

type ChatRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	AgentID   string `json:"agent_id,omitempty"`
	Channel   string `json:"channel"`
	Message   string `json:"message"`
	Mode      string `json:"mode"`
	Provider  string `json:"provider,omitempty"`
	Model     string `json:"model,omitempty"`
	CatalogID string `json:"catalog_id,omitempty"`
	ParentProvider string `json:"-"`
	ParentModel    string `json:"-"`
}

type ChatResponse struct {
	OK        bool    `json:"ok"`
	SessionID string  `json:"session_id"`
	AgentID   string  `json:"agent_id,omitempty"`
	Intent    string  `json:"intent"`
	Reply     string  `json:"reply"`
	Provider  string  `json:"provider"`
	Model     string  `json:"model"`
	CostUSD   float64 `json:"cost_usd"`
	Error     string  `json:"error,omitempty"`
}

type Session struct {
	ID        string         `json:"id"`
	Channel   string         `json:"channel"`
	UserID    string         `json:"user_id"`
	Title     sql.NullString `json:"title"`
	Summary   sql.NullString `json:"summary"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

type Message struct {
	ID        int64  `json:"id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type ToolEvent struct {
	ID        int64  `json:"id"`
	SessionID string `json:"session_id"`
	ToolName  string `json:"tool_name"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}

type SessionDetail struct {
	Session    Session     `json:"session"`
	Messages   []Message   `json:"messages"`
	ToolEvents []ToolEvent `json:"tool_events"`
	RunStatus  string      `json:"run_status,omitempty"`
	RunSummary string      `json:"run_summary,omitempty"`
}

type ChildSession struct {
	ID          string `json:"id"`
	AgentID     string `json:"agent_id"`
	TaskPreview string `json:"task_preview"`
	RunStatus   string `json:"run_status"`
	HasReply    bool   `json:"has_reply"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
