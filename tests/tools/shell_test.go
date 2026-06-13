package tools_test

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"vietclaw/internal/config"
	"vietclaw/internal/tools"
)

func TestBuildDockerShellArgs(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default(config.Paths{DataDir: dir})
	cfg.Tools.Shell.Sandbox = "docker"
	cfg.Tools.Shell.DockerImage = "busybox:1.36"
	cfg.Tools.Shell.DockerNetwork = "bridge"
	cfg.Tools.Shell.WorkspaceMode = "ro"

	workspace, err := filepath.Abs(cfg.Agent.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	got := tools.BuildDockerShellArgs(cfg, []string{"echo", "hello"})
	want := []string{
		"run",
		"--rm",
		"--network", "bridge",
		"-v", workspace + ":/workspace:ro",
		"-w", "/workspace",
		"busybox:1.36",
		"echo", "hello",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("docker args = %#v, want %#v", got, want)
	}
}

func TestShellNetworkPolicyBlocksPrivateTargets(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Tools.Shell.Enabled = true
	cfg.Tools.Shell.NetworkPolicy.Enabled = true
	cfg.Tools.Shell.NetworkPolicy.DenyPrivate = true
	policy := tools.NewPolicy(cfg)

	if err := policy.ShellNetworkAllowed("curl http://169.254.169.254/latest/meta-data"); err == nil {
		t.Fatal("expected metadata IP to be blocked")
	}
	if err := policy.ShellNetworkAllowed("curl http://127.0.0.1:8080"); err == nil {
		t.Fatal("expected localhost to be blocked")
	}
	if err := policy.ShellNetworkAllowed("curl https://example.com"); err != nil {
		t.Fatalf("expected public URL to be allowed: %v", err)
	}
}

func TestShellExecBlocksNetworkPolicyBeforeRun(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Tools.Shell.Enabled = true
	cfg.Tools.Shell.NetworkPolicy.Enabled = true
	cfg.Tools.Shell.NetworkPolicy.DenyPrivate = true
	execTool := tools.ShellExec{Policy: tools.NewPolicy(cfg)}
	if _, err := execTool.Run(context.Background(), "curl http://localhost"); err == nil {
		t.Fatal("expected shell exec to reject localhost")
	}
}

func TestHTTPToolsRespectNetworkPolicy(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Tools.Shell.NetworkPolicy.Enabled = true
	cfg.Tools.Shell.NetworkPolicy.DenyPrivate = true
	registry := tools.NewRegistry(cfg)

	if _, err := registry.Execute(context.Background(), "http_request", `{"url":"http://169.254.169.254/latest/meta-data"}`); err == nil {
		t.Fatal("expected http_request to reject metadata IP")
	}
	if _, err := registry.Execute(context.Background(), "web_fetch", `{"url":"http://127.0.0.1"}`); err == nil {
		t.Fatal("expected web_fetch to reject localhost")
	}
}

func TestShellNetworkPolicyBlocksIPBypass(t *testing.T) {
	cfg := config.Default(config.Paths{DataDir: t.TempDir()})
	cfg.Tools.Shell.Enabled = true
	cfg.Tools.Shell.NetworkPolicy.Enabled = true
	cfg.Tools.Shell.NetworkPolicy.DenyPrivate = true
	policy := tools.NewPolicy(cfg)

	// These should all resolve to 127.0.0.1 and be blocked
	bypassAttempts := []string{
		"curl 2130706433",       // Integer
		"curl 0x7f000001",       // Hex
		"curl 0177.0.0.1",       // Octal
		"curl 127.1",            // Shorthand
		"curl 127.0.1",          // Shorthand
		"curl 0177.000.000.001", // Octal with leading zeros
	}

	for _, cmd := range bypassAttempts {
		err := policy.ShellNetworkAllowed(cmd)
		if err == nil {
			t.Errorf("expected bypass attempt %q to be blocked, but it was allowed", cmd)
		} else if !strings.Contains(err.Error(), "blocked shell network IP") && !strings.Contains(err.Error(), "blocked shell network host") {
			t.Errorf("expected bypass attempt %q to be blocked due to network IP policy, got err: %v", cmd, err)
		}
	}
}
