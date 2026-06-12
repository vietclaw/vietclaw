package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"vietclaw/internal/agentfs"
	"vietclaw/internal/framework"
	"vietclaw/internal/tools"
)

type spawnNotifier func(agentID, status, summary, childSessionID, parentSessionID string)

func (s *Service) handleFrameworkTool(ctx context.Context, parentReq ChatRequest, parentRunID, toolName, argsJSON string, notify spawnNotifier) (string, error) {
	switch strings.TrimSpace(toolName) {
	case tools.ToolAgentDelegate:
		return s.handleAgentSpawn(ctx, parentReq, parentRunID, argsJSON, true, notify)
	case tools.ToolAgentSpawn:
		return s.handleAgentSpawn(ctx, parentReq, parentRunID, argsJSON, true, notify)
	case tools.ToolAgentSpawnBatch:
		return s.handleAgentSpawnBatch(ctx, parentReq, parentRunID, argsJSON, notify)
	case tools.ToolAgentCreate:
		return s.handleAgentCreate(ctx, argsJSON)
	default:
		return "", fmt.Errorf("unknown framework tool: %s", toolName)
	}
}

func (s *Service) handleAgentSpawn(ctx context.Context, parentReq ChatRequest, parentRunID, argsJSON string, wait bool, notify spawnNotifier) (string, error) {
	var args struct {
		AgentID string `json:"agent_id"`
		Message string `json:"message"`
		Model   string `json:"model"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid spawn args: %w", err)
	}
	agentID := strings.TrimSpace(args.AgentID)
	message := strings.TrimSpace(args.Message)
	if agentID == "" || message == "" {
		return "", fmt.Errorf("agent_id and message are required")
	}
	def, ok := s.agents.Get(agentID)
	if !ok {
		return "", fmt.Errorf("agent not found: %s", agentID)
	}
	if !def.Spawnable {
		return "", fmt.Errorf("agent is not spawnable: %s", agentID)
	}

	modelHint := strings.TrimSpace(args.Model)
	if modelHint == "" {
		modelHint = s.agentModelHint(agentID)
	}

	spawnPrefix, childSessionID := buildChildSessionIDs(parentReq.SessionID, agentID)
	parentSessionID := parentReq.SessionID

	run := func() (ChatResponse, error) {
		if err := s.pool.Acquire(ctx, parentRunID); err != nil {
			return ChatResponse{}, err
		}
		defer s.pool.Release(parentRunID)
		childReq := parentReq
		childReq.AgentID = agentID
		childReq.Message = message
		childReq.SessionID = spawnPrefix
		if providerID, modelID := s.resolveChildModel(parentReq, modelHint); providerID != "" {
			childReq.Provider = providerID
			childReq.Model = modelID
		}
		return s.Delegate(ctx, childReq, parentRunID, agentID, message)
	}

	emit := func(status, summary string) {
		if notify != nil {
			notify(agentID, status, summary, childSessionID, parentSessionID)
		}
	}

	if wait {
		emit("running", message)
		resp, err := run()
		if err != nil {
			emit("failed", err.Error())
			return "", err
		}
		emit("done", resp.Reply)
		return fmt.Sprintf("Spawned %s: %s", agentID, resp.Reply), nil
	}

	go func() {
		emit("running", message)
		resp, err := run()
		if err != nil {
			emit("failed", err.Error())
			return
		}
		emit("done", resp.Reply)
		_ = resp
	}()
	return fmt.Sprintf("Spawned %s asynchronously", agentID), nil
}

func (s *Service) handleAgentSpawnBatch(ctx context.Context, parentReq ChatRequest, parentRunID, argsJSON string, notify spawnNotifier) (string, error) {
	var args struct {
		Tasks []struct {
			AgentID string `json:"agent_id"`
			Message string `json:"message"`
			Model   string `json:"model"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid spawn batch args: %w", err)
	}
	if len(args.Tasks) == 0 {
		return "", fmt.Errorf("tasks are required")
	}

	type result struct {
		AgentID string
		Reply   string
		Error   string
	}
	results := make([]result, len(args.Tasks))
	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)

	for i, task := range args.Tasks {
		i, task := i, task
		g.Go(func() error {
			payload, _ := json.Marshal(map[string]string{
				"agent_id": task.AgentID,
				"message":  task.Message,
				"model":    task.Model,
			})
			text, err := s.handleAgentSpawn(gctx, parentReq, parentRunID, string(payload), true, notify)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results[i] = result{AgentID: task.AgentID, Error: err.Error()}
				return nil
			}
			results[i] = result{AgentID: task.AgentID, Reply: text}
			return nil
		})
	}
	_ = g.Wait()

	var lines []string
	for _, item := range results {
		if item.Error != "" {
			lines = append(lines, fmt.Sprintf("- %s: ERROR %s", item.AgentID, item.Error))
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", item.AgentID, item.Reply))
	}
	return strings.Join(lines, "\n"), nil
}

func (s *Service) handleAgentCreate(ctx context.Context, argsJSON string) (string, error) {
	if !s.cfg.Framework.AllowAutoCreate {
		return "", fmt.Errorf("agent auto-create is disabled in settings")
	}
	var args struct {
		ID          string                `json:"id"`
		Name        string                `json:"name"`
		Language    string                `json:"language"`
		Persona     string                `json:"persona"`
		Tools       []string              `json:"tools"`
		Providers   []string              `json:"providers"`
		Model       string                `json:"model"`
		MemoryScope string                `json:"memory_scope"`
		MaxSteps    int                   `json:"max_steps"`
		Spawnable   bool                  `json:"spawnable"`
		Skills      []agentfs.SkillInput    `json:"skills"`
		ToolGuides  []agentfs.ToolGuideInput `json:"tool_guides"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid create args: %w", err)
	}
	if err := agentfs.ValidateAgentID(args.ID); err != nil {
		return "", err
	}
	if s.agents.Count() >= s.cfg.Framework.MaxTotalAgents {
		return "", fmt.Errorf("max total agents reached (%d)", s.cfg.Framework.MaxTotalAgents)
	}

	req := agentfs.CreateRequest{
		ID:          args.ID,
		Name:        args.Name,
		Language:    args.Language,
		Persona:     args.Persona,
		Tools:       args.Tools,
		Providers:   args.Providers,
		Model:       args.Model,
		MemoryScope: args.MemoryScope,
		MaxSteps:    args.MaxSteps,
		Spawnable:   args.Spawnable,
		Skills:      args.Skills,
		ToolGuides:  args.ToolGuides,
	}
	if req.Name == "" {
		req.Name = req.ID
	}
	if req.Language == "" {
		req.Language = s.cfg.Agent.Language
	}
	if req.Model == "" {
		req.Model = "inherit"
	}
	if !req.Spawnable {
		req.Spawnable = true
	}

	dir, err := agentfs.CreateAgent(s.agents.Root(), req)
	if err != nil {
		return "", err
	}
	if err := s.agents.Reload(); err != nil {
		return "", err
	}
	s.tools.ReloadAgentTools(s.agents)

	if s.framework != nil && s.framework.Hooks != nil {
		_ = s.framework.Hooks.Emit(ctx, framework.EventRunStart, framework.HookContext{
			AgentID: args.ID,
			Message: "agent_create",
		})
	}
	return fmt.Sprintf("Created agent %s at %s", args.ID, dir), nil
}
