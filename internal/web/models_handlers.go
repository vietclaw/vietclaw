package web

import (
	"encoding/json"
	"net/http"

	"vietclaw/internal/app"
)

func handleModelsCatalog(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"catalog":            application.Config.EnabledCatalog(),
			"default_catalog_id": application.Config.Models.DefaultCatalogID,
		})
	}
}

func handleSessionModel(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")
		if sessionID == "" {
			writeError(w, http.StatusBadRequest, "session id required")
			return
		}
		switch r.Method {
		case http.MethodPut:
			var body struct {
				CatalogID string `json:"catalog_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json")
				return
			}
			if body.CatalogID != "" {
				if _, ok := application.Config.CatalogEntry(body.CatalogID); !ok {
					writeError(w, http.StatusBadRequest, "catalog entry not found or disabled")
					return
				}
			}
			_, err := application.DB.Exec(`UPDATE sessions SET preferred_catalog_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, body.CatalogID, sessionID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "catalog_id": body.CatalogID})
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}
