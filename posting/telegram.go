package posting

import (
	"fmt"
	"log"
)

// PostToTelegram sends a message to a Telegram channel using a bot token.
// This is a placeholder implementation.
func PostToTelegram(messageContent string, botToken string, chatID string) error {
	if botToken == "" {
		return fmt.Errorf("telegram bot token is empty")
	}
	if chatID == "" {
		return fmt.Errorf("telegram chat ID is empty")
	}

	log.Printf("[TELEGRAM] Posting to ChatID %s: \"%s\" (using token starting with %s...)",
		chatID,
		messageContent,
		truncateToken(botToken))

	// Placeholder: Simulate API call
	log.Println("[TELEGRAM] Message sent successfully (simulated).")
	return nil
}

// truncateToken is a helper to avoid logging full tokens.
func truncateToken(token string) string {
	if len(token) > 8 {
		return token[:4]
	}
	return "****"
}
