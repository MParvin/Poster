package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MParvin/Poster/config"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "social_poster",
	Short: "Monitor a Git repository and post updates to social media.",
	Long: `Social Poster monitors a specified private GitHub repository for new markdown posts.
When new content is detected, it formats and posts it to configured social media channels
like Telegram, Twitter, Mastodon, Dev.to, and LinkedIn.

Credentials and configuration are managed via an .env file or environment variables.
Optional ENV_FILE overrides the .env path.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			log.Printf("Configuration error:\n%v", err)
			return fmt.Errorf("failed to load configuration: replace default .env values and restart")
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
