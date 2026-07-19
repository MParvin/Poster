package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadSHA(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".social_poster_state")

	if err := SaveLastProcessedSHA(path, "abc123"); err != nil {
		t.Fatal(err)
	}
	got, err := LoadLastProcessedSHA(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc123" {
		t.Fatalf("got %q", got)
	}
}

func TestDeliveryLedger(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state")
	deliveriesPath := filepath.Join(dir, "deliveries.json")

	store := &Store{Deliveries: map[string]string{}}
	store.MarkDelivered("c1", "post.md", "telegram")
	if err := store.SaveCheckpoint(statePath, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveDeliveries(deliveriesPath); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadStore(statePath, deliveriesPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.LastSHA != "c1" {
		t.Fatalf("sha=%q", loaded.LastSHA)
	}
	if !loaded.IsDelivered("c1", "post.md", "telegram") {
		t.Fatal("expected delivered")
	}
	if loaded.IsDelivered("c1", "post.md", "twitter") {
		t.Fatal("did not expect twitter delivery")
	}
}

func TestAcquireLockExclusive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lock")

	lock1, err := AcquireLock(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock1.Release(); err != nil {
			t.Errorf("release lock: %v", err)
		}
	}()

	if _, err := AcquireLock(path); err == nil {
		t.Fatal("expected second lock to fail")
	}
}

func TestResolvePaths(t *testing.T) {
	if got := ResolveStateFilePath("/x/state", "/repo"); got != "/x/state" {
		t.Fatalf("got %q", got)
	}
	if got := ResolveDeliveriesPath("", "/repo", "/x/state"); got != "/x/state.deliveries.json" {
		t.Fatalf("got %q", got)
	}
	_ = os.TempDir()
}
