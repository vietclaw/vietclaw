package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// WebSearch searches the web using DuckDuckGo HTML search.
type WebSearch struct{}

func (t WebSearch) Name() string { return "web_search" }

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

func (t WebSearch) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		// Fallback to raw string input if not JSON
		args.Query = strings.TrimSpace(input)
	}

	if args.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(args.Query)
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return "", err
	}

	// Use a standard browser User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("duckduckgo returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyHTML := string(bodyBytes)

	results := ParseDDGHTML(bodyHTML)
	if len(results) == 0 {
		return "[]", nil
	}

	// Limit to top 8 results
	if len(results) > 8 {
		results = results[:8]
	}

	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// ParseDDGHTML parses DuckDuckGo's HTML search page results.
func ParseDDGHTML(htmlContent string) []SearchResult {
	var results []SearchResult

	// DDG HTML search results are typically divs containing class="result"
	chunks := strings.Split(htmlContent, "class=\"result ")
	if len(chunks) <= 1 {
		return results
	}

	reURL := regexp.MustCompile(`href="([^"]+)"`)
	reTag := regexp.MustCompile(`<[^>]*>`)

	for _, chunk := range chunks[1:] {
		// Find title anchor which has class="result__a"
		idxA := strings.Index(chunk, "class=\"result__a\"")
		if idxA == -1 {
			continue
		}
		aBlock := chunk[idxA:]
		endA := strings.Index(aBlock, "</a>")
		if endA == -1 {
			continue
		}
		titleHTML := aBlock[strings.Index(aBlock, ">")+1 : endA]
		title := cleanHTMLText(titleHTML, reTag)

		// Extract raw URL from the same anchor block
		matches := reURL.FindStringSubmatch(aBlock)
		if len(matches) < 2 {
			continue
		}
		rawURL := matches[1]
		resolvedURL := cleanDDGRedirect(rawURL)

		// Extract snippet
		idxSnippet := strings.Index(chunk, "class=\"result__snippet\"")
		var snippet string
		if idxSnippet != -1 {
			snippetBlock := chunk[idxSnippet:]
			endSnippet := strings.Index(snippetBlock, "</a>")
			if endSnippet != -1 {
				snippetHTML := snippetBlock[strings.Index(snippetBlock, ">")+1 : endSnippet]
				snippet = cleanHTMLText(snippetHTML, reTag)
			}
		}

		if resolvedURL != "" && title != "" {
			results = append(results, SearchResult{
				Title:   title,
				URL:     resolvedURL,
				Snippet: snippet,
			})
		}
	}

	return results
}

func cleanDDGRedirect(rawURL string) string {
	if strings.Contains(rawURL, "uddg=") {
		parts := strings.Split(rawURL, "uddg=")
		if len(parts) > 1 {
			subParts := strings.Split(parts[1], "&")
			if decoded, err := url.QueryUnescape(subParts[0]); err == nil {
				return decoded
			}
		}
	}
	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}
	return rawURL
}

func cleanHTMLText(input string, reTag *regexp.Regexp) string {
	text := reTag.ReplaceAllString(input, "")
	text = html.UnescapeString(text)
	return strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
}

// WebFetch fetches the clean text content from a web URL.
type WebFetch struct {
	Policy Policy
}

func (t WebFetch) Name() string { return "web_fetch" }

func (t WebFetch) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		args.URL = strings.TrimSpace(input)
	}

	if args.URL == "" {
		return "", fmt.Errorf("url is required")
	}
	if err := t.Policy.HTTPURLAllowed(args.URL); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", args.URL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Limit reading to 2MB to prevent memory exhaustion
	limitReader := io.LimitReader(resp.Body, 2*1024*1024)
	bodyBytes, err := io.ReadAll(limitReader)
	if err != nil {
		return "", err
	}

	cleanText := StripHTML(string(bodyBytes))

	// Truncate clean text to a safe length for LLMs to consume (e.g. 15000 characters)
	if len(cleanText) > 15000 {
		cleanText = cleanText[:15000] + "\n...[content truncated]..."
	}

	return cleanText, nil
}

func StripHTML(htmlContent string) string {
	// Strip script tags and their content
	reScript := regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	htmlContent = reScript.ReplaceAllString(htmlContent, "")

	// Strip style tags and their content
	reStyle := regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	htmlContent = reStyle.ReplaceAllString(htmlContent, "")

	// Strip all other HTML tags
	reTags := regexp.MustCompile(`<[^>]*>`)
	text := reTags.ReplaceAllString(htmlContent, " ")

	// Normalize whitespace
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Replace multiple spaces/tabs inside a single line
			reSpaces := regexp.MustCompile(`\s+`)
			normalizedLine := reSpaces.ReplaceAllString(trimmed, " ")
			cleanedLines = append(cleanedLines, normalizedLine)
		}
	}

	return html.UnescapeString(strings.Join(cleanedLines, "\n"))
}
