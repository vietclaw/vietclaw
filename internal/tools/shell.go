package tools

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"vietclaw/internal/config"
)

type ShellExec struct {
	Policy Policy
}

func (t ShellExec) Name() string { return "shell.exec" }

func (t ShellExec) Run(ctx context.Context, input string) (string, error) {
	if !t.Policy.ShellAllowed() {
		return "", fmt.Errorf("shell.exec disabled")
	}
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return "", fmt.Errorf("empty command")
	}
	if err := t.Policy.ShellNetworkAllowed(input); err != nil {
		return "", err
	}
	if t.Policy.cfg.Tools.Shell.Sandbox == "docker" {
		return t.runDocker(ctx, fields)
	}
	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", input)
		cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8", "PYTHONUTF8=1")
		out, err := cmd.CombinedOutput()
		return CombinedOutputResult(out, err)
	}
	out, err := exec.CommandContext(ctx, fields[0], fields[1:]...).CombinedOutput()
	return CombinedOutputResult(out, err)
}

func (t ShellExec) runDocker(ctx context.Context, command []string) (string, error) {
	cfg := t.Policy.cfg.Tools.Shell
	workspace, err := filepath.Abs(config.ExpandPath(t.Policy.cfg.Agent.Workspace))
	if err != nil {
		return "", err
	}
	runCtx := ctx
	cancel := func() {}
	if cfg.TimeoutSeconds > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(cfg.TimeoutSeconds)*time.Second)
	}
	defer cancel()

	args := BuildDockerShellArgs(t.Policy.cfg, command)
	out, err := exec.CommandContext(runCtx, defaultString(cfg.DockerBinary, config.DefaultDockerBinary), args...).CombinedOutput()
	if runCtx.Err() == context.DeadlineExceeded {
		return CombinedOutputResult(out, fmt.Errorf("shell.exec docker timeout"))
	}
	if !strings.Contains(strings.Join(args, " "), workspace) {
		return CombinedOutputResult(out, fmt.Errorf("shell.exec docker workspace mount missing"))
	}
	return CombinedOutputResult(out, err)
}

func BuildDockerShellArgs(cfg config.Config, command []string) []string {
	shellCfg := cfg.Tools.Shell
	workspace, _ := filepath.Abs(config.ExpandPath(cfg.Agent.Workspace))
	mode := defaultString(shellCfg.WorkspaceMode, config.DefaultWorkspaceMode)
	if mode != "rw" {
		mode = "ro"
	}
	network := defaultString(shellCfg.DockerNetwork, config.DefaultDockerNetwork)
	image := defaultString(shellCfg.DockerImage, config.DefaultDockerImage)

	args := []string{
		"run",
		"--rm",
		"--network", network,
		"-v", workspace + ":/workspace:" + mode,
		"-w", "/workspace",
		image,
	}
	return append(args, command...)
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

var shellURLPattern = regexp.MustCompile(`(?i)\bhttps?://[^\s"'<>]+`)

func (p Policy) ShellNetworkAllowed(command string) error {
	policy := p.cfg.Tools.Shell.NetworkPolicy
	if !policy.Enabled {
		return nil
	}
	for _, rawURL := range shellURLPattern.FindAllString(command, -1) {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("blocked shell network target: invalid URL %s", rawURL)
		}
		if err := p.hostAllowed(parsed.Hostname()); err != nil {
			return err
		}
	}
	for _, field := range strings.Fields(command) {
		candidate := strings.Trim(field, `"'[](),;`)
		if strings.Contains(candidate, "/") || strings.Contains(candidate, "\\") || strings.Contains(candidate, "$") {
			continue
		}
		normalized := normalizeIPv4(candidate)
		if ip := net.ParseIP(normalized); ip != nil {
			if err := p.ipAllowed(candidate, ip); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p Policy) HTTPURLAllowed(rawURL string) error {
	policy := p.cfg.Tools.Shell.NetworkPolicy
	if !policy.Enabled {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("blocked network target: invalid URL %s", rawURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("blocked network target: unsupported scheme %s", parsed.Scheme)
	}
	return p.hostAllowed(parsed.Hostname())
}

func (p Policy) hostAllowed(host string) error {
	host = strings.Trim(strings.ToLower(host), ".")
	if host == "" {
		return nil
	}
	policy := p.cfg.Tools.Shell.NetworkPolicy
	for _, pattern := range policy.DenyHosts {
		if hostMatches(host, pattern) {
			return fmt.Errorf("blocked shell network host: %s", host)
		}
	}
	if ip := net.ParseIP(host); ip != nil {
		return p.ipAllowed(host, ip)
	}
	if policy.DenyPrivate {
		addrs, err := net.LookupIP(host)
		if err == nil {
			for _, ip := range addrs {
				if privateIP(ip) {
					return fmt.Errorf("blocked shell network host: %s resolves to private IP", host)
				}
			}
		}
	}
	if policy.RestrictToAllowHosts && !hostInPatterns(host, policy.AllowHosts) {
		return fmt.Errorf("blocked shell network host: %s is not in allow list", host)
	}
	return nil
}

func (p Policy) ipAllowed(label string, ip net.IP) error {
	if p.cfg.Tools.Shell.NetworkPolicy.DenyPrivate && privateIP(ip) {
		return fmt.Errorf("blocked shell network IP: %s", label)
	}
	return nil
}

func hostMatches(host string, pattern string) bool {
	pattern = strings.Trim(strings.ToLower(pattern), ".")
	if pattern == "" {
		return false
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(host, suffix)
	}
	return host == pattern
}

func hostInPatterns(host string, patterns []string) bool {
	for _, pattern := range patterns {
		if hostMatches(host, pattern) {
			return true
		}
	}
	return false
}

func privateIP(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

// normalizeIPv4 attempts to parse a string as an IPv4 address according to
// standard inet_aton rules (allowing hex, octal, integer, and < 4 parts), and
// returns it in dotted-quad format. If parsing fails, it returns the original string.
func normalizeIPv4(ipStr string) string {
	parts := strings.Split(ipStr, ".")
	if len(parts) > 4 {
		return ipStr
	}

	var parsedParts []int64
	for _, part := range parts {
		if part == "" {
			return ipStr
		}

		var val int64
		var err error

		if strings.HasPrefix(strings.ToLower(part), "0x") {
			val, err = strconv.ParseInt(part[2:], 16, 64)
		} else if strings.HasPrefix(part, "0") && len(part) > 1 {
			val, err = strconv.ParseInt(part[1:], 8, 64)
		} else {
			val, err = strconv.ParseInt(part, 10, 64)
		}

		if err != nil || val < 0 || val > 0xffffffff {
			return ipStr
		}
		parsedParts = append(parsedParts, val)
	}

	if len(parsedParts) == 0 {
		return ipStr
	}

	var finalParts [4]int64

	switch len(parsedParts) {
	case 1:
		finalParts[0] = (parsedParts[0] >> 24) & 0xff
		finalParts[1] = (parsedParts[0] >> 16) & 0xff
		finalParts[2] = (parsedParts[0] >> 8) & 0xff
		finalParts[3] = parsedParts[0] & 0xff
	case 2:
		if parsedParts[0] > 255 {
			return ipStr
		}
		finalParts[0] = parsedParts[0]
		finalParts[1] = (parsedParts[1] >> 16) & 0xff
		finalParts[2] = (parsedParts[1] >> 8) & 0xff
		finalParts[3] = parsedParts[1] & 0xff
	case 3:
		if parsedParts[0] > 255 || parsedParts[1] > 255 {
			return ipStr
		}
		finalParts[0] = parsedParts[0]
		finalParts[1] = parsedParts[1]
		finalParts[2] = (parsedParts[2] >> 8) & 0xff
		finalParts[3] = parsedParts[2] & 0xff
	case 4:
		if parsedParts[0] > 255 || parsedParts[1] > 255 || parsedParts[2] > 255 || parsedParts[3] > 255 {
			return ipStr
		}
		finalParts[0] = parsedParts[0]
		finalParts[1] = parsedParts[1]
		finalParts[2] = parsedParts[2]
		finalParts[3] = parsedParts[3]
	}

	return fmt.Sprintf("%d.%d.%d.%d", finalParts[0], finalParts[1], finalParts[2], finalParts[3])
}
