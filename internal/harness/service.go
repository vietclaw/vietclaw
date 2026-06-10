package harness

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"vietclaw/internal/agent"
	"vietclaw/internal/config"
)

const (
	defaultMode       = "agentless"
	defaultRisk       = "medium"
	defaultMaxTokens  = 24000
	defaultMaxUSD     = 0.25
	defaultMaxMinutes = 20
)

type Service struct {
	cfg config.Config
	db  *sql.DB
}

func New(cfg config.Config, db *sql.DB) *Service {
	return &Service{cfg: cfg, db: db}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Capsule, error) {
	req = normalizeCreateRequest(req)
	runID := newID("harness")
	now := time.Now().UTC().Format(time.RFC3339)
	capsule := Capsule{
		ID:             runID,
		SessionID:      req.SessionID,
		Goal:           req.Goal,
		Mode:           req.Mode,
		Risk:           req.Risk,
		Status:         StatusPlanned,
		Budget:         Budget{MaxTokens: req.MaxTokens, MaxUSD: req.MaxUSD, MaxMinutes: req.MaxMinutes},
		AllowedTools:   req.AllowedTools,
		ForbiddenTools: req.ForbiddenTools,
		SuccessChecks:  req.SuccessChecks,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.insertRun(ctx, capsule); err != nil {
		return Capsule{}, err
	}
	_ = s.addEvent(ctx, runID, "capsule.created", mustJSON(map[string]any{
		"goal":            capsule.Goal,
		"mode":            capsule.Mode,
		"risk":            capsule.Risk,
		"allowed_tools":   capsule.AllowedTools,
		"forbidden_tools": capsule.ForbiddenTools,
		"success_checks":  capsule.SuccessChecks,
	}))

	plan, providerID, model, err := s.planWithProvider(ctx, capsule)
	if err != nil {
		capsule.Status = StatusNeedsApproval
		capsule.Summary = err.Error()
		capsule.Plan = fallbackPlan(capsule, err.Error())
		_ = s.updateRunPlan(ctx, capsule)
		_ = s.addEvent(ctx, runID, "plan.failed", mustJSON(map[string]any{"error": err.Error()}))
		return capsule, nil
	}

	capsule.Provider = providerID
	capsule.Model = model
	capsule.Plan = plan
	capsule.Summary = plan.Summary
	if err := s.updateRunPlan(ctx, capsule); err != nil {
		return Capsule{}, err
	}
	_ = s.addEvent(ctx, runID, "plan.created", mustJSON(map[string]any{
		"provider": providerID,
		"model":    model,
	}))
	return capsule, nil
}

func (s *Service) List(ctx context.Context, limit int) ([]Capsule, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, session_id, goal, mode, risk, status, budget_json, allowed_tools_json,
       forbidden_tools_json, success_checks_json, provider, model, summary, plan_json, created_at, updated_at
FROM harness_runs
ORDER BY created_at DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Capsule
	for rows.Next() {
		capsule, err := scanCapsule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, capsule)
	}
	return out, rows.Err()
}

func (s *Service) Detail(ctx context.Context, id string) (RunDetail, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, session_id, goal, mode, risk, status, budget_json, allowed_tools_json,
       forbidden_tools_json, success_checks_json, provider, model, summary, plan_json, created_at, updated_at
FROM harness_runs
WHERE id = ?`, id)
	capsule, err := scanCapsule(row)
	if err != nil {
		return RunDetail{}, err
	}
	events, err := s.events(ctx, id)
	if err != nil {
		return RunDetail{}, err
	}
	return RunDetail{Run: capsule, Events: events}, nil
}

func (s *Service) planWithProvider(ctx context.Context, capsule Capsule) (Plan, string, string, error) {
	prompt := "Lập kế hoạch JSON cho VietClaw Harness. Chỉ trả JSON có summary, steps, stop_rules. " +
		"Steps gồm 3 mục localize, patch, verify. " +
		"Goal: " + capsule.Goal + ". Mode: " + capsule.Mode + ". Risk: " + capsule.Risk + ". " +
		"Allowed tools: " + strings.Join(capsule.AllowedTools, ", ") + ". " +
		"Forbidden tools: " + strings.Join(capsule.ForbiddenTools, ", ") + ". " +
		"Success checks: " + strings.Join(capsule.SuccessChecks, "; ") + "."
	resp, err := agent.NewService(s.cfg, s.db).Chat(ctx, agent.ChatRequest{
		SessionID: capsule.ID + "_planner",
		UserID:    agent.DefaultUserID,
		Channel:   "harness",
		Message:   prompt,
		Mode:      capsule.Mode,
	})
	if err != nil {
		return Plan{}, resp.Provider, resp.Model, err
	}
	plan := parsePlan(resp.Reply)
	plan.Goal = capsule.Goal
	plan.Mode = capsule.Mode
	plan.Risk = capsule.Risk
	if plan.Summary == "" {
		plan.Summary = "đã tạo plan harness"
	}
	if len(plan.Steps) == 0 {
		_ = s.addEvent(ctx, capsule.ID, "plan.parse_failed", mustJSON(map[string]any{
			"text_len": len(resp.Reply),
			"preview":  truncate(resp.Reply, 600),
		}))
		plan = fallbackPlan(capsule, "provider returned an empty plan")
	}
	return plan, resp.Provider, resp.Model, nil
}

func normalizeCreateRequest(req CreateRequest) CreateRequest {
	req.Goal = strings.TrimSpace(req.Goal)
	req.Mode = strings.ToLower(strings.TrimSpace(req.Mode))
	req.Risk = strings.ToLower(strings.TrimSpace(req.Risk))
	if req.Mode == "" {
		req.Mode = defaultMode
	}
	if req.Risk == "" {
		req.Risk = inferRisk(req)
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = defaultMaxTokens
	}
	if req.MaxUSD <= 0 {
		req.MaxUSD = defaultMaxUSD
	}
	if req.MaxMinutes <= 0 {
		req.MaxMinutes = defaultMaxMinutes
	}
	if len(req.AllowedTools) == 0 {
		req.AllowedTools = []string{"file.read", "file.patch", "test.run"}
	}
	if len(req.ForbiddenTools) == 0 {
		req.ForbiddenTools = []string{"secret.read", "shell.network", "shell.delete", "git.push"}
	}
	if len(req.SuccessChecks) == 0 {
		req.SuccessChecks = []string{"focused tests pass", "diff stays inside requested scope", "evidence ledger has command/test results"}
	}
	return req
}

func inferRisk(req CreateRequest) string {
	text := strings.ToLower(req.Goal + " " + strings.Join(req.AllowedTools, " "))
	highMarkers := []string{"push", "deploy", "delete", "secret", "token", "network", "shell", "production", "prod"}
	for _, marker := range highMarkers {
		if strings.Contains(text, marker) {
			return "high"
		}
	}
	if strings.Contains(text, "docs") || strings.Contains(text, "readme") || strings.Contains(text, "test") {
		return "low"
	}
	return defaultRisk
}

func fallbackPlan(capsule Capsule, reason string) Plan {
	return Plan{
		Goal:    capsule.Goal,
		Mode:    capsule.Mode,
		Risk:    capsule.Risk,
		Summary: reason,
		Assumptions: []string{
			"provider plan không khả dụng nên dùng plan an toàn mặc định",
		},
		Steps: []PlanStep{
			{ID: "step_1", Title: "localize", Description: "xác định file và context liên quan trước khi sửa", Tools: []string{"file.read"}},
			{ID: "step_2", Title: "patch", Description: "áp dụng thay đổi nhỏ theo đúng scope", Tools: []string{"file.patch"}},
			{ID: "step_3", Title: "verify", Description: "chạy check tập trung và ghi bằng chứng", Tools: []string{"test.run"}, Checks: capsule.SuccessChecks},
		},
		StopRules: []string{"dừng nếu cần secret/network/shell ngoài allowlist", "dừng nếu diff vượt scope"},
	}
}

func parsePlan(text string) Plan {
	cleaned := strings.TrimSpace(text)
	if start := strings.Index(cleaned, "{"); start >= 0 {
		cleaned = cleaned[start:]
	}
	if end := strings.LastIndex(cleaned, "}"); end >= 0 {
		cleaned = cleaned[:end+1]
	}
	var plan Plan
	if err := json.Unmarshal([]byte(cleaned), &plan); err == nil && len(plan.Steps) > 0 {
		normalizePlanSteps(&plan)
		return plan
	}
	var flexible struct {
		Summary     string         `json:"summary"`
		Assumptions []string       `json:"assumptions"`
		Steps       map[string]any `json:"steps"`
		StopRules   []string       `json:"stop_rules"`
	}
	if err := json.Unmarshal([]byte(cleaned), &flexible); err == nil {
		plan.Summary = flexible.Summary
		plan.Assumptions = flexible.Assumptions
		plan.StopRules = flexible.StopRules
		for _, key := range []string{"localize", "patch", "verify"} {
			if value, ok := flexible.Steps[key]; ok {
				plan.Steps = append(plan.Steps, PlanStep{ID: key, Title: key, Description: stepDescription(value)})
			}
		}
	}
	return plan
}

func normalizePlanSteps(plan *Plan) {
	for i := range plan.Steps {
		step := &plan.Steps[i]
		if step.ID == "" && step.Name != "" {
			step.ID = strings.ToLower(strings.TrimSpace(step.Name))
		}
		if step.Title == "" && step.Name != "" {
			step.Title = step.Name
		}
		if step.Description == "" && step.Detail != "" {
			step.Description = step.Detail
		}
		if step.ID == "" {
			step.ID = fmt.Sprintf("step_%d", i+1)
		}
		if step.Title == "" {
			step.Title = step.ID
		}
		step.Name = ""
		step.Detail = ""
	}
}

func stepDescription(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case map[string]any:
		for _, key := range []string{"description", "action", "summary"} {
			if text, ok := typed[key].(string); ok && text != "" {
				return text
			}
		}
		data, _ := json.Marshal(typed)
		return string(data)
	default:
		data, _ := json.Marshal(typed)
		return string(data)
	}
}

func truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCapsule(row scanner) (Capsule, error) {
	var capsule Capsule
	var sessionID, providerID, model, summary sql.NullString
	var budgetJSON, allowedJSON, forbiddenJSON, checksJSON, planJSON string
	if err := row.Scan(&capsule.ID, &sessionID, &capsule.Goal, &capsule.Mode, &capsule.Risk,
		&capsule.Status, &budgetJSON, &allowedJSON, &forbiddenJSON, &checksJSON,
		&providerID, &model, &summary, &planJSON, &capsule.CreatedAt, &capsule.UpdatedAt); err != nil {
		return Capsule{}, err
	}
	capsule.SessionID = sessionID.String
	capsule.Provider = providerID.String
	capsule.Model = model.String
	capsule.Summary = summary.String
	_ = json.Unmarshal([]byte(budgetJSON), &capsule.Budget)
	_ = json.Unmarshal([]byte(allowedJSON), &capsule.AllowedTools)
	_ = json.Unmarshal([]byte(forbiddenJSON), &capsule.ForbiddenTools)
	_ = json.Unmarshal([]byte(checksJSON), &capsule.SuccessChecks)
	_ = json.Unmarshal([]byte(planJSON), &capsule.Plan)
	return capsule, nil
}

func (s *Service) insertRun(ctx context.Context, capsule Capsule) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO harness_runs (
  id, session_id, goal, mode, risk, status, budget_json, allowed_tools_json,
  forbidden_tools_json, success_checks_json, provider, model, summary, plan_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		capsule.ID, nullable(capsule.SessionID), capsule.Goal, capsule.Mode, capsule.Risk, capsule.Status,
		mustJSON(capsule.Budget), mustJSON(capsule.AllowedTools), mustJSON(capsule.ForbiddenTools),
		mustJSON(capsule.SuccessChecks), nullable(capsule.Provider), nullable(capsule.Model),
		nullable(capsule.Summary), "{}", capsule.CreatedAt, capsule.UpdatedAt)
	return err
}

func (s *Service) updateRunPlan(ctx context.Context, capsule Capsule) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
UPDATE harness_runs
SET status = ?, provider = ?, model = ?, summary = ?, plan_json = ?, updated_at = ?
WHERE id = ?`,
		capsule.Status, nullable(capsule.Provider), nullable(capsule.Model), nullable(capsule.Summary),
		mustJSON(capsule.Plan), now, capsule.ID)
	return err
}

func (s *Service) addEvent(ctx context.Context, runID, eventType, payload string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO harness_events (run_id, type, payload, created_at)
VALUES (?, ?, ?, ?)`, runID, eventType, payload, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (s *Service) events(ctx context.Context, runID string) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, run_id, type, payload, created_at
FROM harness_events
WHERE run_id = ?
ORDER BY id ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.RunID, &event.Type, &event.Payload, &event.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func nullable(value string) sql.NullString {
	return sql.NullString{String: value, Valid: strings.TrimSpace(value) != ""}
}

func newID(prefix string) string {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(data[:])
}
