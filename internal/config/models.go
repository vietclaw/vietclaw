package config

import "strings"

func (cfg Config) CatalogEntry(id string) (CatalogModelConfig, bool) {
	id = strings.TrimSpace(id)
	for _, entry := range cfg.Models.Catalog {
		if entry.ID == id && entry.Enabled {
			return entry, true
		}
	}
	return CatalogModelConfig{}, false
}

func (cfg Config) EnabledCatalog() []CatalogModelConfig {
	out := make([]CatalogModelConfig, 0, len(cfg.Models.Catalog))
	for _, entry := range cfg.Models.Catalog {
		if entry.Enabled {
			out = append(out, entry)
		}
	}
	return out
}

func (cfg Config) DefaultCatalogEntry() (CatalogModelConfig, bool) {
	if entry, ok := cfg.CatalogEntry(cfg.Models.DefaultCatalogID); ok {
		return entry, true
	}
	for _, entry := range cfg.EnabledCatalog() {
		return entry, true
	}
	return CatalogModelConfig{}, false
}
