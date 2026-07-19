package posting

import (
	"fmt"
	"log"
	"strings"

	"github.com/MParvin/Poster/config"
	"github.com/MParvin/Poster/github"
	"github.com/MParvin/Poster/state"
)

// PublishResult summarizes a publish attempt across platforms.
type PublishResult struct {
	Attempted int
	Succeeded int
	Skipped   int
	Failures  []string
}

// PublishPost sends content to every outbound social platform.
func PublishPost(post github.Post, cfg *config.Config, store *state.Store) error {
	result, err := PublishPostDetailed(post, cfg, store)
	if err != nil {
		return err
	}
	if len(result.Failures) > 0 {
		return fmt.Errorf("one or more platforms failed: %s", strings.Join(result.Failures, "; "))
	}
	return nil
}

// PublishPostDetailed publishes with structured result metadata.
func PublishPostDetailed(post github.Post, cfg *config.Config, store *state.Store) (*PublishResult, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}
	if !cfg.HasOutboundPlatform() {
		return nil, fmt.Errorf("no outbound social platforms configured; refusing to publish or advance state")
	}

	result := &PublishResult{}
	type job struct {
		platform string
		run      func(content string) error
	}

	enabled := map[string]struct{}{}
	for _, name := range cfg.OutboundPlatforms() {
		enabled[name] = struct{}{}
	}

	jobs := make([]job, 0, 5)
	if _, ok := enabled["telegram"]; ok {
		jobs = append(jobs, job{
			platform: "telegram",
			run: func(content string) error {
				return PostToTelegram(content, cfg.TelegramBotToken, cfg.TelegramChatID)
			},
		})
	}
	if _, ok := enabled["twitter"]; ok {
		jobs = append(jobs, job{
			platform: "twitter",
			run: func(content string) error {
				return PostToTwitter(content, TwitterCredentials{
					APIKey:            cfg.TwitterAPIKey,
					APISecretKey:      cfg.TwitterAPISecretKey,
					AccessToken:       cfg.TwitterAccessToken,
					AccessTokenSecret: cfg.TwitterAccessTokenSecret,
				})
			},
		})
	}
	if _, ok := enabled["mastodon"]; ok {
		jobs = append(jobs, job{
			platform: "mastodon",
			run: func(content string) error {
				if err := ValidateMastodonServer(cfg.MastodonServer); err != nil {
					return err
				}
				return PostToMastodon(content, cfg.MastodonServer, cfg.MastodonAccessToken)
			},
		})
	}
	if _, ok := enabled["devto"]; ok {
		jobs = append(jobs, job{
			platform: "devto",
			run: func(content string) error {
				return PostToDevTo(content, cfg.DevToAPIKey)
			},
		})
	}
	if _, ok := enabled["linkedin"]; ok {
		jobs = append(jobs, job{
			platform: "linkedin",
			run: func(content string) error {
				return PostToLinkedIn(content, cfg.LinkedInAccessToken, cfg.LinkedInPersonURN)
			},
		})
	}

	for _, j := range jobs {
		if store != nil && store.IsDelivered(post.CommitSHA, post.Path, j.platform) {
			log.Printf("Skipping %s for %s@%s (already delivered)", j.platform, post.Path, shortSHA(post.CommitSHA))
			result.Skipped++
			continue
		}

		content := AdaptForPlatform(j.platform, post.Content)
		result.Attempted++

		if cfg.DryRun {
			log.Printf("[DRY-RUN] Would post to %s (%d chars) for %s@%s", j.platform, len(content), post.Path, shortSHA(post.CommitSHA))
			result.Succeeded++
			continue
		}

		if err := j.run(content); err != nil {
			result.Failures = append(result.Failures, fmt.Sprintf("%s: %v", j.platform, err))
			continue
		}
		result.Succeeded++
		if store != nil {
			store.MarkDelivered(post.CommitSHA, post.Path, j.platform)
		}
	}

	return result, nil
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}
