package profile

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	ConfigDirName   = "mrkto"
	ProfileFileName = ".mrkto-profile"
)

func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", ConfigDirName), nil
}

func ProfilesDir() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "profiles"), nil
}

func ConfigFileFor(profileName string) (string, error) {
	if profileName == "" || profileName == "default" {
		configDir, err := ConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(configDir, "config"), nil
	}

	profilesDir, err := ProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(profilesDir, profileName), nil
}

func ListProfiles() ([]string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	results := map[string]struct{}{}
	if _, err := os.Stat(filepath.Join(configDir, "config")); err == nil {
		results["default"] = struct{}{}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	profilesDir, err := ProfilesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return sortedKeys(results), nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.Type().IsRegular() {
			results[entry.Name()] = struct{}{}
		}
	}

	return sortedKeys(results), nil
}

func FindProfileName(startDir string) (string, error) {
	current := startDir
	for {
		candidate := filepath.Join(current, ProfileFileName)
		contents, err := os.ReadFile(candidate)
		if err == nil {
			name := strings.TrimSpace(string(contents))
			if name != "" {
				return name, nil
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", nil
		}
		current = parent
	}
}

func ResolveProfile(explicit, cwd string, getenv func(string) string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if getenv != nil {
		if envProfile := strings.TrimSpace(getenv("MRKTO_PROFILE")); envProfile != "" {
			return envProfile, nil
		}
	}
	if cwd != "" {
		fileProfile, err := FindProfileName(cwd)
		if err != nil {
			return "", err
		}
		if fileProfile != "" {
			return fileProfile, nil
		}
	}
	return "default", nil
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
