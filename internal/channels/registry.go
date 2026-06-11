package channels

import (
	"fmt"

	"vietclaw/internal/config"
)

type AdapterFactory func(cfg config.Config, handler *Handler) (Adapter, error)

var adapterFactories = map[string]AdapterFactory{}

func RegisterAdapter(name string, factory AdapterFactory) {
	adapterFactories[name] = factory
}

func RegisteredAdapters() []string {
	out := make([]string, 0, len(adapterFactories))
	for name := range adapterFactories {
		out = append(out, name)
	}
	return out
}

func BuildAdapters(cfg config.Config, handler *Handler) ([]Adapter, error) {
	var adapters []Adapter
	var errs []error

	if cfg.Channels.Discord.Enabled {
		a, err := buildAdapter(PlatformDiscord, cfg, handler)
		if err != nil {
			errs = append(errs, err)
		} else if a != nil {
			adapters = append(adapters, a)
		}
	}
	if cfg.Channels.Telegram.Enabled {
		a, err := buildAdapter(PlatformTelegram, cfg, handler)
		if err != nil {
			errs = append(errs, err)
		} else if a != nil {
			adapters = append(adapters, a)
		}
	}
	if len(errs) > 0 {
		return adapters, fmt.Errorf("channel adapter errors: %v", errs)
	}
	return adapters, nil
}

func buildAdapter(name string, cfg config.Config, handler *Handler) (Adapter, error) {
	factory, ok := adapterFactories[name]
	if !ok {
		return nil, fmt.Errorf("channel adapter not registered: %s", name)
	}
	return factory(cfg, handler)
}
