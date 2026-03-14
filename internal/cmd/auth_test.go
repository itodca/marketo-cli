package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthSetupWritesDefaultProfileFromFlags(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	exitCode, stdout, stderr := executeTestCommand(t, nil, nil, "auth", "setup", "--munchkin-id", "123-ABC-456", "--client-id", "test-id", "--client-secret", "test-secret")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if payload["status"] != "saved" || payload["profile"] != "default" {
		t.Fatalf("unexpected payload: %#v", payload)
	}

	configPath := filepath.Join(homeDir, ".config", "mrkto", "config")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "MARKETO_MUNCHKIN_ID=123-ABC-456\nMARKETO_CLIENT_ID=test-id\nMARKETO_CLIENT_SECRET=test-secret\n" {
		t.Fatalf("unexpected file content: %q", string(content))
	}
}

func TestSetupPromptsForMissingValuesAndWritesNamedProfile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	stdin := bytes.NewBufferString("123-ABC-456\ntest-id\ntest-secret\n")
	exitCode, stdout, stderr := executeTestCommand(t, nil, stdin, "setup", "--profile", "sandbox")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if !strings.Contains(stderr, "Munchkin ID: ") || !strings.Contains(stderr, "Client ID: ") || !strings.Contains(stderr, "Client Secret: ") {
		t.Fatalf("unexpected prompt output: %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if payload["profile"] != "sandbox" {
		t.Fatalf("unexpected payload: %#v", payload)
	}

	configPath := filepath.Join(homeDir, ".config", "mrkto", "profiles", "sandbox")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "MARKETO_MUNCHKIN_ID=123-ABC-456\nMARKETO_CLIENT_ID=test-id\nMARKETO_CLIENT_SECRET=test-secret\n" {
		t.Fatalf("unexpected file content: %q", string(content))
	}
}
