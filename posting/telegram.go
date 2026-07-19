package posting

import (
	"fmt"
	"log"
)

// PostToTelegram sends a message to a Telegram channel using a bot token.
func PostToTelegram(messageContent, botToken, chatID string) error {
	if botToken == "" {
		return fmt.Errorf("telegram bot token is empty")
	}
	if chatID == "" {
		return fmt.Errorf("telegram chat ID is empty")
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]string{
		"chat_id": chatID,
		"text":    messageContent,
	}

	log.Printf("[TELEGRAM] Posting to chat %s (%d chars, token %s)", chatID, len(messageContent), truncateToken(botToken))
	_, _, err := doJSONRequest("POST", endpoint, nil, payload)
	if err != nil {
		return fmt.Errorf("telegram post failed: %w", err)
	}

	log.Println("[TELEGRAM] Message sent successfully.")
	return nil
}
