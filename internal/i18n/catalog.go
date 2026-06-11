package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed locales/*.json
var localeFS embed.FS

var catalog map[Language]map[MessageID]string

func init() {
	catalog = make(map[Language]map[MessageID]string)
	for _, lang := range []Language{LanguageVietnamese, LanguageEnglish} {
		data, err := localeFS.ReadFile("locales/" + string(lang) + ".json")
		if err != nil {
			panic(fmt.Sprintf("i18n: load %s: %v", lang, err))
		}
		m := make(map[MessageID]string)
		if err := json.Unmarshal(data, &m); err != nil {
			panic(fmt.Sprintf("i18n: parse %s: %v", lang, err))
		}
		catalog[lang] = m
	}
}

func Normalize(language string) Language {
	switch Language(strings.ToLower(strings.TrimSpace(language))) {
	case LanguageEnglish:
		return LanguageEnglish
	default:
		return LanguageVietnamese
	}
}

func T(language string, id MessageID, args ...any) string {
	lang := Normalize(language)
	template := catalog[lang][id]
	if template == "" {
		template = catalog[LanguageVietnamese][id]
	}
	if template == "" {
		return string(id)
	}
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func ToolUILabel(language string, toolName string) string {
	normalized := normalizeToolName(toolName)
	id := MessageID("tool.ui." + normalized)
	label := T(language, id)
	if label == string(id) {
		return toolName
	}
	return label
}

func normalizeToolName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
