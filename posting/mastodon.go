package posting

import (
	"fmt"
	"log"
)

// PostToMastodon sends a toot to a Mastodon instance.
// This is a placeholder implementation.
func PostToMastodon(messageContent string, serverURL string, accessToken string) error {
	if serverURL == "" {
		return fmt.Errorf("mastodon server URL is empty")
	}
	if accessToken == "" {
		return fmt.Errorf("mastodon access token is empty")
	}

	log.Printf("[MASTODON] Posting toot to %s: \"%s\" (using token %s...)",
		serverURL,
		messageContent,
		truncateToken(accessToken))

	// Placeholder: Simulate API call
	log.Println("[MASTODON] Toot sent successfully (simulated).")
	return nil
}
