package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/itodca/marketo-cli/internal/profile"
)

type Config struct {
	MunchkinID   string
	ClientID     string
	ClientSecret string
	RestURL      string
	IdentityURL  string
	Profile      string
}

func Load(explicitProfile, cwd string, getenv func(string) string) (Config, error) {
	resolvedProfile, err := profile.ResolveProfile(explicitProfile, cwd, getenv)
	if err != nil {
		return Config{}, err
	}

	filePath, err := profile.ConfigFileFor(resolvedProfile)
	if err != nil {
		return Config{}, err
	}

	fileValues, fileExists, err := readKeyValueFile(filePath)
	if err != nil {
		return Config{}, err
	}

	munchkinID := envOr(getenv, "MARKETO_MUNCHKIN_ID", fileValues["MARKETO_MUNCHKIN_ID"])
	clientID := envOr(getenv, "MARKETO_CLIENT_ID", fileValues["MARKETO_CLIENT_ID"])
	clientSecret := envOr(getenv, "MARKETO_CLIENT_SECRET", fileValues["MARKETO_CLIENT_SECRET"])

	if munchkinID == "" || clientID == "" || clientSecret == "" {
		if !fileExists {
			return Config{}, fmt.Errorf("Marketo credentials not found for profile %q. Expected config at %s or MARKETO_MUNCHKIN_ID, MARKETO_CLIENT_ID, MARKETO_CLIENT_SECRET env vars.", resolvedProfile, filePath)
		}
		return Config{}, fmt.Errorf("Marketo credentials not found (profile: %s). Run 'mrkto setup' or set MARKETO_MUNCHKIN_ID, MARKETO_CLIENT_ID, MARKETO_CLIENT_SECRET as env vars.", resolvedProfile)
	}

	restURL := envOr(getenv, "MARKETO_REST_URL", fileValues["MARKETO_REST_URL"])
	if restURL == "" {
		restURL = fmt.Sprintf("https://%s.mktorest.com/rest", munchkinID)
	}

	identityURL := envOr(getenv, "MARKETO_IDENTITY_URL", fileValues["MARKETO_IDENTITY_URL"])
	if identityURL == "" {
		identityURL = fmt.Sprintf("https://%s.mktorest.com/identity", munchkinID)
	}

	return Config{
		MunchkinID:   munchkinID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RestURL:      restURL,
		IdentityURL:  identityURL,
		Profile:      resolvedProfile,
	}, nil
}

func readKeyValueFile(path string) (map[string]string, bool, error) {
	values := map[string]string{}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return values, false, nil
		}
		return nil, false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(value) == "" {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, true, err
	}

	return values, true, nil
}

func envOr(getenv func(string) string, key, fallback string) string {
	if getenv == nil {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}
	if value := getenv(key); value != "" {
		return value
	}
	return fallback
}
