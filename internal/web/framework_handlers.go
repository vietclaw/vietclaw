package web

import (
	"net/http"

	"vietclaw/internal/app"
	"vietclaw/internal/channels"
	"vietclaw/internal/framework"
	"vietclaw/internal/plugins"
)

func handleFramework(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		hooks := 0
		if application.Framework != nil && application.Framework.Hooks != nil {
			hooks = application.Framework.Hooks.Count()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"enabled":          application.Config.Framework.Enabled,
			"delegate_enabled": application.Config.Framework.DelegateEnabled,
			"hooks_enabled":    application.Config.Framework.HooksEnabled,
			"hooks_registered": hooks,
			"agents":           application.Config.Agents,
			"extensions":       plugins.BuiltinRegistry(),
			"channels":         channels.RegisteredAdapters(),
		})
	}
}

func handleFrameworkExtensions(_ *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, framework.BuiltinExtensions())
	}
}
