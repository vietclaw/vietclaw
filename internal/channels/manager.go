package channels

import (
	"context"
	"log"
	"sync"
	"time"

	"vietclaw/internal/config"
)

const adapterRestartDelay = 10 * time.Second

type Manager struct {
	adapters []Adapter
	logger   *log.Logger
	statuses map[string]Status
	mu       sync.Mutex
}

func NewManager(cfg config.Config, logger *log.Logger, adapters []Adapter) *Manager {
	statuses := map[string]Status{}
	for _, name := range RegisteredAdapters() {
		statuses[name] = Status{Name: name, Enabled: channelEnabled(cfg, name)}
	}
	for _, adapter := range adapters {
		status := statuses[adapter.Name()]
		status.Name = adapter.Name()
		status.Enabled = true
		statuses[adapter.Name()] = status
	}
	return &Manager{logger: logger, adapters: adapters, statuses: statuses}
}

func channelEnabled(cfg config.Config, name string) bool {
	switch name {
	case PlatformDiscord:
		return cfg.Channels.Discord.Enabled
	case PlatformTelegram:
		return cfg.Channels.Telegram.Enabled
	default:
		return false
	}
}

func (m *Manager) Start(ctx context.Context) {
	for _, adapter := range m.adapters {
		adapter := adapter
		go func() {
			for ctx.Err() == nil {
				m.setRunning(adapter.Name(), true, "")
				if m.logger != nil {
					m.logger.Printf("channel adapter starting name=%s", adapter.Name())
				}
				err := adapter.Start(ctx)
				if err == nil || ctx.Err() != nil {
					break
				}
				m.setRunning(adapter.Name(), false, err.Error())
				if m.logger != nil {
					m.logger.Printf("channel adapter failed name=%s err=%v; retrying in %s", adapter.Name(), err, adapterRestartDelay)
				}
				select {
				case <-ctx.Done():
				case <-time.After(adapterRestartDelay):
				}
			}
			m.setRunning(adapter.Name(), false, "")
			if m.logger != nil {
				m.logger.Printf("channel adapter stopped name=%s", adapter.Name())
			}
		}()
	}
}

func (m *Manager) Statuses() []Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Status, 0, len(m.statuses))
	for _, status := range m.statuses {
		out = append(out, status)
	}
	return out
}

func (m *Manager) setRunning(name string, running bool, errText string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	status := m.statuses[name]
	status.Name = name
	status.Running = running
	status.Error = errText
	m.statuses[name] = status
}

func StatusFromConfig(cfg config.Config) []Status {
	return []Status{
		{Name: PlatformDiscord, Enabled: cfg.Channels.Discord.Enabled},
		{Name: PlatformTelegram, Enabled: cfg.Channels.Telegram.Enabled},
	}
}
