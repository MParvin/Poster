package github

import "testing"

func TestValidateHTTPSGitHubURL(t *testing.T) {
	ok, err := validateHTTPSGitHubURL("https://github.com/MParvin/posts.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok != "https://github.com/MParvin/posts.git" {
		t.Fatalf("got %q", ok)
	}

	cases := []string{
		"http://github.com/MParvin/posts.git",
		"https://gitlab.com/MParvin/posts.git",
		"https://oauth2:token@github.com/MParvin/posts.git",
		"https://github.com/",
	}
	for _, raw := range cases {
		if _, err := validateHTTPSGitHubURL(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestIsMarkdownPost(t *testing.T) {
	if !isMarkdownPost("posts/hello.md") {
		t.Fatal("expected markdown")
	}
	if !isMarkdownPost("posts/Hello.MARKDOWN") {
		t.Fatal("expected markdown")
	}
	if isMarkdownPost("posts/hello.txt") {
		t.Fatal("did not expect markdown")
	}
}

func TestRedactURL(t *testing.T) {
	got := redactURL("https://oauth2:secret@github.com/a/b.git")
	if got != "https://oauth2:REDACTED@github.com/a/b.git" {
		t.Fatalf("got %q", got)
	}
}
