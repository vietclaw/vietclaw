package channels

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"vietclaw/internal/config"
)

func DownloadTextAttachment(ctx context.Context, url, name, contentType string, size int64, cfg config.AttachmentConfig) (Attachment, error) {
	if !AttachmentAllowed(name, contentType, size, cfg) {
		return Attachment{}, fmt.Errorf("attachment not allowed: %s", name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Attachment{}, err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Attachment{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Attachment{}, fmt.Errorf("download attachment %s: %s", name, resp.Status)
	}
	limit := cfg.MaxBytes
	if limit <= 0 {
		limit = config.DefaultAttachmentMaxBytes
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return Attachment{}, err
	}
	if int64(len(data)) > limit {
		return Attachment{}, fmt.Errorf("attachment too large: %s", name)
	}
	return Attachment{Name: name, ContentType: contentType, Size: int64(len(data)), Text: string(data)}, nil
}

func AttachmentAllowed(name, contentType string, size int64, cfg config.AttachmentConfig) bool {
	if !cfg.Enabled {
		return false
	}
	limit := cfg.MaxBytes
	if limit <= 0 {
		limit = config.DefaultAttachmentMaxBytes
	}
	if size > limit {
		return false
	}
	lowerType := strings.ToLower(contentType)
	if strings.HasPrefix(lowerType, "text/") || strings.Contains(lowerType, "json") || strings.Contains(lowerType, "xml") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		base := strings.ToLower(filepath.Base(name))
		ext = "." + base
	}
	allowed := cfg.AllowedExtensions
	if len(allowed) == 0 {
		allowed = config.DefaultTextAttachmentExtensions()
	}
	for _, item := range allowed {
		if strings.EqualFold(ext, item) {
			return true
		}
	}
	return false
}

func PromptWithAttachments(text string, attachments []Attachment) string {
	prompt := strings.TrimSpace(text)
	for _, att := range attachments {
		content := strings.TrimSpace(att.Text)
		if content == "" {
			continue
		}
		if prompt != "" {
			prompt += "\n\n"
		}
		prompt += fmt.Sprintf("Attachment: %s\nContent-Type: %s\nSize: %d bytes\n```text\n%s\n```", att.Name, att.ContentType, att.Size, content)
	}
	return prompt
}
