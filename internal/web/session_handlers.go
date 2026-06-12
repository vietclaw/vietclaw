package web

import (
	"fmt"
	"net/http"
	"strconv"

	"vietclaw/internal/agent"
	"vietclaw/internal/app"
)

func handleSessions(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := application.Agent.Sessions(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sessions)
	}
}

func handleSessionDetail(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		detail, err := application.Agent.SessionMessages(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeJSON(w, http.StatusOK, detail)
	}
}

func handleSessionChildren(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		children, err := application.Agent.SessionChildren(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, children)
	}
}

func handleSessionWatch(application *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, http.StatusInternalServerError, "streaming not supported")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher.Flush()

		status, summary := application.Agent.SessionRunStatus(r.Context(), sessionID)
		writeSSEJSON(w, "run_status", map[string]string{"status": status, "summary": summary})
		flusher.Flush()

		if afterID := parseAfterToolEventID(r.URL.Query().Get("after")); afterID > 0 {
			events, err := application.Agent.SessionToolEventsAfter(r.Context(), sessionID, afterID)
			if err == nil {
				for _, item := range events {
					writeSSEJSON(w, "tool_call", map[string]string{
						"name":  item.ToolName,
						"input": item.Input,
					})
					writeSSEJSON(w, "tool_result", map[string]string{
						"name":   item.ToolName,
						"result": item.Output,
					})
				}
				flusher.Flush()
			}
		}

		if isTerminalRunStatus(status) {
			fmt.Fprintf(w, "event: done\ndata: [DONE]\n\n")
			flusher.Flush()
			return
		}

		ch, unsub := application.Agent.SubscribeSession(sessionID)
		defer unsub()

		for {
			select {
			case <-r.Context().Done():
				return
			case ev, ok := <-ch:
				if !ok {
					return
				}
				switch ev.Event {
				case "tool_call":
					writeSSEJSON(w, "tool_call", map[string]string{
						"name":  ev.ToolName,
						"input": ev.ToolInput,
					})
				case "tool_result":
					writeSSEJSON(w, "tool_result", map[string]string{
						"name":   ev.ToolName,
						"result": ev.ToolResult,
					})
				case "text":
					if ev.Text != "" {
						writeSSEJSON(w, "text", map[string]string{"text": ev.Text})
					}
				case "run_status":
					writeSSEJSON(w, "run_status", map[string]string{
						"status":  ev.Status,
						"summary": ev.Summary,
					})
				case "error":
					writeSSEJSON(w, "error", map[string]string{"error": ev.Error})
					flusher.Flush()
					return
				case "done":
					fmt.Fprintf(w, "event: done\ndata: [DONE]\n\n")
					flusher.Flush()
					return
				}
				flusher.Flush()
			}
		}
	}
}

func parseAfterToolEventID(raw string) int64 {
	if raw == "" {
		return 0
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func isTerminalRunStatus(status string) bool {
	switch status {
	case agent.RunStatusCompleted, agent.RunStatusFailed, agent.RunStatusBlocked, agent.RunStatusNeedsApproval:
		return true
	default:
		return false
	}
}
