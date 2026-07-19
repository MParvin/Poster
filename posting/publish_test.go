package posting

import (
	"strings"
	"testing"

	"github.com/MParvin/Poster/config"
	"github.com/MParvin/Poster/github"
	"github.com/MParvin/Poster/state"
)

func TestPublishPostRequiresPlatform(t *testing.T) {
	_, err := PublishPostDetailed(github.Post{Content: "hi"}, &config.Config{Source: config.SourceTelegram}, nil)
	if err == nil || !strings.Contains(err.Error(), "no outbound social platforms") {
		t.Fatalf("expected no-platform error, got %v", err)
	}
}

func TestPublishPostDryRunSkipsNetworkAndLedger(t *testing.T) {
	cfg := &config.Config{
		TelegramBotToken: "123:abc",
		TelegramChatID:   "@channel",
		DryRun:           true,
	}
	store := &state.Store{Deliveries: map[string]string{}}
	result, err := PublishPostDetailed(github.Post{
		CommitSHA: "abc",
		Path:      "p.md",
		Content:   "hello",
	}, cfg, store)
	if err != nil {
		t.Fatal(err)
	}
	if result.Succeeded != 1 || result.Attempted != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if store.IsDelivered("abc", "p.md", "telegram") {
		t.Fatal("dry-run must not mark deliveries")
	}
}

func TestAdaptForPlatformTruncatesTwitter(t *testing.T) {
	long := strings.Repeat("a", 500)
	got := AdaptForPlatform("twitter", long)
	if len([]rune(got)) != 280 {
		t.Fatalf("len=%d", len([]rune(got)))
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis, got %q", got[len(got)-5:])
	}
}

func TestSkipAlreadyDelivered(t *testing.T) {
	cfg := &config.Config{
		TelegramBotToken: "123:abc",
		TelegramChatID:   "@channel",
		DryRun:           true,
	}
	store := &state.Store{Deliveries: map[string]string{}}
	store.MarkDelivered("abc", "p.md", "telegram")
	result, err := PublishPostDetailed(github.Post{
		CommitSHA: "abc",
		Path:      "p.md",
		Content:   "hello",
	}, cfg, store)
	if err != nil {
		t.Fatal(err)
	}
	if result.Skipped != 1 || result.Attempted != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}
