package config

import "fmt"

const (
	ChannelDiscord  = "discord"
	ChannelTelegram = "telegram"
)

func UpdateChannelEnabled(cfg Config, name string, enabled bool) (Config, error) {
	switch name {
	case ChannelDiscord:
		cfg.Channels.Discord.Enabled = enabled
	case ChannelTelegram:
		cfg.Channels.Telegram.Enabled = enabled
	default:
		return cfg, fmt.Errorf("unknown channel %q", name)
	}
	return cfg, nil
}
