package channels_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"vietclaw/internal/channels"
	"vietclaw/internal/config"
)

type flakyAdapter struct {
	starts chan struct{}
}

func (a flakyAdapter) Name() string { return channels.PlatformTelegram }
func (a flakyAdapter) Start(context.Context) error {
	a.starts <- struct{}{}
	return errors.New("temporary failure")
}

func TestManagerRetriesFailedAdapters(t *testing.T) {
	adapter := flakyAdapter{starts: make(chan struct{}, 2)}
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Channels.Telegram.Enabled = true
	manager := channels.NewManager(cfg, nil, []channels.Adapter{adapter})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.Start(ctx)

	for i := 0; i < 2; i++ {
		select {
		case <-adapter.starts:
		case <-time.After(12 * time.Second):
			t.Fatal("expected adapter retry")
		}
	}
}
