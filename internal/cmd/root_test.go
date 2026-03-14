package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAuthListPrintsConfiguredProfiles(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configDir := filepath.Join(homeDir, ".config", "mrkto")
	profilesDir := filepath.Join(configDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte("MARKETO_CLIENT_ID=id\n"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(profilesDir, "sandbox"), []byte("MARKETO_CLIENT_ID=id\n"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runtime := &Runtime{
		Stdout: &stdout,
		Stderr: &stderr,
		Getenv: func(string) string { return "" },
	}

	cmd := NewRootCmd(runtime)
	cmd.SetArgs([]string{"auth", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	expected := "[\n  {\n    \"profile\": \"default\"\n  },\n  {\n    \"profile\": \"sandbox\"\n  }\n]\n"
	if stdout.String() != expected {
		t.Fatalf("expected %q, got %q", expected, stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}
