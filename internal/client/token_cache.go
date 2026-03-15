package client

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/itodca/marketo-cli/internal/profile"
)

var profileSlugRe = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

type CachedToken struct {
	AccessToken string  `json:"access_token"`
	Expiry      float64 `json:"expiry"`
}

type TokenCache struct {
	Dir string
	Now func() time.Time
}

func NewTokenCache(dir string) *TokenCache {
	return &TokenCache{Dir: dir, Now: time.Now}
}

func (cache *TokenCache) PathForProfile(profileName string) (string, error) {
	if cache.Dir != "" {
		return filepath.Join(cache.Dir, "token-"+profileSlug(profileName)+".json"), nil
	}

	configDir, err := profile.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "token-"+profileSlug(profileName)+".json"), nil
}

func (cache *TokenCache) Load(profileName string) (CachedToken, bool, error) {
	path, err := cache.PathForProfile(profileName)
	if err != nil {
		return CachedToken{}, false, err
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CachedToken{}, false, nil
		}
		return CachedToken{}, false, err
	}

	var token CachedToken
	if err := json.Unmarshal(contents, &token); err != nil {
		return CachedToken{}, false, nil
	}
	if token.AccessToken == "" || token.Expiry <= 0 {
		return CachedToken{}, false, nil
	}
	if token.Expiry <= float64(cache.now().Unix()) {
		return CachedToken{}, false, nil
	}
	return token, true, nil
}

func (cache *TokenCache) Delete(profileName string) error {
	path, err := cache.PathForProfile(profileName)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func (cache *TokenCache) Save(profileName, accessToken string, expiresIn int) error {
	path, err := cache.PathForProfile(profileName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload := CachedToken{
		AccessToken: accessToken,
		Expiry:      float64(cache.now().Add(time.Duration(expiresIn-60) * time.Second).Unix()),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return os.WriteFile(path, encoded, 0o600)
}

func (cache *TokenCache) now() time.Time {
	if cache != nil && cache.Now != nil {
		return cache.Now()
	}
	return time.Now()
}

func profileSlug(profileName string) string {
	if strings := profileSlugRe.ReplaceAllString(profileName, "-"); strings != "" {
		return strings
	}
	return "default"
}
