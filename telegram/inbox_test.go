package telegram

import "testing"

func TestExtractTextMessage(t *testing.T) {
	msg, ok := extractTextMessage(update{
		UpdateID: 10,
		Message: &message{
			MessageID: 3,
			Text:      "Hello world",
			Chat:      &chat{ID: 42},
			From:      &user{Username: "alice"},
		},
	})
	if !ok {
		t.Fatal("expected message")
	}
	if msg.ChatID != "42" || msg.Text != "Hello world" || msg.FromUser != "alice" {
		t.Fatalf("unexpected %#v", msg)
	}

	if _, ok := extractTextMessage(update{
		UpdateID: 11,
		Message:  &message{Text: "/start", Chat: &chat{ID: 1}},
	}); ok {
		t.Fatal("expected /start to be ignored")
	}
}

func TestIsAllowedChat(t *testing.T) {
	allow := normalizeAllowlist([]string{"123", " 456 "})
	if !isAllowedChat(allow, "123") {
		t.Fatal("expected 123 allowed")
	}
	if isAllowedChat(allow, "999") {
		t.Fatal("expected 999 denied")
	}
}
