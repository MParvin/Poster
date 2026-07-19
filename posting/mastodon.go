package posting

import (
	"fmt"
	"log"

	"github.com/MParvin/Poster/config"
)

// PostToMastodon sends a toot to a Mastodon instance.
func PostToMastodon(messageContent, serverURL, accessToken string) error {
	if serverURL == "" {
		return fmt.Errorf("mastodon server URL is empty")
	}
	if accessToken == "" {
		return fmt.Errorf("mastodon access token is empty")
	}

	endpoint := serverURL + "/api/v1/statuses"
	payload := map[string]string{
		"status": messageContent,
	}

	log.Printf("[MASTODON] Posting to %s (%d chars, token %s)", serverURL, len(messageContent), truncateToken(accessToken))
	_, _, err := doJSONRequest("POST", endpoint, map[string]string{
		"Authorization": "Bearer " + accessToken,
	}, payload)
	if err != nil {
		return fmt.Errorf("mastodon post failed: %w", err)
	}

	log.Println("[MASTODON] Toot sent successfully.")
	return nil
}

// ValidateMastodonServer ensures the server URL is safe to call.
func ValidateMastodonServer(serverURL string) error {
	return config.ValidatePublicHTTPSURL(serverURL, "MASTODON_SERVER")
}
