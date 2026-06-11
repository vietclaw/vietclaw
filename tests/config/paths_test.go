package config_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"vietclaw/internal/config"
)

func TestDefaultPathsUsesHomeDotVietclaw(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("VIETCLAW_DATA_DIR", "")

	paths, err := config.DefaultPaths()
	if err != nil {
		t.Fatalf("DefaultPaths: %v", err)
	}
	want := filepath.Join(home, ".vietclaw")
	if paths.DataDir != want {
		t.Fatalf("DataDir = %q, want %q", paths.DataDir, want)
	}
	if paths.ConfigFile != filepath.Join(want, "config.json") {
		t.Fatalf("ConfigFile = %q", paths.ConfigFile)
	}
}

func TestDefaultPathsVIETCLAW_DATA_DIR(t *testing.T) {
	custom := filepath.Join(t.TempDir(), "custom-data")
	t.Setenv("VIETCLAW_DATA_DIR", custom)

	paths, err := config.DefaultPaths()
	if err != nil {
		t.Fatalf("DefaultPaths: %v", err)
	}
	if paths.DataDir != custom {
		t.Fatalf("DataDir = %q, want %q", paths.DataDir, custom)
	}
}

func TestDefaultPathsVIETCLAW_DATA_DIRExpandHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("VIETCLAW_DATA_DIR", "~/.vietclaw")

	paths, err := config.DefaultPaths()
	if err != nil {
		t.Fatalf("DefaultPaths: %v", err)
	}
	want := filepath.Join(home, ".vietclaw")
	if paths.DataDir != want {
		t.Fatalf("DataDir = %q, want %q", paths.DataDir, want)
	}
}

func TestLegacyWindowsDataDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		dir, ok := config.LegacyWindowsDataDir()
		if !ok || dir == "" {
			t.Fatalf("expected legacy dir on windows")
		}
		return
	}
	dir, ok := config.LegacyWindowsDataDir()
	if ok || dir != "" {
		t.Fatalf("legacy dir on non-windows = %q, ok=%v", dir, ok)
	}
}
