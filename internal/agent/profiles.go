package agent

import (
	"context"
	"strings"

	"vietclaw/internal/config"
	"vietclaw/internal/providers"
)

func (s *Service) profile(agentID string) config.AgentProfileConfig {
	if s.agents == nil {
		return config.AgentProfileConfig{ID: config.DefaultAgentID}
	}
	def, ok := s.agents.Get(agentID)
	if !ok {
		return config.AgentProfileConfig{
			ID:       config.DefaultAgentID,
			Name:     s.cfg.Agent.Name,
			Language: s.cfg.Agent.Language,
		}
	}
	return def.Profile()
}

func (s *Service) memoryScope(req ChatRequest) string {
	profile := s.profile(req.AgentID)
	if profile.MemoryScope != "" {
		return profile.MemoryScope + ":user:" + req.UserID
	}
	if profile.ID != "" && profile.ID != config.DefaultAgentID {
		return "agent:" + profile.ID + ":user:" + req.UserID
	}
	return "user:" + req.UserID
}

func (s *Service) selectDelegateAgent(message string) string {
	if s.agents == nil {
		return ""
	}
	return s.router.SelectAgent(context.Background(), message, s.Language(), s.agents.Profiles())
}

func (s *Service) applyAgentProfile(ctx context.Context, req ChatRequest) ChatRequest {
	if s.agents == nil {
		return req
	}
	if delegate := s.router.SelectAgent(ctx, req.Message, s.Language(), s.agents.Profiles()); delegate != "" {
		req.AgentID = delegate
	}
	return req
}

func (s *Service) profileLanguage(agentID string) string {
	profile := s.profile(agentID)
	if strings.TrimSpace(profile.Language) != "" {
		return profile.Language
	}
	return s.cfg.Agent.Language
}

func (s *Service) profileMaxSteps(agentID string) int {
	profile := s.profile(agentID)
	if profile.MaxSteps > 0 {
		return profile.MaxSteps
	}
	return s.maxAgentSteps()
}

func (s *Service) applyProfilePersona(req ChatRequest, messages []providers.Message) []providers.Message {
	profile := s.profile(req.AgentID)
	if strings.TrimSpace(profile.Persona) == "" || len(messages) == 0 || messages[0].Role != "system" {
		return messages
	}
	messages[0].Content += "\n\nAgent persona:\n" + profile.Persona
	return s.applyToolGuides(req.AgentID, messages)
}

func (s *Service) applyToolGuides(agentID string, messages []providers.Message) []providers.Message {
	if s.agents == nil || len(messages) == 0 || messages[0].Role != "system" {
		return messages
	}
	guides := s.agents.ToolGuidesFor(agentID)
	if len(guides) == 0 {
		return messages
	}
	lines := []string{"Tool guides:"}
	for _, guide := range guides {
		lines = append(lines, "- "+guide.Tool+": "+guide.Instructions)
	}
	messages[0].Content += "\n\n" + strings.Join(lines, "\n")
	return messages
}
