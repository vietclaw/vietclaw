package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func EnsureDefault(paths Paths) (Config, bool, error) {
	if err := os.MkdirAll(paths.DataDir, 0o755); err != nil {
		return Config{}, false, fmt.Errorf("create data dir: %w", err)
	}
	if err := os.MkdirAll(paths.LogDir, 0o755); err != nil {
		return Config{}, false, fmt.Errorf("create log dir: %w", err)
	}

	cfg, raw, err := LoadWithRaw(paths.ConfigFile)
	if err == nil {
		def := Default(paths)
		merged := MergeDefault(cfg, def)
		merged.Agent = MergeAgentOptional(merged.Agent, def.Agent, raw)
		merged.Framework = MergeFrameworkOptional(merged.Framework, def.Framework, raw)
		merged = ApplyLegacyMigrations(merged)
		if !Equal(cfg, merged) {
			if err := Save(paths.ConfigFile, merged); err != nil {
				return Config{}, false, err
			}
		}
		merged.Database.Path = ExpandPath(merged.Database.Path)
		merged.Agent.Workspace = ExpandPath(merged.Agent.Workspace)
		return merged, false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return Config{}, false, err
	}

	cfg = Default(paths)
	if err := Save(paths.ConfigFile, cfg); err != nil {
		return Config{}, false, err
	}
	return cfg, true, nil
}

func Load(path string) (Config, error) {
	cfg, _, err := LoadWithRaw(path)
	return cfg, err
}

func LoadWithRaw(path string) (Config, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, data, fmt.Errorf("parse config: %w", err)
	}
	cfg.Database.Path = ExpandPath(cfg.Database.Path)
	cfg.Agent.Workspace = ExpandPath(cfg.Agent.Workspace)
	return cfg, data, nil
}

func Save(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func Equal(a, b Config) bool {
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}
	right, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(left) == string(right)
}
