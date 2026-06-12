package agent

import "sync"

type SessionEvent struct {
	Event       string `json:"event"`
	SessionID   string `json:"session_id,omitempty"`
	Text        string `json:"text,omitempty"`
	ToolName    string `json:"tool_name,omitempty"`
	ToolInput   string `json:"tool_input,omitempty"`
	ToolResult  string `json:"tool_result,omitempty"`
	Status      string `json:"status,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Error       string `json:"error,omitempty"`
	ToolEventID int64  `json:"tool_event_id,omitempty"`
}

type RunEventHub struct {
	mu   sync.RWMutex
	subs map[string]map[chan SessionEvent]struct{}
}

func NewRunEventHub() *RunEventHub {
	return &RunEventHub{subs: make(map[string]map[chan SessionEvent]struct{})}
}

func (h *RunEventHub) Subscribe(sessionID string) (<-chan SessionEvent, func()) {
	ch := make(chan SessionEvent, 64)
	h.mu.Lock()
	if h.subs[sessionID] == nil {
		h.subs[sessionID] = make(map[chan SessionEvent]struct{})
	}
	h.subs[sessionID][ch] = struct{}{}
	h.mu.Unlock()

	unsub := func() {
		h.mu.Lock()
		if set, ok := h.subs[sessionID]; ok {
			delete(set, ch)
			if len(set) == 0 {
				delete(h.subs, sessionID)
			}
		}
		h.mu.Unlock()
		close(ch)
	}
	return ch, unsub
}

func (h *RunEventHub) Publish(sessionID string, ev SessionEvent) {
	if sessionID == "" {
		return
	}
	ev.SessionID = sessionID
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[sessionID] {
		select {
		case ch <- ev:
		default:
		}
	}
}
