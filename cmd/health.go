package cmd

import (
	"fmt"
	"os"

	"github.com/MParvin/Poster/config"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Validate configuration and report process health.",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg == nil {
			fmt.Fprintln(os.Stderr, "unhealthy: configuration not loaded")
			os.Exit(1)
		}
		switch cfg.Source {
		case config.SourceGitHub:
			if cfg.PostsRepoURL == "" {
				fmt.Fprintln(os.Stderr, "unhealthy: POSTS_REPO_URL missing")
				os.Exit(1)
			}
		case config.SourceTelegram:
			if cfg.TelegramBotToken == "" {
				fmt.Fprintln(os.Stderr, "unhealthy: TELEGRAM_BOT_TOKEN missing")
				os.Exit(1)
			}
			if len(cfg.TelegramAllowedChatIDs) == 0 {
				fmt.Fprintln(os.Stderr, "unhealthy: TELEGRAM_ALLOWED_CHAT_IDS missing")
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "unhealthy: unsupported SOURCE %q\n", cfg.Source)
			os.Exit(1)
		}
		if !cfg.HasOutboundPlatform() {
			fmt.Fprintln(os.Stderr, "unhealthy: no outbound social platforms configured")
			os.Exit(1)
		}
		fmt.Printf("ok source=%s platforms=%v interval=%s dry_run=%v\n",
			cfg.Source, cfg.OutboundPlatforms(), cfg.ProcessInterval, cfg.DryRun)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
