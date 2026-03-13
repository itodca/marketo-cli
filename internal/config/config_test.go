package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReadsDefaultProfileKeyValueConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configDir := filepath.Join(homeDir, ".config", "mrkto")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte(
		"MARKETO_MUNCHKIN_ID=123-ABC-456\nMARKETO_CLIENT_ID=test-id\nMARKETO_CLIENT_SECRET=test-secret\n",
	), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	cfg, err := Load("", t.TempDir(), func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Profile != "default" {
		t.Fatalf("expected default profile, got %s", cfg.Profile)
	}
	if cfg.RestURL != "https://123-ABC-456.mktorest.com/rest" {
		t.Fatalf("unexpected rest URL: %s", cfg.RestURL)
	}
	if cfg.IdentityURL != "https://123-ABC-456.mktorest.com/identity" {
		t.Fatalf("unexpected identity URL: %s", cfg.IdentityURL)
	}
}

func TestLoadEnvOverridesFileValues(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configDir := filepath.Join(homeDir, ".config", "mrkto")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte(
		"MARKETO_MUNCHKIN_ID=123-ABC-456\nMARKETO_CLIENT_ID=file-id\nMARKETO_CLIENT_SECRET=file-secret\n",
	), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	env := map[string]string{
		"MARKETO_CLIENT_ID":     "env-id",
		"MARKETO_CLIENT_SECRET": "env-secret",
	}
	cfg, err := Load("", t.TempDir(), func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ClientID != "env-id" || cfg.ClientSecret != "env-secret" {
		t.Fatalf("expected env overrides, got %#v", cfg)
	}
}

func TestLoadReturnsHelpfulErrorWhenProfileFileIsMissing(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	_, err := Load("", t.TempDir(), func(string) string { return "" })
	if err == nil {
		t.Fatal("expected missing profile error")
	}

	message := err.Error()
	if !strings.Contains(message, "Expected config at") {
		t.Fatalf("expected missing-file guidance, got %q", message)
	}
	if !strings.Contains(message, filepath.Join(homeDir, ".config", "mrkto", "config")) {
		t.Fatalf("expected config path in error, got %q", message)
	}
}

func TestLoadUsesEnvWithoutProfileFile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	env := map[string]string{
		"MARKETO_MUNCHKIN_ID":   "123-ABC-456",
		"MARKETO_CLIENT_ID":     "env-id",
		"MARKETO_CLIENT_SECRET": "env-secret",
	}

	cfg, err := Load("", t.TempDir(), func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ClientID != "env-id" || cfg.ClientSecret != "env-secret" || cfg.MunchkinID != "123-ABC-456" {
		t.Fatalf("expected env-backed config, got %#v", cfg)
	}
}
