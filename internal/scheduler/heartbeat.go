package scheduler

import (
	"context"
	"log"
	"time"

	"vietclaw/internal/agent"
	"vietclaw/internal/config"
)

// Heartbeat runs periodic proactive agent checks (OpenClaw-style heartbeat polling).
type Heartbeat struct {
	cfg    config.Config
	agent  *agent.Service
	logger *log.Logger
}

func NewHeartbeat(cfg config.Config, service *agent.Service, logger *log.Logger) *Heartbeat {
	return &Heartbeat{cfg: cfg, agent: service, logger: logger}
}

func (h *Heartbeat) Start(ctx context.Context) {
	if !h.cfg.Agent.Heartbeat.Enabled || h.agent == nil {
		return
	}
	interval := h.cfg.Agent.Heartbeat.IntervalSeconds
	if interval <= 0 {
		interval = config.DefaultHeartbeatIntervalSec
	}
	sessionID := h.cfg.Agent.Heartbeat.SessionID
	if sessionID == "" {
		sessionID = config.DefaultHeartbeatSessionID
	}
	userID := h.cfg.Agent.Heartbeat.UserID
	if userID == "" {
		userID = config.DefaultHeartbeatUserID
	}
	prompt := h.cfg.Agent.Heartbeat.Prompt
	if prompt == "" {
		prompt = config.DefaultHeartbeatPrompt
	}

	h.logf("heartbeat enabled interval=%ds session=%s", interval, sessionID)

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.runOnce(ctx, sessionID, userID, prompt)
			}
		}
	}()
}

func (h *Heartbeat) runOnce(ctx context.Context, sessionID, userID, prompt string) {
	runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	resp, err := h.agent.Chat(runCtx, agent.ChatRequest{
		SessionID: sessionID,
		UserID:    userID,
		Channel:   "heartbeat",
		Message:   prompt,
		Mode:      h.cfg.Runtime.Mode,
	})
	if err != nil {
		h.logf("heartbeat error: %v", err)
		return
	}
	if resp.Reply != "" {
		h.logf("heartbeat reply session=%s provider=%s: %s", sessionID, resp.Provider, truncate(resp.Reply, 120))
	}
}

func (h *Heartbeat) logf(format string, args ...any) {
	if h.logger != nil {
		h.logger.Printf(format, args...)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
