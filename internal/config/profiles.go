package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	AppName            = "poster"
	DefaultProfileName = "default"
	profilesConfigFile = "config.json"
)

var errInvalidProfileName = errors.New("invalid profile name")

// Profile holds non-secret configuration values.
type Profile struct {
	IGUserID   string `json:"ig_user_id,omitempty"`
	PageID     string `json:"page_id,omitempty"`
	BusinessID string `json:"business_id,omitempty"`
}

type ProfilesFile struct {
	KeyringBackend string             `json:"keyring_backend,omitempty"`
	Profiles       map[string]Profile `json:"profiles,omitempty"`
}

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(base, AppName), nil
}

func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("ensure config dir: %w", err)
	}

	return dir, nil
}

func ProfilesPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, profilesConfigFile), nil
}

func ReadProfiles() (ProfilesFile, error) {
	path, err := ProfilesPath()
	if err != nil {
		return ProfilesFile{}, err
	}

	b, err := os.ReadFile(path) //nolint:gosec // config file path
	if err != nil {
		if os.IsNotExist(err) {
			return ProfilesFile{}, nil
		}

		return ProfilesFile{}, fmt.Errorf("read profiles: %w", err)
	}

	var cfg ProfilesFile
	if err := json.Unmarshal(b, &cfg); err != nil {
		return ProfilesFile{}, fmt.Errorf("parse profiles %s: %w", path, err)
	}

	return cfg, nil
}

func WriteProfiles(cfg ProfilesFile) error {
	_, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	path, err := ProfilesPath()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profiles: %w", err)
	}

	b = append(b, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write profiles: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("commit profiles: %w", err)
	}

	return nil
}

func NormalizeProfileName(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if name == "" {
		return "", fmt.Errorf("%w: empty", errInvalidProfileName)
	}

	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}

		return "", fmt.Errorf("%w: %q", errInvalidProfileName, raw)
	}

	return name, nil
}

func NormalizeProfileNameOrDefault(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return DefaultProfileName, nil
	}

	return NormalizeProfileName(raw)
}

func ListProfiles(cfg ProfilesFile) []string {
	out := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		out = append(out, name)
	}

	sort.Strings(out)

	return out
}
