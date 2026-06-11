package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func DefaultPaths() (Paths, error) {
	dataDir, err := defaultDataDir()
	if err != nil {
		return Paths{}, err
	}
	return Paths{
		DataDir:    dataDir,
		ConfigFile: filepath.Join(dataDir, ConfigFileName),
		LogDir:     filepath.Join(dataDir, LogDirName),
		LogFile:    filepath.Join(dataDir, LogDirName, LogFileName),
	}, nil
}

func ExpandPath(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	if len(path) >= 2 && path[0] == '~' && (path[1] == '/' || path[1] == '\\') {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func defaultDataDir() (string, error) {
	if env := strings.TrimSpace(os.Getenv("VIETCLAW_DATA_DIR")); env != "" {
		return ExpandPath(env), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".vietclaw"), nil
}

// LegacyWindowsDataDir is the pre-unification Windows default (%APPDATA%\VietClaw).
func LegacyWindowsDataDir() (string, bool) {
	if runtime.GOOS != "windows" {
		return "", false
	}
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		return "", false
	}
	return filepath.Join(configDir, AppName), true
}
