package harness

type Status string

const (
	StatusPlanned       Status = "planned"
	StatusNeedsApproval Status = "needs_approval"
	StatusFailed        Status = "failed"
)

type Budget struct {
	MaxTokens  int     `json:"max_tokens"`
	MaxUSD     float64 `json:"max_usd"`
	MaxMinutes int     `json:"max_minutes"`
}

type PlanStep struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tools       []string `json:"tools,omitempty"`
	Checks      []string `json:"checks,omitempty"`
	Name        string   `json:"name,omitempty"`
	Detail      string   `json:"detail,omitempty"`
}

type Plan struct {
	Goal        string     `json:"goal"`
	Mode        string     `json:"mode"`
	Risk        string     `json:"risk"`
	Summary     string     `json:"summary"`
	Assumptions []string   `json:"assumptions,omitempty"`
	Steps       []PlanStep `json:"steps"`
	StopRules   []string   `json:"stop_rules,omitempty"`
}

type Capsule struct {
	ID             string   `json:"id"`
	SessionID      string   `json:"session_id,omitempty"`
	Goal           string   `json:"goal"`
	Mode           string   `json:"mode"`
	Risk           string   `json:"risk"`
	Status         Status   `json:"status"`
	Budget         Budget   `json:"budget"`
	AllowedTools   []string `json:"allowed_tools"`
	ForbiddenTools []string `json:"forbidden_tools"`
	SuccessChecks  []string `json:"success_checks"`
	Provider       string   `json:"provider,omitempty"`
	Model          string   `json:"model,omitempty"`
	Summary        string   `json:"summary,omitempty"`
	Plan           Plan     `json:"plan"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

type Event struct {
	ID        int64  `json:"id"`
	RunID     string `json:"run_id"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}

type RunDetail struct {
	Run    Capsule `json:"run"`
	Events []Event `json:"events"`
}

type CreateRequest struct {
	SessionID      string   `json:"session_id,omitempty"`
	Goal           string   `json:"goal"`
	Mode           string   `json:"mode,omitempty"`
	Risk           string   `json:"risk,omitempty"`
	MaxTokens      int      `json:"max_tokens,omitempty"`
	MaxUSD         float64  `json:"max_usd,omitempty"`
	MaxMinutes     int      `json:"max_minutes,omitempty"`
	AllowedTools   []string `json:"allowed_tools,omitempty"`
	ForbiddenTools []string `json:"forbidden_tools,omitempty"`
	SuccessChecks  []string `json:"success_checks,omitempty"`
}
