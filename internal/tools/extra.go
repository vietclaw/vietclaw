package tools

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"vietclaw/internal/providers"
)

const extraHTTPTimeout = 15 * time.Second

type toolSpec struct {
	Name        string
	Description string
	Properties  map[string]any
	Required    []string
}

func registerExtraTools(registry map[string]Tool, policy Policy) {
	registry["uuid_generate"] = UUIDGenerate{}
	registry["random_string"] = RandomString{}
	registry["regex_extract"] = RegexExtract{}
	registry["regex_replace"] = RegexReplace{}
	registry["text_stats"] = TextStats{}
	registry["markdown_to_text"] = MarkdownToText{}
	registry["csv_preview"] = CSVPreview{}
	registry["csv_to_json"] = CSVToJSON{}
	registry["json_validate"] = JSONValidate{}
	registry["json_query"] = JSONQuery{}
	registry["url_parse"] = URLParse{}
	registry["html_to_text"] = HTMLToText{}
	registry["dns_lookup"] = DNSLookup{}
	registry["http_request"] = HTTPRequest{Policy: policy}
	registry["timestamp_parse"] = TimestampParse{}
	registry["timestamp_format"] = TimestampFormat{}
	registry["file_stat"] = FileStat{Policy: policy}
	registry["file_head"] = FileHead{Policy: policy}
	registry["file_tail"] = FileTail{Policy: policy}
	registry["path_info"] = PathInfo{Policy: policy}
}

func extraToolDefinitions() []providers.ToolDefinition {
	defs := make([]providers.ToolDefinition, 0, len(extraToolSpecs()))
	for _, spec := range extraToolSpecs() {
		params := map[string]any{"type": "object", "properties": spec.Properties}
		if len(spec.Required) > 0 {
			params["required"] = spec.Required
		}
		defs = append(defs, providers.ToolDefinition{
			Type: "function",
			Function: providers.FunctionDetail{
				Name:        spec.Name,
				Description: spec.Description,
				Parameters:  params,
			},
		})
	}
	return defs
}

func extraToolSpecs() []toolSpec {
	return []toolSpec{
		{Name: "uuid_generate", Description: "Generate one or more random UUID v4 identifiers.", Properties: props("count", "number", "Number of UUIDs to generate."), Required: nil},
		{Name: "random_string", Description: "Generate a random string using url-safe characters.", Properties: props("length", "number", "String length."), Required: []string{"length"}},
		{Name: "regex_extract", Description: "Extract matches from text using a regular expression.", Properties: mergeProps(props("text", "string", "Input text."), props("pattern", "string", "Regular expression pattern.")), Required: []string{"text", "pattern"}},
		{Name: "regex_replace", Description: "Replace regex matches in text.", Properties: mergeProps(mergeProps(props("text", "string", "Input text."), props("pattern", "string", "Regular expression pattern.")), props("replacement", "string", "Replacement text.")), Required: []string{"text", "pattern", "replacement"}},
		{Name: "text_stats", Description: "Count characters, words, lines, and bytes in text.", Properties: props("text", "string", "Input text."), Required: []string{"text"}},
		{Name: "markdown_to_text", Description: "Convert simple Markdown into plain text.", Properties: props("text", "string", "Markdown text."), Required: []string{"text"}},
		{Name: "csv_preview", Description: "Preview CSV/TSV text with row and column counts.", Properties: mergeProps(props("text", "string", "CSV or TSV text."), props("delimiter", "string", "Optional delimiter, defaults to comma.")), Required: []string{"text"}},
		{Name: "csv_to_json", Description: "Convert CSV/TSV text into JSON objects using the first row as headers.", Properties: mergeProps(props("text", "string", "CSV or TSV text."), props("delimiter", "string", "Optional delimiter, defaults to comma.")), Required: []string{"text"}},
		{Name: "json_validate", Description: "Validate JSON text and return a compact status.", Properties: props("text", "string", "JSON text."), Required: []string{"text"}},
		{Name: "json_query", Description: "Read a simple dot-separated path from JSON text.", Properties: mergeProps(props("text", "string", "JSON text."), props("path", "string", "Dot-separated path, e.g. user.name or items.0.id.")), Required: []string{"text", "path"}},
		{Name: "url_parse", Description: "Parse a URL into scheme, host, path, query, and fragment.", Properties: props("url", "string", "URL to parse."), Required: []string{"url"}},
		{Name: "html_to_text", Description: "Convert HTML text into readable plain text.", Properties: props("html", "string", "HTML text."), Required: []string{"html"}},
		{Name: "dns_lookup", Description: "Resolve a hostname to IP addresses.", Properties: props("host", "string", "Hostname to resolve."), Required: []string{"host"}},
		{Name: "http_request", Description: "Send a GET or HEAD HTTP request and return status, headers, and a short body preview.", Properties: mergeProps(props("url", "string", "URL to request."), props("method", "string", "GET or HEAD, defaults to GET.")), Required: []string{"url"}},
		{Name: "timestamp_parse", Description: "Parse an RFC3339 timestamp or Unix timestamp into common formats.", Properties: props("value", "string", "Timestamp value."), Required: []string{"value"}},
		{Name: "timestamp_format", Description: "Format a Unix timestamp with an optional timezone.", Properties: mergeProps(props("unix", "number", "Unix timestamp seconds."), props("tz", "string", "Optional IANA timezone.")), Required: []string{"unix"}},
		{Name: "file_stat", Description: "Return metadata for a file or directory in the workspace.", Properties: props("path", "string", "Workspace file or directory path."), Required: []string{"path"}},
		{Name: "file_head", Description: "Read the first N lines from a workspace file.", Properties: mergeProps(props("path", "string", "Workspace file path."), props("lines", "number", "Number of lines, defaults to 20.")), Required: []string{"path"}},
		{Name: "file_tail", Description: "Read the last N lines from a workspace file.", Properties: mergeProps(props("path", "string", "Workspace file path."), props("lines", "number", "Number of lines, defaults to 20.")), Required: []string{"path"}},
		{Name: "path_info", Description: "Normalize and inspect a workspace path.", Properties: props("path", "string", "Workspace path."), Required: []string{"path"}},
	}
}

func props(name, typ, description string) map[string]any {
	return map[string]any{name: map[string]any{"type": typ, "description": description}}
}

func mergeProps(a, b map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

type UUIDGenerate struct{}

func (t UUIDGenerate) Name() string { return "uuid_generate" }
func (t UUIDGenerate) Run(_ context.Context, input string) (string, error) {
	var args struct {
		Count int `json:"count"`
	}
	_ = json.Unmarshal([]byte(input), &args)
	if args.Count <= 0 {
		args.Count = 1
	}
	if args.Count > 100 {
		args.Count = 100
	}
	ids := make([]string, 0, args.Count)
	for i := 0; i < args.Count; i++ {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		ids = append(ids, fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
	}
	return marshalJSON(ids)
}

type RandomString struct{}

func (t RandomString) Name() string { return "random_string" }
func (t RandomString) Run(_ context.Context, input string) (string, error) {
	var args struct {
		Length int `json:"length"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	if args.Length <= 0 || args.Length > 4096 {
		return "", fmt.Errorf("length must be between 1 and 4096")
	}
	buf := make([]byte, args.Length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf)[:args.Length], nil
}

type RegexExtract struct{}

func (t RegexExtract) Name() string { return "regex_extract" }
func (t RegexExtract) Run(_ context.Context, input string) (string, error) {
	var args struct{ Text, Pattern string }
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return "", err
	}
	return marshalJSON(re.FindAllString(args.Text, 100))
}

type RegexReplace struct{}

func (t RegexReplace) Name() string { return "regex_replace" }
func (t RegexReplace) Run(_ context.Context, input string) (string, error) {
	var args struct{ Text, Pattern, Replacement string }
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(args.Text, args.Replacement), nil
}

type TextStats struct{}

func (t TextStats) Name() string { return "text_stats" }
func (t TextStats) Run(_ context.Context, input string) (string, error) {
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	return marshalJSON(map[string]any{
		"bytes":      len([]byte(args.Text)),
		"characters": len([]rune(args.Text)),
		"words":      len(strings.Fields(args.Text)),
		"lines":      len(strings.Split(args.Text, "\n")),
	})
}

type MarkdownToText struct{}

func (t MarkdownToText) Name() string { return "markdown_to_text" }
func (t MarkdownToText) Run(_ context.Context, input string) (string, error) {
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	text := regexp.MustCompile("(?s)```.*?```").ReplaceAllString(args.Text, "")
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`[*_#>`+"`"+`~-]+`).ReplaceAllString(text, "")
	return strings.TrimSpace(text), nil
}

type CSVPreview struct{}

func (t CSVPreview) Name() string { return "csv_preview" }
func (t CSVPreview) Run(_ context.Context, input string) (string, error) {
	records, err := readCSVInput(input)
	if err != nil {
		return "", err
	}
	preview := records
	if len(preview) > 5 {
		preview = preview[:5]
	}
	cols := 0
	if len(records) > 0 {
		cols = len(records[0])
	}
	return marshalJSON(map[string]any{"rows": len(records), "columns": cols, "preview": preview})
}

type CSVToJSON struct{}

func (t CSVToJSON) Name() string { return "csv_to_json" }
func (t CSVToJSON) Run(_ context.Context, input string) (string, error) {
	records, err := readCSVInput(input)
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "[]", nil
	}
	headers := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		row := map[string]string{}
		for i, header := range headers {
			if i < len(record) {
				row[header] = record[i]
			}
		}
		rows = append(rows, row)
	}
	return marshalJSON(rows)
}

type JSONValidate struct{}

func (t JSONValidate) Name() string { return "json_validate" }
func (t JSONValidate) Run(_ context.Context, input string) (string, error) {
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	var v any
	err := json.Unmarshal([]byte(args.Text), &v)
	return marshalJSON(map[string]any{"valid": err == nil, "error": errorString(err)})
}

type JSONQuery struct{}

func (t JSONQuery) Name() string { return "json_query" }
func (t JSONQuery) Run(_ context.Context, input string) (string, error) {
	var args struct{ Text, Path string }
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	var value any
	if err := json.Unmarshal([]byte(args.Text), &value); err != nil {
		return "", err
	}
	for _, part := range strings.Split(args.Path, ".") {
		switch current := value.(type) {
		case map[string]any:
			value = current[part]
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(current) {
				return "", fmt.Errorf("array index out of range: %s", part)
			}
			value = current[i]
		default:
			return "", fmt.Errorf("path not found: %s", args.Path)
		}
	}
	return marshalJSON(value)
}

type URLParse struct{}

func (t URLParse) Name() string { return "url_parse" }
func (t URLParse) Run(_ context.Context, input string) (string, error) {
	var args struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	u, err := url.Parse(args.URL)
	if err != nil {
		return "", err
	}
	return marshalJSON(map[string]any{"scheme": u.Scheme, "host": u.Host, "path": u.Path, "query": u.Query(), "fragment": u.Fragment})
}

type HTMLToText struct{}

func (t HTMLToText) Name() string { return "html_to_text" }
func (t HTMLToText) Run(_ context.Context, input string) (string, error) {
	var args struct {
		HTML string `json:"html"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	return StripHTML(args.HTML), nil
}

type DNSLookup struct{}

func (t DNSLookup) Name() string { return "dns_lookup" }
func (t DNSLookup) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Host string `json:"host"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	resolver := net.Resolver{}
	addrs, err := resolver.LookupHost(ctx, args.Host)
	if err != nil {
		return "", err
	}
	return marshalJSON(addrs)
}

type HTTPRequest struct {
	Policy Policy
}

func (t HTTPRequest) Name() string { return "http_request" }
func (t HTTPRequest) Run(ctx context.Context, input string) (string, error) {
	var args struct{ URL, Method string }
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	method := strings.ToUpper(strings.TrimSpace(args.Method))
	if method == "" {
		method = http.MethodGet
	}
	if method != http.MethodGet && method != http.MethodHead {
		return "", fmt.Errorf("method must be GET or HEAD")
	}
	if err := t.Policy.HTTPURLAllowed(args.URL); err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, method, args.URL, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: extraHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	return marshalJSON(map[string]any{"status": resp.Status, "headers": resp.Header, "body": string(body)})
}

type TimestampParse struct{}

func (t TimestampParse) Name() string { return "timestamp_parse" }
func (t TimestampParse) Run(_ context.Context, input string) (string, error) {
	var args struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	parsed, err := time.Parse(time.RFC3339, args.Value)
	if err != nil {
		sec, parseErr := strconv.ParseInt(args.Value, 10, 64)
		if parseErr != nil {
			return "", err
		}
		parsed = time.Unix(sec, 0)
	}
	return marshalJSON(map[string]any{"unix": parsed.Unix(), "rfc3339": parsed.Format(time.RFC3339), "utc": parsed.UTC().Format(time.RFC3339)})
}

type TimestampFormat struct{}

func (t TimestampFormat) Name() string { return "timestamp_format" }
func (t TimestampFormat) Run(_ context.Context, input string) (string, error) {
	var args struct {
		Unix int64  `json:"unix"`
		TZ   string `json:"tz"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	loc := time.Local
	if args.TZ != "" {
		var err error
		loc, err = time.LoadLocation(args.TZ)
		if err != nil {
			return "", err
		}
	}
	return time.Unix(args.Unix, 0).In(loc).Format(time.RFC3339), nil
}

type FileStat struct{ Policy Policy }

func (t FileStat) Name() string { return "file_stat" }
func (t FileStat) Run(_ context.Context, input string) (string, error) {
	path, err := pathArg(input)
	if err != nil {
		return "", err
	}
	allowed, err := t.Policy.FileAllowed(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(allowed)
	if err != nil {
		return "", err
	}
	return marshalJSON(map[string]any{"path": allowed, "name": info.Name(), "is_dir": info.IsDir(), "size": info.Size(), "mode": info.Mode().String(), "mod_time": info.ModTime().Format(time.RFC3339)})
}

type FileHead struct{ Policy Policy }

func (t FileHead) Name() string { return "file_head" }
func (t FileHead) Run(_ context.Context, input string) (string, error) {
	path, lines, err := pathLinesArgs(input)
	if err != nil {
		return "", err
	}
	return readHead(t.Policy, path, lines)
}

type FileTail struct{ Policy Policy }

func (t FileTail) Name() string { return "file_tail" }
func (t FileTail) Run(_ context.Context, input string) (string, error) {
	path, lines, err := pathLinesArgs(input)
	if err != nil {
		return "", err
	}
	allowed, err := t.Policy.FileAllowed(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(allowed)
	if err != nil {
		return "", err
	}
	all := strings.Split(string(data), "\n")
	if lines > len(all) {
		lines = len(all)
	}
	return strings.Join(all[len(all)-lines:], "\n"), nil
}

type PathInfo struct{ Policy Policy }

func (t PathInfo) Name() string { return "path_info" }
func (t PathInfo) Run(_ context.Context, input string) (string, error) {
	path, err := pathArg(input)
	if err != nil {
		return "", err
	}
	allowed, err := t.Policy.FileAllowed(path)
	if err != nil {
		return "", err
	}
	return marshalJSON(map[string]any{"input": path, "absolute": allowed, "base": filepath.Base(allowed), "dir": filepath.Dir(allowed), "ext": filepath.Ext(allowed), "clean": filepath.Clean(allowed)})
}

func readCSVInput(input string) ([][]string, error) {
	var args struct{ Text, Delimiter string }
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return nil, err
	}
	reader := csv.NewReader(strings.NewReader(args.Text))
	if args.Delimiter != "" {
		runes := []rune(args.Delimiter)
		reader.Comma = runes[0]
	}
	return reader.ReadAll()
}

func pathArg(input string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", err
	}
	if strings.TrimSpace(args.Path) == "" {
		return "", fmt.Errorf("path is required")
	}
	return args.Path, nil
}

func pathLinesArgs(input string) (string, int, error) {
	var args struct {
		Path  string `json:"path"`
		Lines int    `json:"lines"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", 0, err
	}
	if strings.TrimSpace(args.Path) == "" {
		return "", 0, fmt.Errorf("path is required")
	}
	if args.Lines <= 0 {
		args.Lines = 20
	}
	if args.Lines > 500 {
		args.Lines = 500
	}
	return args.Path, args.Lines, nil
}

func readHead(policy Policy, path string, lines int) (string, error) {
	allowed, err := policy.FileAllowed(path)
	if err != nil {
		return "", err
	}
	file, err := os.Open(allowed)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var buf bytes.Buffer
	scanner := bufio.NewScanner(file)
	for i := 0; i < lines && scanner.Scan(); i++ {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(scanner.Text())
	}
	return buf.String(), scanner.Err()
}

func marshalJSON(value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
