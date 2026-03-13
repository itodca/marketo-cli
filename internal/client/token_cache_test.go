package client

import (
	"path/filepath"
	"testing"
	"time"
)

func TestTokenCacheSaveAndLoadUsesPythonCompatibleShape(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	cache := NewTokenCache(t.TempDir())
	cache.Now = func() time.Time { return now }

	if err := cache.Save("default", "token-123", 3600); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	token, ok, err := cache.Load("default")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected cached token to be valid")
	}
	if token.AccessToken != "token-123" {
		t.Fatalf("expected token-123, got %s", token.AccessToken)
	}

	path, err := cache.PathForProfile("default")
	if err != nil {
		t.Fatalf("PathForProfile returned error: %v", err)
	}
	expectedPath := filepath.Join(cache.Dir, "token-default.json")
	if path != expectedPath {
		t.Fatalf("expected %s, got %s", expectedPath, path)
	}
}
