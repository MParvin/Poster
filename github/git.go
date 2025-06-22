package github

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// CloneOrPullRepo clones a repository if it doesn't exist locally, or pulls changes if it does.
// It uses the provided token for authentication with HTTPS.
func CloneOrPullRepo(repoURL, localPath, token string) error {
	// Construct the authenticated URL
	// Example: https://oauth2:TOKEN@github.com/user/repo.git
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL %s: %w", repoURL, err)
	}

	// Ensure token is not empty for private repos
	if token == "" && !strings.HasPrefix(parsedURL.Scheme, "http") { // crude check for public repo
		log.Println("Warning: GitHub token is empty. Operations on private repositories will likely fail.")
		// For public repos, token might not be needed for clone, but will be for private.
		// Let it proceed, git will fail if auth is needed and not provided.
	}

	authenticatedURL := ""
	if token != "" {
		authenticatedURL = fmt.Sprintf("%s://oauth2:%s@%s%s", parsedURL.Scheme, token, parsedURL.Host, parsedURL.Path)
	} else {
		authenticatedURL = repoURL // Use original URL if no token (for public repos)
	}


	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		// Path does not exist, clone the repo
		log.Printf("Cloning repository %s into %s...\n", repoURL, localPath)
		cmd := exec.Command("git", "clone", authenticatedURL, localPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %w. Ensure git is installed and configured.", err)
		}
		log.Printf("Repository cloned successfully to %s.\n", localPath)
	} else if err == nil {
		// Path exists, pull the latest changes
		log.Printf("Repository exists at %s. Pulling latest changes...\n", localPath)
		cmd := exec.Command("git", "-C", localPath, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to pull repository: %w", err)
		}
		log.Println("Repository pulled successfully.")
	} else {
		// Some other error accessing localPath
		return fmt.Errorf("error checking local path %s: %w", localPath, err)
	}

	return nil
}

// GetNewCommits (Placeholder for now)
// This function would determine new posts/commits since the last check.
// For now, it's a placeholder.
// A real implementation might involve:
// - Storing the last processed commit SHA.
// - Using `git log <last_sha>..HEAD` to find new commits.
// - Parsing commit messages or checking file changes in those commits.
func GetNewPosts(repoPath string, lastProcessedCommitSHA string) ([]string, string, error) {
	log.Printf("Checking for new posts in %s since commit %s (placeholder implementation)\n", repoPath, lastProcessedCommitSHA)

	// Placeholder: Simulate finding one new post.
	// In reality, you'd use git commands to list commits and identify changes.
	// For example: git log --pretty=oneline <lastSHA>..HEAD

	// Simulate getting the latest commit SHA
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current HEAD SHA: %w", err)
	}
	currentHEAD := strings.TrimSpace(string(out))

	if lastProcessedCommitSHA == currentHEAD {
		log.Println("No new commits found.")
		return []string{}, currentHEAD, nil
	}

	// This is a mock. A real version would inspect files in commits between lastProcessedCommitSHA and currentHEAD
	mockPosts := []string{"This is a new mock post from commit " + currentHEAD}
	log.Printf("Found %d new post(s). New HEAD is %s.\n", len(mockPosts), currentHEAD)

	return mockPosts, currentHEAD, nil
}
