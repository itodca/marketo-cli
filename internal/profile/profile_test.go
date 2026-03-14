package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListProfilesIncludesDefaultAndNamedProfiles(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configDir := filepath.Join(homeDir, ".config", ConfigDirName)
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

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles returned error: %v", err)
	}

	expected := []string{"default", "sandbox"}
	if len(profiles) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, profiles)
	}
	for index, profileName := range expected {
		if profiles[index] != profileName {
			t.Fatalf("expected %v, got %v", expected, profiles)
		}
	}
}

func TestResolveProfileUsesProjectFileBeforeDefault(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, ProfileFileName), []byte("sandbox\n"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	resolved, err := ResolveProfile("", cwd, func(string) string { return "" })
	if err != nil {
		t.Fatalf("ResolveProfile returned error: %v", err)
	}
	if resolved != "sandbox" {
		t.Fatalf("expected sandbox, got %s", resolved)
	}
}
