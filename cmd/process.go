package cmd

import (
	"log"
	"os" // For os.Exit if critical errors occur

	"github.com/MY_USERNAME/social_poster/github" // Corrected import path
	"github.com/spf13/cobra"
)

// This would ideally be stored persistently (e.g., in a file, database)
var lastProcessedCommitSHA string

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Fetches new posts from the configured GitHub repository and posts them.",
	Long: `This command initiates the primary workflow:
1. Clones or pulls the latest changes from the GitHub posts repository.
2. Checks for new commits/posts since the last run.
3. For each new post, it (will eventually) trigger posting to configured social media.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cfg == nil {
			log.Println("Error: Configuration not loaded.")
			os.Exit(1) // Or handle error more gracefully
		}

		log.Println("Starting processing...")

		// Validate essential configurations for this command
		if cfg.PostsRepoURL == "" {
			log.Println("Error: POSTS_REPO_URL is not configured. Cannot process.")
			return // Or os.Exit(1)
		}
		// GitHubToken can be empty for public repos, but CloneOrPullRepo will use it if present.
		// Let CloneOrPullRepo handle errors related to auth.

		log.Printf("Using repository URL: %s", cfg.PostsRepoURL)
		log.Printf("Repository local path: %s", cfg.PostsRepoPath)

		// 1. Clone or Pull Repository
		err := github.CloneOrPullRepo(cfg.PostsRepoURL, cfg.PostsRepoPath, cfg.GitHubToken)
		if err != nil {
			log.Printf("Error with repository operation: %v", err)
			// Depending on the error, might want to exit or just log and skip further processing
			return
		}
		log.Println("Repository operation successful.")

		// 2. Check for new posts
		// In a real app, lastProcessedCommitSHA would be loaded from a persistent store.
		// For now, it's an in-memory variable, so it resets each run.
		// This means GetNewPosts will always fetch "new" posts based on its current logic.
		log.Printf("Checking for new posts. Last known commit SHA: '%s'", lastProcessedCommitSHA)
		newPosts, latestSHA, err := github.GetNewPosts(cfg.PostsRepoPath, lastProcessedCommitSHA)
		if err != nil {
			log.Printf("Error checking for new posts: %v", err)
			return
		}

		if len(newPosts) > 0 {
			log.Printf("Found %d new post(s):", len(newPosts))
			for i, post := range newPosts {
				log.Printf("Post %d: %s", i+1, post)
				// TODO: Here, we would call the functions to post to social media
				// e.g., posting.PostToTelegram(post, cfg.TelegramBotToken)
				//      posting.PostToTwitter(post, twitterConfig)
				//      ...and so on for other platforms.
			}
			lastProcessedCommitSHA = latestSHA
			log.Printf("Updated last processed commit SHA to: %s", lastProcessedCommitSHA)
			// TODO: Persist lastProcessedCommitSHA here
		} else {
			log.Println("No new posts found to process.")
		}

		log.Println("Processing finished.")
	},
}

func init() {
	rootCmd.AddCommand(processCmd)
	// Here you can define flags specific to processCmd if needed
	// processCmd.Flags().StringVarP(&someFlag, "someflag", "s", "", "A description for someflag")
}
