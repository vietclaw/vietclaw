package tools_test

import (
	"context"
	"path/filepath"
	"reflect"
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
