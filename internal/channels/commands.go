package channels

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"vietclaw/internal/config"
)

type CommandResult struct {
	Handled bool
	Reply   string
}

func HandleModelCommand(ctx context.Context, db *sql.DB, cfg config.Config, platform, scopeID, text string) (CommandResult, error) {
	mode := "slash"
	prefix := "/"
	if platform == PlatformTelegram {
		mode = cfg.Channels.Telegram.CommandMode
		prefix = cfg.Channels.Telegram.CommandPrefix
	}
	cmd, arg := parseCommand(text, mode, prefix)
	if cmd != "models" {
		return CommandResult{}, nil
	}

	catalog := cfg.EnabledCatalog()
	if len(catalog) == 0 {
		return CommandResult{Handled: true, Reply: "No models configured."}, nil
	}

	if arg == "" || arg == "list" {
		lines := []string{"Available models:"}
		current := readCatalogPreference(db, platform, scopeID)
		for _, entry := range catalog {
			marker := " "
			if entry.ID == current {
				marker = "*"
			}
			label := entry.Label
			if label == "" {
				label = entry.ID
			}
			lines = append(lines, fmt.Sprintf("%s %s — %s/%s", marker, label, entry.Provider, entry.Model))
		}
		lines = append(lines, "", "Use /models <id> to select.")
		return CommandResult{Handled: true, Reply: strings.Join(lines, "\n")}, nil
	}

	if _, ok := cfg.CatalogEntry(arg); !ok {
		return CommandResult{Handled: true, Reply: fmt.Sprintf("Unknown model id: %s", arg)}, nil
	}
	if err := writeCatalogPreference(db, platform, scopeID, arg); err != nil {
		return CommandResult{}, err
	}
	entry, _ := cfg.CatalogEntry(arg)
	label := entry.Label
	if label == "" {
		label = entry.ID
	}
	return CommandResult{Handled: true, Reply: fmt.Sprintf("Model set to %s (%s/%s)", label, entry.Provider, entry.Model)}, nil
}

func CatalogPreferenceKey(platform, scopeID string) string {
	return platform + ":" + scopeID + ":catalog_id"
}

func ReadCatalogPreference(db *sql.DB, platform, scopeID string) string {
	return readCatalogPreference(db, platform, scopeID)
}

func readCatalogPreference(db *sql.DB, platform, scopeID string) string {
	if db == nil {
		return ""
	}
	var value string
	_ = db.QueryRow(`SELECT value FROM settings WHERE key = ?`, CatalogPreferenceKey(platform, scopeID)).Scan(&value)
	return strings.TrimSpace(value)
}

func writeCatalogPreference(db *sql.DB, platform, scopeID, catalogID string) error {
	if db == nil {
		return fmt.Errorf("database unavailable")
	}
	_, err := db.Exec(`INSERT INTO settings(key, value, updated_at) VALUES(?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
		CatalogPreferenceKey(platform, scopeID), catalogID)
	return err
}

func ResolveCatalogID(db *sql.DB, cfg config.Config, platform, scopeID, sessionID string) string {
	if db == nil {
		return cfg.Models.DefaultCatalogID
	}
	if sessionID != "" {
		var catalogID string
		if db.QueryRow(`SELECT preferred_catalog_id FROM sessions WHERE id = ?`, sessionID).Scan(&catalogID) == nil {
			catalogID = strings.TrimSpace(catalogID)
			if catalogID != "" {
				if _, ok := cfg.CatalogEntry(catalogID); ok {
					return catalogID
				}
			}
		}
	}
	if pref := readCatalogPreference(db, platform, scopeID); pref != "" {
		if _, ok := cfg.CatalogEntry(pref); ok {
			return pref
		}
	}
	return cfg.Models.DefaultCatalogID
}

func parseCommand(text, mode, prefix string) (cmd, arg string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ""
	}
	if mode == "prefix" {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			prefix = "/"
		}
		if !strings.HasPrefix(text, prefix) {
			return "", ""
		}
		text = strings.TrimSpace(strings.TrimPrefix(text, prefix))
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "", ""
	}
	cmd = strings.ToLower(strings.TrimPrefix(fields[0], "/"))
	if len(fields) > 1 {
		arg = strings.Join(fields[1:], " ")
	}
	return cmd, arg
}
