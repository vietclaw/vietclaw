package framework

import (
	"context"
	"sync"
)

type Event string

const (
	EventBeforeChat Event = "before_chat"
	EventAfterChat  Event = "after_chat"
	EventBeforeTool Event = "before_tool"
	EventAfterTool  Event = "after_tool"
	EventRunStart   Event = "run_start"
	EventRunFinish  Event = "run_finish"
)

type HookContext struct {
	Event      Event
	SessionID  string
	AgentID    string
	RunID      string
	ParentRunID string
	Message    string
	Reply      string
	ToolName   string
	ToolInput  string
	ToolOutput string
	ToolError  string
	Metadata   map[string]any
}

type Hook func(ctx context.Context, hc HookContext) error

type HookRegistry struct {
	mu    sync.RWMutex
	hooks map[Event][]Hook
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{hooks: make(map[Event][]Hook)}
}

func (r *HookRegistry) Register(event Event, hook Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[event] = append(r.hooks[event], hook)
}

func (r *HookRegistry) Emit(ctx context.Context, event Event, hc HookContext) error {
	r.mu.RLock()
	list := append([]Hook(nil), r.hooks[event]...)
	r.mu.RUnlock()
	hc.Event = event
	for _, hook := range list {
		if err := hook(ctx, hc); err != nil {
			return err
		}
	}
	return nil
}

func (r *HookRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, list := range r.hooks {
		n += len(list)
	}
	return n
}
