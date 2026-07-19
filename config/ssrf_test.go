package config

import "testing"

func TestValidatePublicHTTPSURL(t *testing.T) {
	if err := ValidatePublicHTTPSURL("https://mastodon.social", "MASTODON_SERVER"); err != nil {
		t.Fatal(err)
	}
	blocked := []string{
		"http://mastodon.social",
		"https://127.0.0.1",
		"https://localhost",
		"https://10.0.0.5",
		"https://192.168.1.1",
		"https://169.254.169.254",
		"https://something.local",
	}
	for _, raw := range blocked {
		if err := ValidatePublicHTTPSURL(raw, "MASTODON_SERVER"); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestHasAnyPlatform(t *testing.T) {
	cfg := &Config{Source: SourceGitHub}
	if cfg.HasAnyPlatform() {
		t.Fatal("expected false")
	}
	cfg.DevToAPIKey = "k"
	if !cfg.HasAnyPlatform() {
		t.Fatal("expected true")
	}
}
