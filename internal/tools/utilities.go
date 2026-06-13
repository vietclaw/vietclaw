package tools

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// DirList lists files and subdirectories in a directory path.
type DirList struct {
	Policy Policy
}

func (t DirList) Name() string { return "dir_list" }

type DirEntryInfo struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size_bytes,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
}

func (t DirList) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		args.Path = strings.TrimSpace(input)
	}

	if args.Path == "" {
		args.Path = "."
	}

	allowedPath, err := t.Policy.FileAllowed(args.Path)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(allowedPath)
	if err != nil {
		return "", err
	}

	var results []DirEntryInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		var size int64
		if !entry.IsDir() {
			size = info.Size()
		}
		results = append(results, DirEntryInfo{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    size,
			ModTime: info.ModTime().Format(time.RFC3339),
		})
	}

	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// FileGrep searches file contents for lines matching a pattern.
type FileGrep struct {
	Policy Policy
}

func (t FileGrep) Name() string { return "file_grep" }

type GrepMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

func (t FileGrep) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Path    string `json:"path"`
		Pattern string `json:"pattern"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments format: %w", err)
	}

	if args.Path == "" || args.Pattern == "" {
		return "", fmt.Errorf("path and pattern are required")
	}

	allowedPath, err := t.Policy.FileAllowed(args.Path)
	if err != nil {
		return "", err
	}

	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regular expression: %w", err)
	}

	matches, err := grepInPath(ctx, t.Policy, allowedPath, re)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(matches, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func grepInPath(ctx context.Context, policy Policy, rootPath string, re *regexp.Regexp) ([]GrepMatch, error) {
	var matches []GrepMatch
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return grepInFile(rootPath, re)
	}

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if _, err := policy.FileAllowed(path); err != nil {
			return nil
		}
		fileMatches, err := grepInFile(path, re)
		if err == nil {
			matches = append(matches, fileMatches...)
		}
		if len(matches) >= 100 {
			return io.EOF
		}
		return nil
	})
	if err == io.EOF {
		err = nil
	}
	return matches, err
}

func grepInFile(filePath string, re *regexp.Regexp) ([]GrepMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []GrepMatch
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		if re.MatchString(text) {
			matches = append(matches, GrepMatch{
				File:    filePath,
				Line:    lineNum,
				Content: text,
			})
		}
		if len(matches) >= 100 {
			break
		}
	}
	return matches, scanner.Err()
}

// FileFind finds files matching a glob pattern under a path.
type FileFind struct {
	Policy Policy
}

func (t FileFind) Name() string { return "file_find" }

func (t FileFind) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Path    string `json:"path"`
		Pattern string `json:"pattern"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments format: %w", err)
	}

	if args.Path == "" || args.Pattern == "" {
		return "", fmt.Errorf("path and pattern are required")
	}

	allowedPath, err := t.Policy.FileAllowed(args.Path)
	if err != nil {
		return "", err
	}

	var results []string
	err = filepath.Walk(allowedPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if _, err := t.Policy.FileAllowed(path); err != nil {
			return nil
		}
		matched, err := filepath.Match(args.Pattern, info.Name())
		if err == nil && matched {
			results = append(results, path)
		}
		return nil
	})

	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// SystemInfo retrieves hardware/OS system stats.
type SystemInfo struct{}

func (t SystemInfo) Name() string { return "system_info" }

func (t SystemInfo) Run(ctx context.Context, input string) (string, error) {
	hostname, _ := os.Hostname()
	info := map[string]any{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"cpu_count":  runtime.NumCPU(),
		"go_version": runtime.Version(),
		"hostname":   hostname,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	jsonBytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// NetworkPing checks network latency to a host.
type NetworkPing struct{}

func (t NetworkPing) Name() string { return "network_ping" }

func (t NetworkPing) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Host string `json:"host"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		args.Host = strings.TrimSpace(input)
	}

	if args.Host == "" {
		return "", fmt.Errorf("host is required")
	}

	// Basic host sanitization
	if strings.ContainsAny(args.Host, ";&|`$<>") {
		return "", fmt.Errorf("invalid characters in host name")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "-n", "4", args.Host)
	} else {
		cmd = exec.CommandContext(ctx, "ping", "-c", "4", args.Host)
	}

	out, err := cmd.CombinedOutput()
	return string(out), err
}

// EnvGet retrieves the value of an environment variable.
type EnvGet struct{}

func (t EnvGet) Name() string { return "env_get" }

func (t EnvGet) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		args.Key = strings.TrimSpace(input)
	}

	if args.Key == "" {
		return "", fmt.Errorf("key is required")
	}

	// Security filter: block sensitive terms
	upperKey := strings.ToUpper(args.Key)
	sensitiveTerms := []string{"KEY", "TOKEN", "SECRET", "PASSWORD", "AUTH", "CREDENTIAL", "PASSPHRASE", "PRIVATE", "CERT", "JWT", "SESSION"}
	for _, term := range sensitiveTerms {
		if strings.Contains(upperKey, term) {
			return "", fmt.Errorf("access to sensitive environment variable blocked for security reasons")
		}
	}

	val := os.Getenv(args.Key)
	if val == "" {
		return fmt.Sprintf("environment variable %s is not set", args.Key), nil
	}
	return val, nil
}

// HashCalc computes MD5, SHA-1, or SHA-256 hashes of a file.
type HashCalc struct {
	Policy Policy
}

func (t HashCalc) Name() string { return "hash_calc" }

func (t HashCalc) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Path string `json:"path"`
		Algo string `json:"algo"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments format: %w", err)
	}

	if args.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	allowedPath, err := t.Policy.FileAllowed(args.Path)
	if err != nil {
		return "", err
	}

	file, err := os.Open(allowedPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	algo := strings.ToLower(strings.TrimSpace(args.Algo))
	if algo == "" {
		algo = "sha256"
	}

	var h io.Writer
	var hashFn func() string

	switch algo {
	case "md5":
		hasher := md5.New()
		h = hasher
		hashFn = func() string { return hex.EncodeToString(hasher.Sum(nil)) }
	case "sha1":
		hasher := sha1.New()
		h = hasher
		hashFn = func() string { return hex.EncodeToString(hasher.Sum(nil)) }
	case "sha256":
		hasher := sha256.New()
		h = hasher
		hashFn = func() string { return hex.EncodeToString(hasher.Sum(nil)) }
	default:
		return "", fmt.Errorf("unsupported algorithm: %s. Supported: md5, sha1, sha256", algo)
	}

	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return hashFn(), nil
}

// JSONFormat formats or minifies JSON strings.
type JSONFormat struct{}

func (t JSONFormat) Name() string { return "json_format" }

func (t JSONFormat) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Text   string `json:"text"`
		Minify bool   `json:"minify"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments format: %w", err)
	}

	var raw map[string]any
	var rawArr []any
	isArr := false

	// Attempt to parse text as json object first, then as json array
	if err := json.Unmarshal([]byte(args.Text), &raw); err != nil {
		if errArr := json.Unmarshal([]byte(args.Text), &rawArr); errArr != nil {
			return "", fmt.Errorf("invalid JSON text input: %w", err)
		}
		isArr = true
	}

	var output []byte
	var err error
	if args.Minify {
		var buf bytes.Buffer
		if isArr {
			output, err = json.Marshal(rawArr)
		} else {
			output, err = json.Marshal(raw)
		}
		if err == nil {
			err = json.Compact(&buf, output)
			output = buf.Bytes()
		}
	} else {
		if isArr {
			output, err = json.MarshalIndent(rawArr, "", "  ")
		} else {
			output, err = json.MarshalIndent(raw, "", "  ")
		}
	}

	if err != nil {
		return "", err
	}
	return string(output), nil
}

// StringTransform applies various transformations on strings.
type StringTransform struct{}

func (t StringTransform) Name() string { return "string_transform" }

func (t StringTransform) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Text string `json:"text"`
		Op   string `json:"op"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments format: %w", err)
	}

	op := strings.ToLower(strings.TrimSpace(args.Op))
	switch op {
	case "base64_encode":
		return base64.StdEncoding.EncodeToString([]byte(args.Text)), nil
	case "base64_decode":
		dec, err := base64.StdEncoding.DecodeString(args.Text)
		if err != nil {
			return "", err
		}
		return string(dec), nil
	case "url_encode":
		return url.QueryEscape(args.Text), nil
	case "url_decode":
		dec, err := url.QueryUnescape(args.Text)
		if err != nil {
			return "", err
		}
		return string(dec), nil
	case "upper":
		return strings.ToUpper(args.Text), nil
	case "lower":
		return strings.ToLower(args.Text), nil
	default:
		return "", fmt.Errorf("unsupported operation: %s. Supported: base64_encode, base64_decode, url_encode, url_decode, upper, lower", op)
	}
}

// TimeCurrent gets current time.
type TimeCurrent struct{}

func (t TimeCurrent) Name() string { return "time_current" }

func (t TimeCurrent) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		TZ string `json:"tz"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		args.TZ = strings.TrimSpace(input)
	}

	loc := time.Local
	if args.TZ != "" {
		var err error
		loc, err = time.LoadLocation(args.TZ)
		if err != nil {
			return "", fmt.Errorf("failed to load timezone %s: %w", args.TZ, err)
		}
	}

	return time.Now().In(loc).Format("2006-01-02 15:04:05 MST"), nil
}

// MathCalc evaluates simple arithmetic expressions securely.
type MathCalc struct{}

func (t MathCalc) Name() string { return "math_calc" }

func (t MathCalc) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Expr string `json:"expr"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		args.Expr = strings.TrimSpace(input)
	}

	val, err := EvaluateMath(args.Expr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%g", val), nil
}

func EvaluateMath(expr string) (float64, error) {
	p := &mathParser{expr: strings.ReplaceAll(expr, " ", ""), pos: 0}
	val, err := p.parseExpression()
	if err != nil {
		return 0, err
	}
	if p.pos < len(p.expr) {
		return 0, fmt.Errorf("unexpected character '%c' at position %d", p.expr[p.pos], p.pos)
	}
	return val, nil
}

type mathParser struct {
	expr string
	pos  int
}

func (p *mathParser) parseExpression() (float64, error) {
	val, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.expr) {
		op := p.expr[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++
		nextVal, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			val += nextVal
		} else {
			val -= nextVal
		}
	}
	return val, nil
}

func (p *mathParser) parseTerm() (float64, error) {
	val, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.expr) {
		op := p.expr[p.pos]
		if op != '*' && op != '/' {
			break
		}
		p.pos++
		nextVal, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if op == '*' {
			val *= nextVal
		} else {
			if nextVal == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			val /= nextVal
		}
	}
	return val, nil
}

func (p *mathParser) parseFactor() (float64, error) {
	if p.pos >= len(p.expr) {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	if p.expr[p.pos] == '(' {
		p.pos++
		val, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		if p.pos >= len(p.expr) || p.expr[p.pos] != ')' {
			return 0, fmt.Errorf("expected closing parenthesis")
		}
		p.pos++
		return val, nil
	}

	start := p.pos
	if p.pos < len(p.expr) && (p.expr[p.pos] == '-' || p.expr[p.pos] == '+') {
		p.pos++
	}
	hasDot := false
	for p.pos < len(p.expr) {
		c := p.expr[p.pos]
		if c >= '0' && c <= '9' {
			p.pos++
		} else if c == '.' && !hasDot {
			hasDot = true
			p.pos++
		} else {
			break
		}
	}
	if p.pos == start || (p.pos == start+1 && (p.expr[start] == '-' || p.expr[start] == '+')) {
		return 0, fmt.Errorf("expected a number at position %d", start)
	}
	var val float64
	_, err := fmt.Sscanf(p.expr[start:p.pos], "%f", &val)
	if err != nil {
		return 0, fmt.Errorf("invalid number formatting at position %d", start)
	}
	return val, nil
}

// ProcessList lists active processes on the host.
type ProcessList struct{}

func (t ProcessList) Name() string { return "process_list" }

func (t ProcessList) Run(ctx context.Context, input string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "tasklist")
	} else {
		cmd = exec.CommandContext(ctx, "ps", "-ef")
	}

	out, err := cmd.CombinedOutput()
	return string(out), err
}

// IPLookup fetches external public IP and geolocation of the host.
type IPLookup struct{}

func (t IPLookup) Name() string { return "ip_lookup" }

func (t IPLookup) Run(ctx context.Context, _ string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://ipinfo.io/json", nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query ipinfo: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
