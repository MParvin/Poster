package posting

import (
	"fmt"
	"log"
	"strings"
)

// PostToDevTo creates a post on Dev.to.
func PostToDevTo(markdownContent, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("dev.to API key is empty")
	}

	title, body := splitMarkdownTitle(markdownContent)
	payload := map[string]any{
		"article": map[string]any{
			"title":          title,
			"body_markdown":  body,
			"published":      true,
		},
	}

	log.Printf("[DEV.TO] Creating post %q (%d chars, key %s)", smartTruncate(title, 40), len(markdownContent), truncateToken(apiKey))
	_, _, err := doJSONRequest("POST", "https://dev.to/api/articles", map[string]string{
		"api-key": apiKey,
	}, payload)
	if err != nil {
		return fmt.Errorf("dev.to post failed: %w", err)
	}

	log.Println("[DEV.TO] Post created successfully.")
	return nil
}

func splitMarkdownTitle(markdownContent string) (string, string) {
	lines := strings.Split(markdownContent, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			body := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
			if title == "" {
				title = "Untitled"
			}
			return title, body
		}
	}

	title := smartTruncate(strings.TrimSpace(markdownContent), 80)
	if title == "" {
		title = "Untitled"
	}
	return title, markdownContent
}
