package web

import (
	"encoding/json"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"vietclaw/internal/agent"
	"vietclaw/internal/app"
	"vietclaw/internal/channels"
	"vietclaw/internal/db"
	"vietclaw/internal/memory"
	"vietclaw/internal/providers"
	"vietclaw/internal/router"
)

func NewRouter(application *app.App) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /status", handleStatus(application))
	mux.HandleFunc("GET /logs/recent", handleRecentLogs(application))
	mux.HandleFunc("GET /api/logs/recent", handleRecentLogs(application))
	mux.HandleFunc("POST /api/chat", handleAPIChat(application))
	mux.HandleFunc("GET /api/memory", handleMemoryList(application))
	mux.HandleFunc("POST /api/memory", handleMemoryAdd(application))
	mux.HandleFunc("GET /api/memory/search", handleMemorySearch(application))
	mux.HandleFunc("GET /api/sessions", handleSessions(application))
	mux.HandleFunc("GET /api/sessions/{id}", handleSessionDetail(application))
	mux.HandleFunc("GET /api/costs/today", handleCostsToday(application))
	mux.HandleFunc("GET /api/budget", handleBudget(application))
	mux.HandleFunc("GET /api/providers", handleProviders(application))
	mux.HandleFunc("GET /api/channels", handleChannels(application))
	mux.HandleFunc("POST /api/channels/discord/test", handleDiscordTest(application))
	mux.HandleFunc("POST /api/channels/telegram/test", handleTelegramTest(application))
	mux.HandleFunc("GET /", handleStatic(application))
	return mux
}

func handleStatic(application *app.App) http.HandlerFunc {
	dist, err := fs.Sub(webDist, "dist")
	if err != nil {
		application.Logger.Printf("load embedded web dist: %v", err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "." || name == "" {
			name = "index.html"
		}
		if file, err := dist.Open(name); err == nil {
			defer file.Close()
			if info, statErr := file.Stat(); statErr == nil && !info.IsDir() {
				serveEmbeddedFile(w, r, name, file, info.ModTime())
				return
			}
		}
		file, err := dist.Open("index.html")
		if err != nil {
			http.Error(w, "web UI is not available", http.StatusServiceUnavailable)
			return
		}
		defer file.Close()
		serveEmbeddedFile(w, r, "index.html", file, time.Now())
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func handleStatus(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"version":              application.Version.Version,
			"commit":               application.Version.Commit,
			"uptime":               time.Since(application.StartTime).Round(time.Second).String(),
			"db_ok":                db.Check(application.DB),
			"mode":                 application.Config.Runtime.Mode,
			"max_concurrent_tasks": application.Config.Runtime.MaxConcurrentTasks,
		})
	}
}

func handleRecentLogs(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		lines, err := recentLines(application.LogFile, 80)
		if err != nil {
			application.Logger.Printf("read recent logs: %v", err)
			writeJSON(w, http.StatusOK, []string{})
			return
		}
		writeJSON(w, http.StatusOK, lines)
	}
}

func handleAPIChat(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req agent.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
			return
		}
		resp, err := application.Agent.Chat(r.Context(), req)
		if err != nil {
			resp.OK = false
			if resp.Error == "" {
				resp.Error = err.Error()
			}
			writeJSON(w, http.StatusBadRequest, resp)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func handleMemoryList(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scope := r.URL.Query().Get("scope")
		records, err := application.Agent.Memory().List(r.Context(), scope, 100)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, records)
	}
}

func handleMemoryAdd(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Scope      string `json:"scope"`
			Kind       string `json:"kind"`
			Content    string `json:"content"`
			Confidence string `json:"confidence"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
			return
		}
		rec, err := application.Agent.Memory().Add(r.Context(), memory.Record{
			Scope:      req.Scope,
			Kind:       memory.Kind(req.Kind),
			Content:    req.Content,
			Confidence: memory.Confidence(req.Confidence),
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "memory": rec})
	}
}

func handleMemorySearch(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		records, err := application.Agent.Memory().Search(r.Context(), r.URL.Query().Get("scope"), r.URL.Query().Get("q"), 50)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, records)
	}
}

func handleSessions(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := application.Agent.Sessions(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, sessions)
	}
}

func handleSessionDetail(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		detail, err := application.Agent.SessionMessages(r.Context(), r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "session not found"})
			return
		}
		writeJSON(w, http.StatusOK, detail)
	}
}

func handleCostsToday(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"total_cost_usd": router.TodayCost(r.Context(), application.DB)})
	}
}

func handleBudget(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"total_cost_usd":             router.TodayCost(r.Context(), application.DB),
			"daily_usd_limit":            application.Config.Budget.DailyUSDLimit,
			"require_approval_above_usd": application.Config.Budget.RequireApprovalAboveUSD,
			"cheap_first":                application.Config.Router.CheapFirst,
			"allow_escalation":           application.Config.Router.AllowEscalation,
		})
	}
}

func handleProviders(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, providers.Redact(application.Config.Providers))
	}
}

func handleChannels(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if application.Channels != nil {
			writeJSON(w, http.StatusOK, application.Channels.Statuses())
			return
		}
		writeJSON(w, http.StatusOK, channels.StatusFromConfig(application.Config))
	}
}

func handleDiscordTest(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, channelEnvStatus("discord", application.Config.Channels.Discord.Enabled, application.Config.Channels.Discord.TokenEnv))
	}
}

func handleTelegramTest(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, channelEnvStatus("telegram", application.Config.Channels.Telegram.Enabled, application.Config.Channels.Telegram.TokenEnv))
	}
}

func channelEnvStatus(name string, enabled bool, tokenEnv string) map[string]any {
	_, ok := os.LookupEnv(tokenEnv)
	return map[string]any{
		"name":      name,
		"enabled":   enabled,
		"token_env": tokenEnv,
		"env_found": ok,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func recentLines(path string, maxLines int) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	text := strings.TrimRight(string(data), "\r\n")
	if text == "" {
		return []string{}, nil
	}

	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return lines, nil
	}
	return lines[len(lines)-maxLines:], nil
}

func serveEmbeddedFile(w http.ResponseWriter, r *http.Request, name string, file fs.File, modTime time.Time) {
	if strings.HasPrefix(name, "_nuxt/") || strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	reader, ok := file.(interface {
		Read([]byte) (int, error)
		Seek(int64, int) (int64, error)
	})
	if !ok {
		http.Error(w, "embedded file is not seekable", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, name, modTime, reader)
}
