package router

import "vietclaw/internal/providers"

func (r *ModelRouter) defaultProvider() providers.Provider {
	for _, p := range r.providers {
		if p.ID() == r.cfg.Router.DefaultProvider {
			return p
		}
	}
	if r.cfg.Router.CheapFirst {
		for _, p := range r.providers {
			if p.Type() == providers.TypeMock {
				return p
			}
		}
	}
	return r.providers[0]
}

func (r *ModelRouter) defaultModel(provider providers.Provider) string {
	for _, cfg := range r.cfg.Providers {
		if cfg.ID == provider.ID() && cfg.DefaultModel != "" {
			return cfg.DefaultModel
		}
	}
	if r.cfg.Router.DefaultModel != "" {
		return r.cfg.Router.DefaultModel
	}
	return providers.DefaultMockModel
}
