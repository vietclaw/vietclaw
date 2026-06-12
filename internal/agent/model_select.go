package agent

import (
	"context"
	"strings"

	"vietclaw/internal/providers"
	"vietclaw/internal/router"
)

func (s *Service) resolveSessionCatalogID(ctx context.Context, sessionID string) string {
	if sessionID == "" || s.db == nil {
		return ""
	}
	var catalogID string
	err := s.db.QueryRowContext(ctx, `SELECT preferred_catalog_id FROM sessions WHERE id = ?`, sessionID).Scan(&catalogID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(catalogID)
}

func (s *Service) applyRequestModel(req ChatRequest) (providerID, modelID string) {
	if catalogID := strings.TrimSpace(req.CatalogID); catalogID != "" {
		if entry, ok := s.cfg.CatalogEntry(catalogID); ok {
			return entry.Provider, entry.Model
		}
	}
	if provider := strings.TrimSpace(req.Provider); provider != "" {
		model := strings.TrimSpace(req.Model)
		if model == "" {
			for _, p := range s.cfg.Providers {
				if p.ID == provider && p.DefaultModel != "" {
					model = p.DefaultModel
					break
				}
			}
		}
		return provider, model
	}
	return "", ""
}

func (s *Service) resolveChildModel(parentReq ChatRequest, modelHint string) (providerID, modelID string) {
	hint := strings.TrimSpace(modelHint)
	if hint == "" || hint == "inherit" {
		if parentReq.ParentProvider != "" {
			return parentReq.ParentProvider, parentReq.ParentModel
		}
		if p, m := s.applyRequestModel(parentReq); p != "" {
			return p, m
		}
		return "", ""
	}
	if entry, ok := s.cfg.CatalogEntry(hint); ok {
		return entry.Provider, entry.Model
	}
	if parts := strings.SplitN(hint, "/", 2); len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", ""
}

func (s *Service) enrichRequestModel(ctx context.Context, req ChatRequest) ChatRequest {
	if strings.TrimSpace(req.CatalogID) == "" {
		req.CatalogID = s.resolveSessionCatalogID(ctx, req.SessionID)
	}
	return req
}

func (s *Service) selectLoopProvider(ctx context.Context, req ChatRequest, chatReq providers.ChatRequest, excluded []string) (router.Selection, []string, error) {
	allowed := s.profile(req.AgentID).Providers

	if providerID, modelID := s.applyRequestModel(req); providerID != "" {
		selection, err := s.router.SelectExplicit(ctx, chatReq, excluded, allowed, providerID, modelID)
		if err == nil {
			return selection, excluded, nil
		}
	}

	selection, err := s.router.SelectForProfile(ctx, chatReq, excluded, allowed)
	return selection, excluded, err
}

func (s *Service) agentModelHint(agentID string) string {
	if s.agents == nil {
		return "inherit"
	}
	def, ok := s.agents.Get(agentID)
	if !ok || strings.TrimSpace(def.Model) == "" {
		return "inherit"
	}
	return def.Model
}
