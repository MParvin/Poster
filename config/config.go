package config

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	SourceGitHub   = "github"
	SourceTelegram = "telegram"
)

// Config holds all configuration for the application.
type Config struct {
	Source                   string
	GitHubUsername           string
	GitHubToken              string
	TelegramBotToken         string
	TelegramChatID           string
	TelegramAllowedChatIDs   []string
	TwitterAPIKey            string
	TwitterAPISecretKey      string
	TwitterAccessToken       string
	TwitterAccessTokenSecret string
	MastodonServer           string
	MastodonAccessToken      string
	DevToAPIKey              string
	LinkedInAccessToken      string
	LinkedInPersonURN        string
	PostsRepoURL             string
	PostsRepoPath            string
	StateFilePath            string
	DeliveriesFilePath       string
	ProcessInterval          time.Duration
	DryRun                   bool
	PublishAddedOnly         bool
	// TelegramRepost enables posting back to TELEGRAM_CHAT_ID when SOURCE=telegram.
	TelegramRepost bool
}

// Exact defaults from .env.example / common template copies.
var exactDefaultValues = map[string]struct{}{
	"your_github_username":                              {},
	"your_github_personal_access_token_with_repo_scope": {},
	"https://github.com/my_username/my_posts.git":       {},
	"https://github.com/your_username/my_posts.git":     {},
	"your_telegram_bot_token":                           {},
	"your_telegram_chat_id_or_@channelusername":         {},
	"your_twitter_api_key":                              {},
	"your_twitter_api_secret_key":                       {},
	"your_twitter_access_token":                         {},
	"your_twitter_access_token_secret":                  {},
	"https://your_mastodon_instance.social":             {},
	"your_mastodon_access_token":                        {},
	"your_dev_to_api_key":                               {},
	"your_linkedin_access_token":                        {},
	"urn:li:person:your_linkedin_member_id":             {},
}

// LoadConfig loads configuration from .env file or environment variables.
func LoadConfig() (*Config, error) {
	loadDotEnv()

	source := strings.ToLower(strings.TrimSpace(os.Getenv("SOURCE")))
	if source == "" {
		source = SourceGitHub
	}

	cfg := &Config{
		Source:                   source,
		GitHubUsername:           strings.TrimSpace(os.Getenv("GITHUB_USERNAME")),
		GitHubToken:              strings.TrimSpace(os.Getenv("GITHUB_TOKEN")),
		TelegramBotToken:         strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		TelegramChatID:           strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")),
		TelegramAllowedChatIDs:   splitCSV(os.Getenv("TELEGRAM_ALLOWED_CHAT_IDS")),
		TwitterAPIKey:            strings.TrimSpace(os.Getenv("TWITTER_API_KEY")),
		TwitterAPISecretKey:      strings.TrimSpace(os.Getenv("TWITTER_API_SECRET_KEY")),
		TwitterAccessToken:       strings.TrimSpace(os.Getenv("TWITTER_ACCESS_TOKEN")),
		TwitterAccessTokenSecret: strings.TrimSpace(os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")),
		MastodonServer:           strings.TrimSpace(os.Getenv("MASTODON_SERVER")),
		MastodonAccessToken:      strings.TrimSpace(os.Getenv("MASTODON_ACCESS_TOKEN")),
		DevToAPIKey:              strings.TrimSpace(os.Getenv("DEV_TO_API_KEY")),
		LinkedInAccessToken:      strings.TrimSpace(os.Getenv("LINKEDIN_ACCESS_TOKEN")),
		LinkedInPersonURN:        strings.TrimSpace(os.Getenv("LINKEDIN_PERSON_URN")),
		PostsRepoURL:             strings.TrimSpace(os.Getenv("POSTS_REPO_URL")),
		PostsRepoPath:            strings.TrimSpace(os.Getenv("POSTS_REPO_PATH")),
		StateFilePath:            strings.TrimSpace(os.Getenv("STATE_FILE_PATH")),
		DeliveriesFilePath:       strings.TrimSpace(os.Getenv("DELIVERIES_FILE_PATH")),
		PublishAddedOnly:         envBool("PUBLISH_ADDED_ONLY", true),
		DryRun:                   envBool("DRY_RUN", false),
		TelegramRepost:           envBool("TELEGRAM_REPOST", false),
	}

	// Convenience: if allowlist empty, accept TELEGRAM_CHAT_ID as the only inbound source chat.
	if len(cfg.TelegramAllowedChatIDs) == 0 && cfg.TelegramChatID != "" && !strings.HasPrefix(cfg.TelegramChatID, "@") {
		cfg.TelegramAllowedChatIDs = []string{cfg.TelegramChatID}
	}

	if interval := strings.TrimSpace(os.Getenv("PROCESS_INTERVAL")); interval != "" {
		d, err := time.ParseDuration(interval)
		if err != nil {
			return nil, fmt.Errorf("invalid PROCESS_INTERVAL %q: %w", interval, err)
		}
		if d < 0 {
			return nil, fmt.Errorf("PROCESS_INTERVAL must be >= 0")
		}
		cfg.ProcessInterval = d
	}

	if cfg.PostsRepoPath == "" {
		if _, err := os.Stat("/app/my_posts_data"); err == nil {
			cfg.PostsRepoPath = "/app/my_posts_data"
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				cwd = "."
			}
			cfg.PostsRepoPath = filepath.Join(cwd, "my_posts_repo")
		}
		if cfg.Source == SourceGitHub {
			log.Printf("POSTS_REPO_PATH not set, defaulting to: %s", cfg.PostsRepoPath)
		}
	}

	if err := cfg.rejectDefaultValues(); err != nil {
		return nil, err
	}

	cfg.PostsRepoURL = cfg.resolvedPostsRepoURL()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func splitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func loadDotEnv() {
	candidates := make([]string, 0, 4)
	if explicit := strings.TrimSpace(os.Getenv("ENV_FILE")); explicit != "" {
		candidates = append(candidates, explicit)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, ".env"))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}
	candidates = append(candidates, ".env")

	seen := map[string]struct{}{}
	for _, path := range candidates {
		if path == "" {
			continue
		}
		abs := path
		if !filepath.IsAbs(path) {
			if resolved, err := filepath.Abs(path); err == nil {
				abs = resolved
			}
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}

		info, err := os.Stat(abs)
		if err != nil || info.IsDir() {
			continue
		}
		if err := godotenv.Load(abs); err != nil {
			log.Printf("Warning: could not load .env file from %s: %v", abs, err)
			return
		}
		log.Printf("Loaded configuration from %s", abs)
		return
	}
	log.Println("Info: .env file not found, relying on environment variables.")
}

func envBool(name string, defaultValue bool) bool {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultValue
	}
	return v
}

func (c *Config) rejectDefaultValues() error {
	var problems []string

	switch c.Source {
	case SourceGitHub:
		required := []struct {
			name  string
			value string
		}{
			{"GITHUB_USERNAME", c.GitHubUsername},
			{"GITHUB_TOKEN", c.GitHubToken},
			{"POSTS_REPO_URL", c.PostsRepoURL},
		}
		for _, item := range required {
			switch {
			case item.value == "":
				problems = append(problems, fmt.Sprintf("%s is required", item.name))
			case isDefaultValue(item.value):
				problems = append(problems, fmt.Sprintf("%s is still set to a default/placeholder value", item.name))
			}
		}
		if containsPlaceholderUsername(c.PostsRepoURL) &&
			(c.GitHubUsername == "" || isDefaultValue(c.GitHubUsername)) {
			problems = append(problems, "POSTS_REPO_URL still contains MY_USERNAME/YOUR_USERNAME; replace it or set a real GITHUB_USERNAME")
		}
	case SourceTelegram:
		switch {
		case c.TelegramBotToken == "":
			problems = append(problems, "TELEGRAM_BOT_TOKEN is required when SOURCE=telegram")
		case isDefaultValue(c.TelegramBotToken):
			problems = append(problems, "TELEGRAM_BOT_TOKEN is still set to a default/placeholder value")
		}
		if len(c.TelegramAllowedChatIDs) == 0 {
			problems = append(problems, "TELEGRAM_ALLOWED_CHAT_IDS is required when SOURCE=telegram (numeric chat IDs)")
		}
	default:
		problems = append(problems, fmt.Sprintf("SOURCE must be %q or %q", SourceGitHub, SourceTelegram))
	}

	optional := []struct {
		name  string
		value *string
	}{
		{"TELEGRAM_BOT_TOKEN", &c.TelegramBotToken},
		{"TELEGRAM_CHAT_ID", &c.TelegramChatID},
		{"TWITTER_API_KEY", &c.TwitterAPIKey},
		{"TWITTER_API_SECRET_KEY", &c.TwitterAPISecretKey},
		{"TWITTER_ACCESS_TOKEN", &c.TwitterAccessToken},
		{"TWITTER_ACCESS_TOKEN_SECRET", &c.TwitterAccessTokenSecret},
		{"MASTODON_SERVER", &c.MastodonServer},
		{"MASTODON_ACCESS_TOKEN", &c.MastodonAccessToken},
		{"DEV_TO_API_KEY", &c.DevToAPIKey},
		{"LINKEDIN_ACCESS_TOKEN", &c.LinkedInAccessToken},
		{"LINKEDIN_PERSON_URN", &c.LinkedInPersonURN},
	}
	for _, item := range optional {
		if *item.value != "" && isDefaultValue(*item.value) {
			problems = append(problems, fmt.Sprintf("%s is still set to a default/placeholder value (set a real value or leave it empty)", item.name))
		}
	}

	if len(problems) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("invalid .env configuration; replace default values before starting:\n")
	for _, problem := range problems {
		b.WriteString("  - ")
		b.WriteString(problem)
		b.WriteByte('\n')
	}
	b.WriteString("Copy .env.example to .env and update settings for your SOURCE mode.")
	return fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
}

func (c *Config) resolvedPostsRepoURL() string {
	repoURL := c.PostsRepoURL
	if c.GitHubUsername == "" || isDefaultValue(c.GitHubUsername) {
		return repoURL
	}

	replacer := strings.NewReplacer(
		"MY_USERNAME", c.GitHubUsername,
		"YOUR_USERNAME", c.GitHubUsername,
		"your_github_username", c.GitHubUsername,
	)
	return replacer.Replace(repoURL)
}

func isDefaultValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	lower := strings.ToLower(trimmed)
	if _, ok := exactDefaultValues[lower]; ok {
		return true
	}
	if containsPlaceholderUsername(trimmed) {
		return true
	}

	markers := []string{
		"your_",
		"my_username",
		"your_username",
		"changeme",
		"placeholder",
		"token_here",
		"personal_access_token",
		"replace_me",
		"todo_",
	}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func containsPlaceholderUsername(value string) bool {
	upper := strings.ToUpper(value)
	return strings.Contains(upper, "MY_USERNAME") || strings.Contains(upper, "YOUR_USERNAME")
}

func (c *Config) validate() error {
	if c.Source == SourceGitHub && c.PostsRepoURL != "" {
		parsed, err := url.Parse(c.PostsRepoURL)
		if err != nil {
			return fmt.Errorf("invalid POSTS_REPO_URL: %w", err)
		}
		if parsed.Scheme != "https" {
			return fmt.Errorf("POSTS_REPO_URL must use https")
		}
		if parsed.User != nil {
			return fmt.Errorf("POSTS_REPO_URL must not include embedded credentials")
		}
		if containsPlaceholderUsername(c.PostsRepoURL) {
			return fmt.Errorf("POSTS_REPO_URL still contains a placeholder username")
		}
	}

	if c.MastodonServer != "" {
		if err := ValidatePublicHTTPSURL(c.MastodonServer, "MASTODON_SERVER"); err != nil {
			return err
		}
		c.MastodonServer = strings.TrimRight(c.MastodonServer, "/")
	}

	if c.LinkedInAccessToken != "" && c.LinkedInPersonURN == "" {
		return fmt.Errorf("LINKEDIN_ACCESS_TOKEN is set but LINKEDIN_PERSON_URN is missing")
	}

	if c.Source == SourceTelegram && !c.HasOutboundPlatform() {
		return fmt.Errorf("SOURCE=telegram requires at least one outbound platform (twitter, mastodon, devto, linkedin, or telegram with TELEGRAM_REPOST=true)")
	}

	return nil
}

// ValidatePublicHTTPSURL ensures urlStr is https and does not target private/link-local addresses.
func ValidatePublicHTTPSURL(urlStr, fieldName string) error {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", fieldName, err)
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("%s must be a valid https URL", fieldName)
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("%s must include a host", fieldName)
	}
	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") {
		return fmt.Errorf("%s must not target localhost", fieldName)
	}

	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return fmt.Errorf("%s must not target a private or link-local IP", fieldName)
		}
		return nil
	}

	if strings.HasSuffix(lowerHost, ".local") || strings.HasSuffix(lowerHost, ".internal") {
		return fmt.Errorf("%s must not target local/internal hostnames", fieldName)
	}
	return nil
}

func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return false
		}
		if ip4[0] == 0 {
			return false
		}
	}
	return true
}

// ConfiguredPlatforms returns platforms with complete credentials (including telegram outbound).
func (c *Config) ConfiguredPlatforms() []string {
	return c.publishPlatforms(true)
}

// OutboundPlatforms returns platforms that should receive content for the current source.
func (c *Config) OutboundPlatforms() []string {
	includeTelegram := c.Source != SourceTelegram || c.TelegramRepost
	return c.publishPlatforms(includeTelegram)
}

func (c *Config) publishPlatforms(includeTelegram bool) []string {
	var platforms []string
	if includeTelegram && c.HasTelegram() {
		platforms = append(platforms, "telegram")
	}
	if c.HasTwitter() {
		platforms = append(platforms, "twitter")
	}
	if c.HasMastodon() {
		platforms = append(platforms, "mastodon")
	}
	if c.HasDevTo() {
		platforms = append(platforms, "devto")
	}
	if c.HasLinkedIn() {
		platforms = append(platforms, "linkedin")
	}
	return platforms
}

// HasAnyPlatform reports whether at least one social platform is configured.
func (c *Config) HasAnyPlatform() bool {
	return len(c.ConfiguredPlatforms()) > 0
}

// HasOutboundPlatform reports whether there is somewhere to publish.
func (c *Config) HasOutboundPlatform() bool {
	return len(c.OutboundPlatforms()) > 0
}

// HasTelegram reports whether Telegram outbound posting is configured.
func (c *Config) HasTelegram() bool {
	return c.TelegramBotToken != "" && c.TelegramChatID != ""
}

// HasTwitter reports whether Twitter posting is configured.
func (c *Config) HasTwitter() bool {
	return c.TwitterAPIKey != "" && c.TwitterAPISecretKey != "" &&
		c.TwitterAccessToken != "" && c.TwitterAccessTokenSecret != ""
}

// HasMastodon reports whether Mastodon posting is configured.
func (c *Config) HasMastodon() bool {
	return c.MastodonServer != "" && c.MastodonAccessToken != ""
}

// HasDevTo reports whether Dev.to posting is configured.
func (c *Config) HasDevTo() bool {
	return c.DevToAPIKey != ""
}

// HasLinkedIn reports whether LinkedIn posting is configured.
func (c *Config) HasLinkedIn() bool {
	return c.LinkedInAccessToken != "" && c.LinkedInPersonURN != ""
}
