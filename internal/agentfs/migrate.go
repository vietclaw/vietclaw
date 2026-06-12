package agentfs

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"vietclaw/internal/config"
)

func MigrateFromConfig(db *sql.DB, paths config.Paths, cfg *config.Config) (bool, error) {
	if db != nil {
		var value string
		err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, MigrationKey).Scan(&value)
		if err == nil && value == "1" {
			cfg.Agents = nil
			return false, nil
		}
	}

	root := DefaultRoot(paths.DataDir)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return false, fmt.Errorf("create agents dir: %w", err)
	}

	migrated := false
	for _, profile := range cfg.Agents {
		if profile.ID == "" {
			continue
		}
		agentPath := filepath.Join(root, profile.ID, AgentFileName)
		if _, err := os.Stat(agentPath); err == nil {
			continue
		}
		req := CreateRequest{
			ID:          profile.ID,
			Name:        profile.Name,
			Language:    profile.Language,
			Persona:     profile.Persona,
			Tools:       profile.Tools,
			Providers:   profile.Providers,
			Model:       "inherit",
			MemoryScope: profile.MemoryScope,
			MaxSteps:    profile.MaxSteps,
			Spawnable:   profile.ID != config.DefaultAgentID,
		}
		if profile.ID == config.DefaultAgentID {
			req.Spawnable = false
		}
		if err := WriteAgent(agentPath, req); err != nil {
			return migrated, fmt.Errorf("migrate agent %s: %w", profile.ID, err)
		}
		migrated = true
	}

	if len(cfg.Agents) == 0 {
		defaultPath := filepath.Join(root, config.DefaultAgentID, AgentFileName)
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			if err := WriteAgent(defaultPath, CreateRequest{
				ID:        config.DefaultAgentID,
				Name:      cfg.Agent.Name,
				Language:  cfg.Agent.Language,
				Persona:   "",
				Tools:     []string{},
				Providers: []string{},
				Model:     "inherit",
				Spawnable: false,
			}); err != nil {
				return migrated, err
			}
			migrated = true
		}
	}

	cfg.Agents = nil
	if db != nil {
		_, _ = db.Exec(`INSERT INTO settings(key, value, updated_at) VALUES(?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
			MigrationKey, "1")
	}
	return migrated, nil
}
