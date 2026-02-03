package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/mahmoud/igpostercli/internal/secrets"
)

const (
	DefaultGraphVersion = "v19.0"
	DefaultPollInterval = 5 * time.Second
	DefaultPollTimeout  = 300 * time.Second
)

const (
	envAppID        = "IG_APP_ID"
	envAppSecret    = "IG_APP_SECRET" // #nosec G101 -- env var name, not a credential
	envAccessToken  = "IG_ACCESS_TOKEN"
	envPageID       = "IG_PAGE_ID"
	envBusinessID   = "IG_BUSINESS_ID"
	envIGUserID     = "IG_USER_ID"
	envGraphVersion = "IG_GRAPH_VERSION"
	envPollInterval = "IG_POLL_INTERVAL"
	envPollTimeout  = "IG_POLL_TIMEOUT"
)

type Config struct {
	AppID        string
	AppSecret    string
	AccessToken  string
	PageID       string
	BusinessID   string
	IGUserID     string
	GraphVersion string
	PollInterval time.Duration
	PollTimeout  time.Duration
}

var errConfigNil = errors.New("config is nil")

type MissingEnvError struct {
	Missing []string
}

func (e *MissingEnvError) Error() string {
	if e == nil || len(e.Missing) == 0 {
		return ""
	}

	return fmt.Sprintf("missing required environment variables: %s", strings.Join(e.Missing, ", "))
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	cfg := &Config{
		AppID:        os.Getenv(envAppID),
		AppSecret:    os.Getenv(envAppSecret),
		AccessToken:  os.Getenv(envAccessToken),
		PageID:       os.Getenv(envPageID),
		BusinessID:   os.Getenv(envBusinessID),
		IGUserID:     os.Getenv(envIGUserID),
		GraphVersion: DefaultGraphVersion,
		PollInterval: DefaultPollInterval,
		PollTimeout:  DefaultPollTimeout,
	}

	if v := os.Getenv(envGraphVersion); v != "" {
		cfg.GraphVersion = v
	}

	if v := os.Getenv(envPollInterval); v != "" {
		interval, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", envPollInterval, err)
		}
		cfg.PollInterval = interval
	}

	if v := os.Getenv(envPollTimeout); v != "" {
		timeout, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", envPollTimeout, err)
		}
		cfg.PollTimeout = timeout
	}

	return cfg, nil
}

func LoadWithProfile(profile string) (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	name, err := NormalizeProfileNameOrDefault(profile)
	if err != nil {
		return nil, err
	}

	profiles, err := ReadProfiles()
	if err != nil {
		return nil, err
	}

	if p, ok := profiles.Profiles[name]; ok {
		if p.IGUserID != "" {
			cfg.IGUserID = p.IGUserID
		}

		if p.PageID != "" {
			cfg.PageID = p.PageID
		}

		if p.BusinessID != "" {
			cfg.BusinessID = p.BusinessID
		}
	}

	token, ok, err := secrets.GetAccessToken(name)
	if err != nil {
		return nil, fmt.Errorf("load access token: %w", err)
	}

	if ok {
		cfg.AccessToken = token
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return errConfigNil
	}

	missing := missingEnvs([]requiredEnv{
		{name: envAppID, value: c.AppID},
		{name: envAppSecret, value: c.AppSecret},
		{name: envAccessToken, value: c.AccessToken},
		{name: envPageID, value: c.PageID},
		{name: envIGUserID, value: c.IGUserID},
	})
	if len(missing) > 0 {
		return &MissingEnvError{Missing: missing}
	}

	return nil
}

func (c *Config) ValidateForAccessToken() error {
	if c == nil {
		return errConfigNil
	}

	missing := missingEnvs([]requiredEnv{
		{name: envAccessToken, value: c.AccessToken},
		{name: envIGUserID, value: c.IGUserID},
	})
	if len(missing) > 0 {
		return &MissingEnvError{Missing: missing}
	}

	return nil
}

func (c *Config) ValidateForAccountLookup() error {
	if c == nil {
		return errConfigNil
	}

	missing := missingEnvs([]requiredEnv{
		{name: envPageID, value: c.PageID},
		{name: envAccessToken, value: c.AccessToken},
	})
	if len(missing) > 0 {
		return &MissingEnvError{Missing: missing}
	}

	return nil
}

func (c *Config) ValidateForBusinessLookup() error {
	if c == nil {
		return errConfigNil
	}

	missing := missingEnvs([]requiredEnv{
		{name: envBusinessID, value: c.BusinessID},
		{name: envAccessToken, value: c.AccessToken},
	})
	if len(missing) > 0 {
		return &MissingEnvError{Missing: missing}
	}

	return nil
}

func (c *Config) ValidateForTokenExchange() error {
	if c == nil {
		return errConfigNil
	}

	missing := missingEnvs([]requiredEnv{
		{name: envAppID, value: c.AppID},
		{name: envAppSecret, value: c.AppSecret},
	})
	if len(missing) > 0 {
		return &MissingEnvError{Missing: missing}
	}

	return nil
}

func (c *Config) ValidateForTokenDebug() error {
	if c == nil {
		return errConfigNil
	}

	if c.AccessToken == "" {
		return &MissingEnvError{Missing: []string{envAccessToken}}
	}

	return nil
}

type requiredEnv struct {
	name  string
	value string
}

func missingEnvs(entries []requiredEnv) []string {
	missing := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.value == "" {
			missing = append(missing, entry.name)
		}
	}

	return missing
}
