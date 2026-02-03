package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"golang.org/x/term"
)

const (
	serviceName = "poster"

	keyringPasswordEnv = "POSTER_KEYRING_PASSWORD" //nolint:gosec // env var name, not a credential
	keyringBackendEnv  = "POSTER_KEYRING_BACKEND"  //nolint:gosec // env var name, not a credential
)

var (
	errTokenNotFound         = errors.New("access token not found")
	errNoTTY                 = errors.New("no TTY available for keyring file backend password prompt")
	errInvalidKeyringBackend = errors.New("invalid keyring backend")
	errKeyringTimeout        = errors.New("keyring connection timed out")
	openKeyring              = openKeyringDefault
	keyringOpenFunc          = keyring.Open
)

type KeyringBackendInfo struct {
	Value  string
	Source string
}

const (
	keyringBackendSourceEnv     = "env"
	keyringBackendSourceConfig  = "config"
	keyringBackendSourceDefault = "default"
	keyringBackendAuto          = "auto"
)

func OpenKeyring() (keyring.Keyring, error) {
	return openKeyring()
}

type keyringConfig struct {
	KeyringBackend string `json:"keyring_backend"`
}

func configPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(base, serviceName, "config.json"), nil
}

func readKeyringConfig() (keyringConfig, error) {
	path, err := configPath()
	if err != nil {
		return keyringConfig{}, err
	}

	b, err := os.ReadFile(path) //nolint:gosec // config file path
	if err != nil {
		if os.IsNotExist(err) {
			return keyringConfig{}, nil
		}
		return keyringConfig{}, fmt.Errorf("read config: %w", err)
	}

	var cfg keyringConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return keyringConfig{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	return cfg, nil
}

func ResolveKeyringBackendInfo() (KeyringBackendInfo, error) {
	if v := normalizeKeyringBackend(os.Getenv(keyringBackendEnv)); v != "" {
		return KeyringBackendInfo{Value: v, Source: keyringBackendSourceEnv}, nil
	}

	cfg, err := readKeyringConfig()
	if err != nil {
		return KeyringBackendInfo{}, fmt.Errorf("resolve keyring backend: %w", err)
	}

	if cfg.KeyringBackend != "" {
		if v := normalizeKeyringBackend(cfg.KeyringBackend); v != "" {
			return KeyringBackendInfo{Value: v, Source: keyringBackendSourceConfig}, nil
		}
	}

	return KeyringBackendInfo{Value: keyringBackendAuto, Source: keyringBackendSourceDefault}, nil
}

func allowedBackends(info KeyringBackendInfo) ([]keyring.BackendType, error) {
	switch info.Value {
	case "", keyringBackendAuto:
		return nil, nil
	case "keychain":
		return []keyring.BackendType{keyring.KeychainBackend}, nil
	case "file":
		return []keyring.BackendType{keyring.FileBackend}, nil
	default:
		return nil, fmt.Errorf("%w: %q (expected %s, keychain, or file)", errInvalidKeyringBackend, info.Value, keyringBackendAuto)
	}
}

func fileKeyringPasswordFuncFrom(password string, isTTY bool) keyring.PromptFunc {
	if password != "" {
		return keyring.FixedStringPrompt(password)
	}

	if isTTY {
		return keyring.TerminalPrompt
	}

	return func(_ string) (string, error) {
		return "", fmt.Errorf("%w; set %s", errNoTTY, keyringPasswordEnv)
	}
}

func fileKeyringPasswordFunc() keyring.PromptFunc {
	return fileKeyringPasswordFuncFrom(os.Getenv(keyringPasswordEnv), term.IsTerminal(int(os.Stdin.Fd())))
}

func normalizeKeyringBackend(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// keyringOpenTimeout is the maximum time to wait for keyring.Open() to complete.
// On headless Linux, D-Bus SecretService can hang indefinitely if gnome-keyring
// is installed but not running.
const keyringOpenTimeout = 5 * time.Second

func shouldForceFileBackend(goos string, backendInfo KeyringBackendInfo, dbusAddr string) bool {
	return goos == "linux" && backendInfo.Value == keyringBackendAuto && dbusAddr == ""
}

func shouldUseKeyringTimeout(goos string, backendInfo KeyringBackendInfo, dbusAddr string) bool {
	return goos == "linux" && backendInfo.Value == keyringBackendAuto && dbusAddr != ""
}

func ensureKeyringDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	configDir := filepath.Join(base, serviceName)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return "", fmt.Errorf("ensure config dir: %w", err)
	}

	keyringDir := filepath.Join(configDir, "keyring")
	if err := os.MkdirAll(keyringDir, 0o700); err != nil {
		return "", fmt.Errorf("ensure keyring dir: %w", err)
	}

	return keyringDir, nil
}

func openKeyringDefault() (keyring.Keyring, error) {
	// On Linux/containers, OS keychains (secret-service/kwallet) may be unavailable.
	// In that case github.com/99designs/keyring falls back to the "file" backend,
	// which requires both a directory and a password prompt function.
	keyringDir, err := ensureKeyringDir()
	if err != nil {
		return nil, err
	}

	backendInfo, err := ResolveKeyringBackendInfo()
	if err != nil {
		return nil, err
	}

	backends, err := allowedBackends(backendInfo)
	if err != nil {
		return nil, err
	}

	dbusAddr := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	// On Linux with "auto" backend and no D-Bus session, force file backend.
	// Without DBUS_SESSION_BUS_ADDRESS, SecretService will hang indefinitely
	// trying to connect (common on headless systems).
	if shouldForceFileBackend(runtime.GOOS, backendInfo, dbusAddr) {
		backends = []keyring.BackendType{keyring.FileBackend}
	}

	cfg := keyring.Config{
		ServiceName:      serviceName,
		AllowedBackends:  backends,
		FileDir:          keyringDir,
		FilePasswordFunc: fileKeyringPasswordFunc(),
	}

	// On Linux with D-Bus present, keyring.Open() can still hang if SecretService
	// is unresponsive (e.g., gnome-keyring installed but not running).
	if shouldUseKeyringTimeout(runtime.GOOS, backendInfo, dbusAddr) {
		return openKeyringWithTimeout(cfg, keyringOpenTimeout)
	}

	kr, err := keyringOpenFunc(cfg)
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	return kr, nil
}

type keyringResult struct {
	ring keyring.Keyring
	err  error
}

// openKeyringWithTimeout wraps keyring.Open with a timeout to prevent indefinite
// hangs when D-Bus SecretService is unresponsive (e.g., gnome-keyring installed
// but not running on headless Linux).
//
// Note: If timeout occurs, the spawned goroutine continues blocking on keyring.Open()
// and will leak. This is acceptable for a CLI tool since the process exits on this
// error, but would need refactoring for long-running use.
func openKeyringWithTimeout(cfg keyring.Config, timeout time.Duration) (keyring.Keyring, error) {
	ch := make(chan keyringResult, 1)

	go func() {
		ring, err := keyringOpenFunc(cfg)
		ch <- keyringResult{ring, err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("open keyring: %w", res.err)
		}

		return res.ring, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("%w after %v (D-Bus SecretService may be unresponsive); "+
			"set %s=file and %s=<password> to use encrypted file storage instead",
			errKeyringTimeout, timeout, keyringBackendEnv, keyringPasswordEnv)
	}
}

func tokenKey(profile string) string {
	return fmt.Sprintf("profile:%s", profile)
}

func SetAccessToken(profile string, token string) error {
	kr, err := OpenKeyring()
	if err != nil {
		return err
	}

	item := keyring.Item{
		Key:   tokenKey(profile),
		Data:  []byte(token),
		Label: "poster access token",
	}

	if err := kr.Set(item); err != nil {
		return fmt.Errorf("store access token: %w", err)
	}

	return nil
}

func GetAccessToken(profile string) (string, bool, error) {
	kr, err := OpenKeyring()
	if err != nil {
		return "", false, err
	}

	item, err := kr.Get(tokenKey(profile))
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", false, nil
		}

		return "", false, fmt.Errorf("get access token: %w", err)
	}

	return string(item.Data), true, nil
}

func DeleteAccessToken(profile string) (bool, error) {
	kr, err := OpenKeyring()
	if err != nil {
		return false, err
	}

	if err := kr.Remove(tokenKey(profile)); err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) || os.IsNotExist(err) {
			return false, errTokenNotFound
		}

		return false, fmt.Errorf("delete access token: %w", err)
	}

	return true, nil
}

func ErrTokenNotFound() error {
	return errTokenNotFound
}

// SetOpenKeyringForTests overrides keyring opener in tests.
func SetOpenKeyringForTests(fn func() (keyring.Keyring, error)) func() {
	prev := openKeyring
	openKeyring = fn

	return func() {
		openKeyring = prev
	}
}
