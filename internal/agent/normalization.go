package agent

import (
	"strings"

	"vietclaw/internal/config"
)

var memoryAddPrefixes = []string{"nhớ là", "lưu lại", "từ nay", "remember that", "save this", "note that"}
var memoryQueryReplacers = []string{"mày nhớ gì về", "mày nhớ gì", "nhớ gì về", "server chính dùng gì", "server chính là gì", "what do you remember about", "what do you remember", "recall", "?"}

func normalizeRequest(req ChatRequest, cfg config.Config) ChatRequest {
	if req.SessionID == "" {
		req.SessionID = newID("sess")
	}
	if req.UserID == "" {
		req.UserID = DefaultUserID
	}
	if req.AgentID == "" {
		req.AgentID = config.DefaultAgentID
	}
	if req.Channel == "" {
		req.Channel = DefaultChannel
	}
	if req.Mode == "" {
		req.Mode = cfg.Agent.DefaultMode
	}
	return req
}

func cleanMemoryContent(message string) string {
	text := strings.TrimSpace(message)
	lower := strings.ToLower(text)
	for _, prefix := range memoryAddPrefixes {
		if strings.HasPrefix(lower, prefix) {
			runes := []rune(text)
			return strings.TrimSpace(string(runes[len([]rune(prefix)):]))
		}
	}
	return text
}

func cleanMemoryQuery(message string) string {
	original := strings.TrimSpace(message)
	text := strings.ToLower(strings.TrimSpace(message))
	if strings.Contains(text, "server chính") {
		return "server chính"
	}
	for _, item := range memoryQueryReplacers {
		text = strings.ReplaceAll(text, item, "")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return original
	}
	return text
}
