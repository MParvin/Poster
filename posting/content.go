package posting

import (
	"strings"
	"unicode/utf8"
)

const (
	maxTelegramRunes  = 4096
	maxTwitterRunes   = 280
	maxMastodonRunes  = 500
	maxLinkedInRunes  = 3000
)

// AdaptForPlatform returns platform-appropriate text for publishing.
func AdaptForPlatform(platform, content string) string {
	content = strings.TrimSpace(content)
	switch platform {
	case "telegram":
		return truncateRunes(content, maxTelegramRunes)
	case "twitter":
		return truncateRunes(stripMarkdownLight(content), maxTwitterRunes)
	case "mastodon":
		return truncateRunes(stripMarkdownLight(content), maxMastodonRunes)
	case "linkedin":
		return truncateRunes(stripMarkdownLight(content), maxLinkedInRunes)
	case "devto":
		return content
	default:
		return content
	}
}

func stripMarkdownLight(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#"):
			trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		case strings.HasPrefix(trimmed, "```"):
			continue
		case strings.HasPrefix(trimmed, ">"):
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
		}
		trimmed = strings.ReplaceAll(trimmed, "**", "")
		trimmed = strings.ReplaceAll(trimmed, "__", "")
		trimmed = strings.ReplaceAll(trimmed, "*", "")
		trimmed = strings.ReplaceAll(trimmed, "_", "")
		trimmed = strings.ReplaceAll(trimmed, "`", "")
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, "\n")
}

func truncateRunes(text string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(text) <= limit {
		return text
	}
	if limit <= 3 {
		return string([]rune(text)[:limit])
	}
	runes := []rune(text)
	return string(runes[:limit-3]) + "..."
}

func smartTruncate(text string, length int) string {
	return truncateRunes(text, length)
}
