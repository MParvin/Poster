package github

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var allowedRepoHosts = map[string]struct{}{
	"github.com":     {},
	"www.github.com": {},
}

// Post is a markdown file discovered in a new commit.
type Post struct {
	CommitSHA string
	Path      string
	Content   string
}

// CloneOrPullRepo clones a repository if it does not exist locally, or pulls changes if it does.
func CloneOrPullRepo(repoURL, localPath, token string) error {
	cleanURL, err := validateRepoURL(repoURL)
	if err != nil {
		return err
	}
	if err := validateLocalPath(localPath); err != nil {
		return err
	}

	authEnv, cleanup, err := gitAuthEnv(token)
	if err != nil {
		return err
	}
	defer cleanup()

	if !isGitRepository(localPath) {
		if err := ensureEmptyOrMissing(localPath); err != nil {
			return err
		}
		log.Printf("Cloning repository %s into %s...", redactURL(cleanURL), localPath)
		if err := runGit(authEnv, "clone", cleanURL, localPath); err != nil {
			return fmt.Errorf("failed to clone repository: %w. Ensure git is installed and configured", err)
		}
		log.Printf("Repository cloned successfully to %s.", localPath)
		return nil
	}

	if err := ensureCleanRemoteURL(localPath, cleanURL, authEnv); err != nil {
		return err
	}

	log.Printf("Repository exists at %s. Pulling latest changes...", localPath)
	if err := runGit(authEnv, "-C", localPath, "pull", "--ff-only"); err != nil {
		return fmt.Errorf("failed to pull repository: %w", err)
	}
	log.Println("Repository pulled successfully.")
	return nil
}

// GetNewPosts returns markdown posts from commits after lastProcessedCommitSHA.
// On first run (empty lastProcessedCommitSHA), it bootstraps to HEAD without returning posts.
// When addedOnly is true, only newly added markdown files are published (not modifications).
func GetNewPosts(repoPath, lastProcessedCommitSHA string, addedOnly bool) ([]Post, string, error) {
	if err := validateLocalPath(repoPath); err != nil {
		return nil, "", err
	}

	currentHEAD, err := gitOutput(repoPath, "rev-parse", "HEAD")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current HEAD SHA: %w", err)
	}

	if lastProcessedCommitSHA == "" {
		log.Println("No previous commit SHA stored; bootstrapping to current HEAD without posting historical content.")
		return []Post{}, currentHEAD, nil
	}

	if lastProcessedCommitSHA == currentHEAD {
		log.Println("No new commits found.")
		return []Post{}, currentHEAD, nil
	}

	commitSHAs, err := gitLines(repoPath, "log", "--reverse", "--pretty=format:%H", lastProcessedCommitSHA+"..HEAD")
	if err != nil {
		return nil, "", fmt.Errorf("failed to list new commits: %w", err)
	}
	if len(commitSHAs) == 0 {
		return []Post{}, currentHEAD, nil
	}

	seen := make(map[string]struct{})
	var posts []Post

	for _, commitSHA := range commitSHAs {
		changedFiles, err := listChangedMarkdown(repoPath, commitSHA, addedOnly)
		if err != nil {
			return nil, "", err
		}

		for _, filePath := range changedFiles {
			key := commitSHA + ":" + filePath
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			content, err := gitOutput(repoPath, "show", commitSHA+":"+filePath)
			if err != nil {
				log.Printf("Warning: could not read %s at %s: %v", filePath, commitSHA, err)
				continue
			}
			content = strings.TrimSpace(content)
			if content == "" {
				continue
			}
			posts = append(posts, Post{
				CommitSHA: commitSHA,
				Path:      filePath,
				Content:   content,
			})
		}
	}

	log.Printf("Found %d new post(s). New HEAD is %s.", len(posts), currentHEAD)
	return posts, currentHEAD, nil
}

func listChangedMarkdown(repoPath, commitSHA string, addedOnly bool) ([]string, error) {
	args := []string{"diff-tree", "--no-commit-id", "--name-only", "-r"}
	if addedOnly {
		args = []string{"diff-tree", "--no-commit-id", "--name-only", "--diff-filter=A", "-r", commitSHA}
	} else {
		args = append(args, commitSHA)
	}

	changedFiles, err := gitLines(repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list files for commit %s: %w", commitSHA, err)
	}

	var markdown []string
	for _, filePath := range changedFiles {
		if isMarkdownPost(filePath) {
			markdown = append(markdown, filePath)
		}
	}
	return markdown, nil
}

func validateRepoURL(repoURL string) (string, error) {
	return validateHTTPSGitHubURL(repoURL)
}

func validateLocalPath(localPath string) error {
	if localPath == "" {
		return fmt.Errorf("local repository path is empty")
	}
	if !filepath.IsAbs(localPath) {
		absPath, err := filepath.Abs(localPath)
		if err != nil {
			return fmt.Errorf("resolve local path: %w", err)
		}
		localPath = absPath
	}
	cleanPath := filepath.Clean(localPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("local repository path %q is not allowed", localPath)
	}
	return nil
}

func isGitRepository(localPath string) bool {
	gitDir := filepath.Join(localPath, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func ensureEmptyOrMissing(localPath string) error {
	info, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error checking local path %s: %w", localPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("local repository path %s exists but is not a directory", localPath)
	}

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("read local path %s: %w", localPath, err)
	}
	if len(entries) == 0 {
		return nil
	}

	return fmt.Errorf("local path %s exists but is not a git repository; remove its contents or use a fresh volume before cloning", localPath)
}

func ensureCleanRemoteURL(localPath, cleanURL string, authEnv []string) error {
	currentRemote, err := gitOutput(localPath, "remote", "get-url", "origin")
	if err != nil {
		return nil
	}
	if redactURL(currentRemote) != redactURL(cleanURL) {
		return runGit(authEnv, "-C", localPath, "remote", "set-url", "origin", cleanURL)
	}
	// Also scrub credentials if origin still embeds a userinfo token from older versions.
	if strings.Contains(currentRemote, "@") {
		return runGit(authEnv, "-C", localPath, "remote", "set-url", "origin", cleanURL)
	}
	return nil
}

func runGit(extraEnv []string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Env = append(cmd.Env, extraEnv...)
	return cmd.Run()
}

func gitOutput(repoPath string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command("git", fullArgs...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitLines(repoPath string, args ...string) ([]string, error) {
	output, err := gitOutput(repoPath, args...)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}
	lines := strings.Split(output, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

func isMarkdownPost(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}
