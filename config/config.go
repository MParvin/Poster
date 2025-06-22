package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"log"
)

// Config holds all configuration for the application
type Config struct {
	GitHubUsername           string
	GitHubToken              string
	TelegramBotToken         string
	TelegramChatID           string // Added this
	TwitterAPIKey            string
	TwitterAPISecretKey      string
	TwitterAccessToken       string
	TwitterAccessTokenSecret string
	MastodonServer           string
	MastodonAccessToken      string
	DevToAPIKey              string
	LinkedInAccessToken      string // Simplified to AccessToken for now
	PostsRepoURL             string // e.g., "https://github.com/MY_USERNAME/my_posts.git"
	PostsRepoPath            string // Local path to clone the posts repo
}

// LoadConfig loads configuration from .env file or environment variables
func LoadConfig() (*Config, error) {
	// Attempt to load .env file from the project root.
	// Useful for local development.
	// Get the directory of the currently running file (config.go)
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	projectRoot := filepath.Join(basePath, "..") // Assumes config is one level down from project root

	// Construct path to .env file
	dotenvPath := filepath.Join(projectRoot, ".env")

	if _, err := os.Stat(dotenvPath); err == nil {
		err = godotenv.Load(dotenvPath)
		if err != nil {
			log.Printf("Warning: Could not load .env file from %s: %v\n", dotenvPath, err)
		} else {
			log.Printf("Loaded configuration from %s\n", dotenvPath)
		}
	} else {
		log.Printf("Info: .env file not found at %s, relying on environment variables.\n", dotenvPath)
	}


	cfg := &Config{
		GitHubUsername:           os.Getenv("GITHUB_USERNAME"),
		GitHubToken:              os.Getenv("GITHUB_TOKEN"),
		TelegramBotToken:         os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:           os.Getenv("TELEGRAM_CHAT_ID"), // Added this
		TwitterAPIKey:            os.Getenv("TWITTER_API_KEY"),
		TwitterAPISecretKey:      os.Getenv("TWITTER_API_SECRET_KEY"),
		TwitterAccessToken:       os.Getenv("TWITTER_ACCESS_TOKEN"),
		TwitterAccessTokenSecret: os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
		MastodonServer:           os.Getenv("MASTODON_SERVER"),
		MastodonAccessToken:      os.Getenv("MASTODON_ACCESS_TOKEN"),
		DevToAPIKey:              os.Getenv("DEV_TO_API_KEY"),
		LinkedInAccessToken:      os.Getenv("LINKEDIN_ACCESS_TOKEN"),
		PostsRepoURL:             os.Getenv("POSTS_REPO_URL"),
		PostsRepoPath:            os.Getenv("POSTS_REPO_PATH"),
	}

	if cfg.PostsRepoPath == "" {
		// Default posts repo path if not set
		defaultPath := filepath.Join(projectRoot, "my_posts_repo")
		cfg.PostsRepoPath = defaultPath
		log.Printf("POSTS_REPO_PATH not set, defaulting to: %s\n", defaultPath)
	}

	// Basic validation for essential configs
	if cfg.GitHubToken == "" {
		log.Println("Warning: GITHUB_TOKEN is not set.")
	}
	if cfg.PostsRepoURL == "" {
		log.Println("Warning: POSTS_REPO_URL is not set.")
	}


	return cfg, nil
}
