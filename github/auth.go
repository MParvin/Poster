package github

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// gitAuthEnv prepares a temporary HOME with a 0600 .netrc so git can authenticate
// without putting the token on the process argv (unlike http.extraHeader / URL userinfo).
// The token exists only in that temporary .netrc for the duration of the git command.
func gitAuthEnv(token string) (env []string, cleanup func(), err error) {
	cleanup = func() {}
	if token == "" {
		return nil, cleanup, nil
	}

	dir, err := os.MkdirTemp("", "social-poster-git-home-*")
	if err != nil {
		return nil, cleanup, fmt.Errorf("create git auth temp dir: %w", err)
	}

	netrcPath := filepath.Join(dir, ".netrc")
	content := "machine github.com\nlogin x-access-token\npassword " + token + "\n" +
		"machine www.github.com\nlogin x-access-token\npassword " + token + "\n"
	if err := os.WriteFile(netrcPath, []byte(content), 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return nil, cleanup, fmt.Errorf("write netrc: %w", err)
	}

	cleanup = func() { _ = os.RemoveAll(dir) }
	env = []string{
		"HOME=" + dir,
		"GIT_TERMINAL_PROMPT=0",
	}
	return env, cleanup, nil
}

func validateHTTPSGitHubURL(repoURL string) (string, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse repository URL: %w", err)
	}
	if parsedURL.Scheme != "https" {
		return "", fmt.Errorf("repository URL must use https, got %q", parsedURL.Scheme)
	}
	if parsedURL.Host == "" {
		return "", fmt.Errorf("repository URL must include a host")
	}
	host := strings.ToLower(parsedURL.Host)
	if _, ok := allowedRepoHosts[host]; !ok {
		return "", fmt.Errorf("repository host %q is not allowed", parsedURL.Host)
	}
	if parsedURL.User != nil {
		return "", fmt.Errorf("repository URL must not include embedded credentials")
	}
	if strings.Trim(parsedURL.Path, "/") == "" {
		return "", fmt.Errorf("repository URL must include an owner and repository name")
	}

	clean := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   parsedURL.Path,
	}
	return clean.String(), nil
}

func redactURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if parsed.User != nil {
		parsed.User = url.UserPassword("oauth2", "REDACTED")
	}
	return parsed.String()
}
