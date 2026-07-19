package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MParvin/Poster/config"
	"github.com/MParvin/Poster/github"
	"github.com/MParvin/Poster/posting"
	"github.com/MParvin/Poster/state"
	"github.com/MParvin/Poster/telegram"
	"github.com/spf13/cobra"
)

var (
	dryRunFlag bool
	onceFlag   bool
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Fetch new content and publish it to configured social platforms.",
	Long: `Primary workflow depends on SOURCE:

  SOURCE=github   (default) Clone/pull a posts repo and publish new markdown files.
  SOURCE=telegram Poll the Telegram bot inbox (getUpdates) and publish new messages.

Telegram mode is designed for GitHub Actions: Telegram keeps unconfirmed updates
for up to 24 hours, so no database is required if the job runs at least hourly
and confirms updates only after a successful publish.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cfg == nil {
			log.Println("Error: configuration not loaded.")
			os.Exit(1)
		}
		if dryRunFlag {
			cfg.DryRun = true
		}

		if err := runProcessLoop(); err != nil {
			log.Printf("Error: %v", err)
			os.Exit(1)
		}
	},
}

func runProcessLoop() error {
	interval := cfg.ProcessInterval
	if onceFlag {
		interval = 0
	}

	if err := runProcessOnce(); err != nil {
		return err
	}
	if interval <= 0 {
		return nil
	}

	log.Printf("Process interval enabled (%s); waiting for next run.", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	for {
		select {
		case <-ticker.C:
			if err := runProcessOnce(); err != nil {
				log.Printf("Process iteration failed: %v", err)
			}
		case sig := <-signals:
			log.Printf("Received %s; shutting down.", sig)
			return nil
		}
	}
}

func runProcessOnce() error {
	switch cfg.Source {
	case config.SourceTelegram:
		return runTelegramOnce()
	case config.SourceGitHub:
		return runGitHubOnce()
	default:
		return fmt.Errorf("unsupported SOURCE %q", cfg.Source)
	}
}

func runTelegramOnce() error {
	if !cfg.HasOutboundPlatform() {
		return fmt.Errorf("no outbound social platforms configured")
	}

	stateFilePath := state.ResolveStateFilePath(cfg.StateFilePath, ".state")
	deliveriesPath := state.ResolveDeliveriesPath(cfg.DeliveriesFilePath, ".state", stateFilePath)
	lockPath := state.ResolveLockPath(stateFilePath, ".state")

	lock, err := state.AcquireLock(lockPath)
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := lock.Release(); releaseErr != nil {
			log.Printf("Warning: failed to release lock: %v", releaseErr)
		}
	}()

	store, err := state.LoadStore(stateFilePath, deliveriesPath)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	log.Println("Starting Telegram inbox processing...")
	log.Printf("Allowlist chat IDs: %v", cfg.TelegramAllowedChatIDs)
	log.Printf("Outbound platforms: %v", cfg.OutboundPlatforms())
	if cfg.DryRun {
		log.Println("Dry-run mode enabled; no network publishes will be sent and updates will not be confirmed.")
	}

	if err := telegram.DeleteWebhook(cfg.TelegramBotToken); err != nil {
		log.Printf("Warning: deleteWebhook failed: %v", err)
	}

	messages, err := telegram.FetchPendingMessages(cfg.TelegramBotToken, cfg.TelegramAllowedChatIDs)
	if err != nil {
		return fmt.Errorf("fetch telegram updates: %w", err)
	}
	if len(messages) == 0 {
		log.Println("No new Telegram messages to process.")
		return nil
	}

	log.Printf("Found %d Telegram message(s).", len(messages))
	var lastConfirmed int64
	for i, msg := range messages {
		post := github.Post{
			CommitSHA: "tg:" + strconv.FormatInt(msg.UpdateID, 10),
			Path:      fmt.Sprintf("chat/%s/msg/%d", msg.ChatID, msg.MessageID),
			Content:   msg.Text,
		}
		log.Printf("Publishing telegram message %d from chat %s (%d chars, update_id=%d)...",
			i+1, msg.ChatID, len(msg.Text), msg.UpdateID)

		result, err := posting.PublishPostDetailed(post, cfg, store)
		if err != nil {
			return err
		}
		if !cfg.DryRun {
			if err := store.SaveDeliveries(deliveriesPath); err != nil {
				return fmt.Errorf("saving deliveries: %w", err)
			}
		}
		if len(result.Failures) > 0 {
			log.Printf("Error publishing telegram message %d: %s", i+1, joinFailures(result.Failures))
			if lastConfirmed > 0 && !cfg.DryRun {
				if err := telegram.ConfirmThrough(cfg.TelegramBotToken, lastConfirmed); err != nil {
					return fmt.Errorf("confirm successful updates after failure: %w", err)
				}
				log.Printf("Confirmed Telegram updates through update_id=%d", lastConfirmed)
			}
			return fmt.Errorf("publishing incomplete; remaining Telegram updates left unconfirmed for retry")
		}

		lastConfirmed = msg.UpdateID
		// Persist a lightweight checkpoint for operators; Telegram offset is the source of truth.
		if !cfg.DryRun {
			_ = store.SaveCheckpoint(stateFilePath, post.CommitSHA)
		}
	}

	if cfg.DryRun {
		log.Printf("[DRY-RUN] Would confirm Telegram updates through update_id=%d", lastConfirmed)
	} else if lastConfirmed > 0 {
		if err := telegram.ConfirmThrough(cfg.TelegramBotToken, lastConfirmed); err != nil {
			return fmt.Errorf("confirm telegram updates: %w", err)
		}
		if err := store.SaveDeliveries(deliveriesPath); err != nil {
			return fmt.Errorf("saving deliveries: %w", err)
		}
		log.Printf("Confirmed Telegram updates through update_id=%d", lastConfirmed)
	}

	log.Println("Telegram processing finished.")
	return nil
}

func runGitHubOnce() error {
	if cfg.PostsRepoURL == "" {
		return fmt.Errorf("POSTS_REPO_URL is not configured")
	}
	if !cfg.HasOutboundPlatform() {
		return fmt.Errorf("no social platforms configured; set at least one platform before processing")
	}

	stateFilePath := state.ResolveStateFilePath(cfg.StateFilePath, cfg.PostsRepoPath)
	deliveriesPath := state.ResolveDeliveriesPath(cfg.DeliveriesFilePath, cfg.PostsRepoPath, stateFilePath)
	lockPath := state.ResolveLockPath(stateFilePath, cfg.PostsRepoPath)

	lock, err := state.AcquireLock(lockPath)
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := lock.Release(); releaseErr != nil {
			log.Printf("Warning: failed to release lock: %v", releaseErr)
		}
	}()

	store, err := state.LoadStore(stateFilePath, deliveriesPath)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	log.Println("Starting processing...")
	log.Printf("Repository local path: %s", cfg.PostsRepoPath)
	log.Printf("State file: %s", stateFilePath)
	log.Printf("Deliveries file: %s", deliveriesPath)
	log.Printf("Configured platforms: %v", cfg.OutboundPlatforms())
	if cfg.DryRun {
		log.Println("Dry-run mode enabled; no network publishes will be sent.")
	}

	if err := github.CloneOrPullRepo(cfg.PostsRepoURL, cfg.PostsRepoPath, cfg.GitHubToken); err != nil {
		return fmt.Errorf("repository operation: %w", err)
	}
	log.Println("Repository operation successful.")

	log.Printf("Checking for new posts. Last known commit SHA: %q", store.LastSHA)
	newPosts, latestSHA, err := github.GetNewPosts(cfg.PostsRepoPath, store.LastSHA, cfg.PublishAddedOnly)
	if err != nil {
		return fmt.Errorf("checking for new posts: %w", err)
	}

	if len(newPosts) == 0 {
		if store.LastSHA == "" && latestSHA != "" {
			if cfg.DryRun {
				log.Printf("[DRY-RUN] Would bootstrap last processed commit SHA to %s", latestSHA)
			} else if err := store.SaveCheckpoint(stateFilePath, latestSHA); err != nil {
				return fmt.Errorf("saving bootstrap state: %w", err)
			} else {
				log.Printf("Bootstrapped last processed commit SHA to %s", latestSHA)
			}
		} else {
			log.Println("No new posts found to process.")
		}
		log.Println("Processing finished.")
		return nil
	}

	log.Printf("Found %d new post(s).", len(newPosts))
	var publishFailed bool
	for i, post := range newPosts {
		log.Printf("Publishing post %d %s@%s (%d characters)...", i+1, post.Path, short(post.CommitSHA), len(post.Content))
		result, err := posting.PublishPostDetailed(post, cfg, store)
		if err != nil {
			return err
		}
		if !cfg.DryRun {
			if err := store.SaveDeliveries(deliveriesPath); err != nil {
				return fmt.Errorf("saving deliveries: %w", err)
			}
		}
		if len(result.Failures) > 0 {
			log.Printf("Error publishing post %d: %s", i+1, joinFailures(result.Failures))
			publishFailed = true
			break
		}
	}

	if publishFailed {
		return fmt.Errorf("publishing incomplete; checkpoint not advanced (successful platform deliveries were recorded)")
	}

	if cfg.DryRun {
		log.Printf("[DRY-RUN] Would update last processed commit SHA to %s", latestSHA)
	} else {
		if err := store.SaveCheckpoint(stateFilePath, latestSHA); err != nil {
			return fmt.Errorf("saving processed commit SHA: %w", err)
		}
		if err := store.SaveDeliveries(deliveriesPath); err != nil {
			return fmt.Errorf("saving deliveries: %w", err)
		}
		log.Printf("Updated last processed commit SHA to %s", latestSHA)
	}
	log.Println("Processing finished.")
	return nil
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

func joinFailures(failures []string) string {
	return strings.Join(failures, "; ")
}

func init() {
	processCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Log actions without publishing or advancing state")
	processCmd.Flags().BoolVar(&onceFlag, "once", false, "Run a single iteration even if PROCESS_INTERVAL is set")
	rootCmd.AddCommand(processCmd)
}
