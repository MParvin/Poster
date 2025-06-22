package posting

import (
	"fmt"
	"log"
)

// PostToDevTo creates a post on Dev.to.
// This is a placeholder implementation.
// Dev.to posts are usually Markdown.
func PostToDevTo(markdownContent string, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("dev.to API key is empty")
	}

	log.Printf("[DEV.TO] Creating post: \"%s...\" (using API Key %s...)",
		smartTruncate(markdownContent, 30),
		truncateToken(apiKey))

	// Placeholder: Simulate API call
	log.Println("[DEV.TO] Post created successfully (simulated).")
	return nil
}

// smartTruncate truncates a string to a certain length, adding "..." if truncated.
func smartTruncate(text string, length int) string {
	if len(text) > length {
		return text[:length-3] + "..."
	}
	return text
}
