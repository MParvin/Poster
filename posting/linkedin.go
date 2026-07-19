package posting

import (
	"fmt"
	"log"
)

// PostToLinkedIn shares an update on LinkedIn.
func PostToLinkedIn(messageContent, accessToken, personURN string) error {
	if accessToken == "" {
		return fmt.Errorf("linkedin access token is empty")
	}
	if personURN == "" {
		return fmt.Errorf("linkedin person URN is empty")
	}

	payload := map[string]any{
		"author":         personURN,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]any{
			"com.linkedin.ugc.ShareContent": map[string]any{
				"shareCommentary": map[string]any{
					"text": messageContent,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]any{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	log.Printf("[LINKEDIN] Sharing update (%d chars, token %s)", len(messageContent), truncateToken(accessToken))
	_, _, err := doJSONRequest("POST", "https://api.linkedin.com/v2/ugcPosts", map[string]string{
		"Authorization":             "Bearer " + accessToken,
		"X-Restli-Protocol-Version": "2.0.0",
	}, payload)
	if err != nil {
		return fmt.Errorf("linkedin post failed: %w", err)
	}

	log.Println("[LINKEDIN] Update shared successfully.")
	return nil
}
