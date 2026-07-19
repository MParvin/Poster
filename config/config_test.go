package config

import (
	"strings"
	"testing"
)

func TestRejectDefaultValues_RequiredPlaceholders(t *testing.T) {
	cfg := &Config{
		Source:         SourceGitHub,
		GitHubUsername: "your_github_username",
		GitHubToken:    "your_github_personal_access_token_with_repo_scope",
		PostsRepoURL:   "https://github.com/MY_USERNAME/my_posts.git",
	}

	err := cfg.rejectDefaultValues()
	if err == nil {
		t.Fatal("expected error for default GitHub values")
	}

	msg := err.Error()
	for _, want := range []string{
		"GITHUB_USERNAME",
		"GITHUB_TOKEN",
		"POSTS_REPO_URL",
		"invalid .env configuration",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q missing %q", msg, want)
		}
	}
}

func TestRejectDefaultValues_OptionalPlaceholders(t *testing.T) {
	cfg := &Config{
		Source:           SourceGitHub,
		GitHubUsername:   "MParvin",
		GitHubToken:      "ghp_real_token_value_here_1234567890",
		PostsRepoURL:     "https://github.com/MParvin/my_posts.git",
		TelegramBotToken: "your_telegram_bot_token",
		TelegramChatID:   "your_telegram_chat_id_or_@channelusername",
	}

	err := cfg.rejectDefaultValues()
	if err == nil {
		t.Fatal("expected error for default Telegram values")
	}
	if !strings.Contains(err.Error(), "TELEGRAM_BOT_TOKEN") {
		t.Fatalf("expected TELEGRAM_BOT_TOKEN in error, got %v", err)
	}
}

func TestRejectDefaultValues_ValidRequiredOnly(t *testing.T) {
	cfg := &Config{
		Source:         SourceGitHub,
		GitHubUsername: "MParvin",
		GitHubToken:    "ghp_real_token_value_here_1234567890",
		PostsRepoURL:   "https://github.com/MParvin/my_posts.git",
	}

	if err := cfg.rejectDefaultValues(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRejectDefaultValues_TelegramSource(t *testing.T) {
	cfg := &Config{
		Source:                 SourceTelegram,
		TelegramBotToken:       "123456:ABC-DEF",
		TelegramAllowedChatIDs: []string{"42"},
	}
	if err := cfg.rejectDefaultValues(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg.TelegramAllowedChatIDs = nil
	if err := cfg.rejectDefaultValues(); err == nil {
		t.Fatal("expected allowlist error")
	}
}

func TestOutboundPlatformsSkipTelegramByDefault(t *testing.T) {
	cfg := &Config{
		Source:           SourceTelegram,
		TelegramBotToken: "123:abc",
		TelegramChatID:   "99",
		DevToAPIKey:      "k",
	}
	got := cfg.OutboundPlatforms()
	if len(got) != 1 || got[0] != "devto" {
		t.Fatalf("got %v", got)
	}
	cfg.TelegramRepost = true
	got = cfg.OutboundPlatforms()
	if len(got) != 2 {
		t.Fatalf("expected telegram+devto, got %v", got)
	}
}

func TestResolvedPostsRepoURL(t *testing.T) {
	cfg := &Config{
		GitHubUsername: "MParvin",
		PostsRepoURL:   "https://github.com/MY_USERNAME/my_posts.git",
	}
	got := cfg.resolvedPostsRepoURL()
	want := "https://github.com/MParvin/my_posts.git"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestIsDefaultValue(t *testing.T) {
	cases := map[string]bool{
		"":                          false,
		"MParvin":                   false,
		"ghp_abc123":                false,
		"your_github_username":      true,
		"https://github.com/MY_USERNAME/my_posts.git": true,
		"your_telegram_bot_token":   true,
	}
	for value, want := range cases {
		if got := isDefaultValue(value); got != want {
			t.Fatalf("isDefaultValue(%q)=%v want %v", value, got, want)
		}
	}
}
