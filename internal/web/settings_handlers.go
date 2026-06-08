package web

import (
	"encoding/json"
	"net/http"

	"vietclaw/internal/agent"
	"vietclaw/internal/app"
	"vietclaw/internal/config"
)

func handleSettings(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, application.Config)
		case http.MethodPut:
			var cfg config.Config
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json")
				return
			}
			cfg = config.MergeDefault(cfg, application.Config)
			if err := config.Validate(cfg); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if application.ConfigFile != "" {
				if err := config.Save(application.ConfigFile, cfg); err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
			}
			applyConfig(application, cfg)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "config": application.Config})
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

func handleSettingsReload(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if application.ConfigFile == "" {
			writeError(w, http.StatusBadRequest, "config file is not configured")
			return
		}
		cfg, err := config.Load(application.ConfigFile)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		cfg = config.MergeDefault(cfg, application.Config)
		if err := config.Validate(cfg); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		applyConfig(application, cfg)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "config": application.Config})
	}
}

func applyConfig(application *app.App, cfg config.Config) {
	application.Config = cfg
	application.Agent = agent.NewService(cfg, application.DB)
}
