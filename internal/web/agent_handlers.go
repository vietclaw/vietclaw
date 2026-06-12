package web

import (
	"net/http"

	"vietclaw/internal/app"
)

func handleAgents(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if application.Agent == nil || application.Agent.AgentRegistry() == nil {
			writeJSON(w, http.StatusOK, map[string]any{"agents": []any{}})
			return
		}
		list := application.Agent.AgentRegistry().List()
		out := make([]map[string]any, 0, len(list))
		for _, def := range list {
			out = append(out, map[string]any{
				"id":          def.ID,
				"name":        def.Name,
				"language":    def.Language,
				"tools":       def.Tools,
				"providers":   def.Providers,
				"model":       def.Model,
				"memory_scope": def.MemoryScope,
				"max_steps":   def.MaxSteps,
				"spawnable":   def.Spawnable,
				"auto_create": def.AutoCreate,
				"skills":      len(def.Skills),
				"tool_guides": len(def.ToolGuides),
				"custom_tools": len(def.CustomTools),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"agents": out})
	}
}

func handleAgentDetail(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if application.Agent == nil || application.Agent.AgentRegistry() == nil {
			writeError(w, http.StatusNotFound, "agent not found")
			return
		}
		def, ok := application.Agent.AgentRegistry().Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "agent not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":          def.ID,
			"name":        def.Name,
			"language":    def.Language,
			"persona":     def.Persona,
			"tools":       def.Tools,
			"providers":   def.Providers,
			"model":       def.Model,
			"memory_scope": def.MemoryScope,
			"max_steps":   def.MaxSteps,
			"spawnable":   def.Spawnable,
			"auto_create": def.AutoCreate,
		})
	}
}

func handleAgentsReload(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if application.Agent == nil {
			writeError(w, http.StatusInternalServerError, "agent service unavailable")
			return
		}
		if err := application.Agent.ReloadAgents(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}
