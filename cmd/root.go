/*
Copyright © 2025 Jules AI <jules@example.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/MY_USERNAME/social_poster/config"
	"github.com/spf13/cobra"
	"log"
)

var cfg *config.Config
var envFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "social_poster",
	Short: "A tool to monitor a Git repository and post updates to social media.",
	Long: `Social Poster monitors a specified private GitHub repository for new commits/posts.
When new content is detected, it formats and posts it to configured social media channels
like Telegram, Twitter, Mastodon, Dev.to, and LinkedIn.

Credentials and configuration are managed via an .env file or environment variables.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Note: godotenv.Load in LoadConfig already handles .env path finding.
		// If we wanted to make the .env file path configurable via a flag,
		// we would pass `envFile` to `godotenv.Load` or a custom LoadConfig.
		// For now, config.LoadConfig() has its own logic to find .env.
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		if cfg.PostsRepoURL == "" || cfg.GitHubToken == "" {
			log.Println("Warning: POSTS_REPO_URL or GITHUB_TOKEN is not set in the configuration.")
			// Depending on the command, this might be a fatal error.
			// For now, just a warning, specific commands can check required fields.
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// No global flags defined here for now, .env loading is handled by config.LoadConfig()
	// If we wanted to make .env path configurable:
	// cobra.OnInitialize(initConfig)
	// rootCmd.PersistentFlags().StringVar(&envFile, "env", ".env", "Path to .env file")
}

// initConfig reads in config file and ENV variables if set.
// func initConfig() {
//  This is where you might use viper or similar, but we're using godotenv directly in LoadConfig.
// }
