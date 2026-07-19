package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	apiBase          = "https://api.telegram.org"
	defaultHTTPTimeout = 30 * time.Second
	maxUpdatesPerCall  = 100
)

var httpClient = &http.Client{Timeout: defaultHTTPTimeout}

// IncomingMessage is a text message received by the bot.
type IncomingMessage struct {
	UpdateID  int64
	ChatID    string
	MessageID int64
	Text      string
	FromUser  string
}

type apiResponse[T any] struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      T      `json:"result"`
}

type update struct {
	UpdateID int64    `json:"update_id"`
	Message  *message `json:"message"`
}

type message struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	Caption   string `json:"caption"`
	Chat      *chat  `json:"chat"`
	From      *user  `json:"from"`
}

type chat struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Type     string `json:"type"`
}

type user struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

// DeleteWebhook clears any webhook so getUpdates can be used.
func DeleteWebhook(botToken string) error {
	_, err := callAPI[bool](botToken, "deleteWebhook", url.Values{
		"drop_pending_updates": {"false"},
	})
	return err
}

// FetchPendingMessages returns unconfirmed text messages (up to several pages).
// Messages remain pending until ConfirmThrough is called with a higher offset.
func FetchPendingMessages(botToken string, allowedChatIDs []string) ([]IncomingMessage, error) {
	if botToken == "" {
		return nil, fmt.Errorf("telegram bot token is empty")
	}
	allow := normalizeAllowlist(allowedChatIDs)
	if len(allow) == 0 {
		return nil, fmt.Errorf("telegram allowlist is empty; set TELEGRAM_ALLOWED_CHAT_IDS")
	}

	var all []IncomingMessage
	for {
		updates, err := getUpdates(botToken, 0, maxUpdatesPerCall)
		if err != nil {
			return nil, err
		}
		if len(updates) == 0 {
			break
		}

		batchAdded := 0
		highest := updates[len(updates)-1].UpdateID
		for _, u := range updates {
			msg, ok := extractTextMessage(u)
			if !ok {
				continue
			}
			if !isAllowedChat(allow, msg.ChatID) {
				continue
			}
			all = append(all, msg)
			batchAdded++
		}

		// If the page was full, confirm non-text/non-allowed noise up to highest
		// only when nothing publishable remains in this page would stall the queue.
		// We never confirm here — caller confirms after successful publish.
		// To avoid infinite loops on ignored updates, confirm ignored-only pages.
		if batchAdded == 0 && len(updates) > 0 {
			if err := ConfirmThrough(botToken, highest); err != nil {
				return nil, fmt.Errorf("confirm ignored updates: %w", err)
			}
			if len(updates) < maxUpdatesPerCall {
				break
			}
			continue
		}

		if len(updates) < maxUpdatesPerCall {
			break
		}
		// More pages may exist; without confirming we would loop on the same page.
		// Return what we have; next run (or ConfirmThrough of processed ones) advances.
		break
	}

	return all, nil
}

// ConfirmThrough acknowledges all updates with update_id <= updateID.
func ConfirmThrough(botToken string, updateID int64) error {
	if updateID <= 0 {
		return nil
	}
	_, err := getUpdates(botToken, updateID+1, 1)
	return err
}

func getUpdates(botToken string, offset int64, limit int) ([]update, error) {
	values := url.Values{
		"timeout": {"0"},
		"limit":   {strconv.Itoa(limit)},
		"allowed_updates": {`["message"]`},
	}
	if offset > 0 {
		values.Set("offset", strconv.FormatInt(offset, 10))
	}
	return callAPI[[]update](botToken, "getUpdates", values)
}

func callAPI[T any](botToken, method string, values url.Values) (T, error) {
	var zero T
	endpoint := fmt.Sprintf("%s/bot%s/%s", apiBase, botToken, method)
	if values == nil {
		values = url.Values{}
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return zero, err
	}

	var parsed apiResponse[T]
	if err := json.Unmarshal(body, &parsed); err != nil {
		return zero, fmt.Errorf("decode telegram %s response: %w", method, err)
	}
	if !parsed.OK {
		desc := parsed.Description
		if desc == "" {
			desc = strings.TrimSpace(string(body))
		}
		return zero, fmt.Errorf("telegram %s failed: %s", method, desc)
	}
	return parsed.Result, nil
}

func extractTextMessage(u update) (IncomingMessage, bool) {
	if u.Message == nil || u.Message.Chat == nil {
		return IncomingMessage{}, false
	}
	text := strings.TrimSpace(u.Message.Text)
	if text == "" {
		text = strings.TrimSpace(u.Message.Caption)
	}
	if text == "" {
		return IncomingMessage{}, false
	}
	// Ignore bot commands like /start unless they have trailing content.
	if strings.HasPrefix(text, "/") {
		parts := strings.Fields(text)
		if len(parts) == 1 {
			return IncomingMessage{}, false
		}
	}

	from := ""
	if u.Message.From != nil {
		if u.Message.From.Username != "" {
			from = u.Message.From.Username
		} else {
			from = u.Message.From.FirstName
		}
	}

	return IncomingMessage{
		UpdateID:  u.UpdateID,
		ChatID:    strconv.FormatInt(u.Message.Chat.ID, 10),
		MessageID: u.Message.MessageID,
		Text:      text,
		FromUser:  from,
	}, true
}

func normalizeAllowlist(ids []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if strings.HasPrefix(id, "@") {
			out[strings.ToLower(id)] = struct{}{}
			continue
		}
		out[id] = struct{}{}
	}
	return out
}

func isAllowedChat(allow map[string]struct{}, chatID string) bool {
	if _, ok := allow[chatID]; ok {
		return true
	}
	return false
}
