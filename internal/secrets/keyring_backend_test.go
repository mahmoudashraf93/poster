package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveKeyringBackendInfoEnv(t *testing.T) {
	t.Setenv("POSTER_KEYRING_BACKEND", "FILE")

	info, err := ResolveKeyringBackendInfo()
	if err != nil {
		t.Fatalf("resolve keyring backend: %v", err)
	}

	if info.Value != "file" {
		t.Fatalf("expected backend file, got %q", info.Value)
	}

	if info.Source != "env" {
		t.Fatalf("expected source env, got %q", info.Source)
	}
}

func TestResolveKeyringBackendInfoConfig(t *testing.T) {
	base := t.TempDir()
	t.Setenv("POSTER_KEYRING_BACKEND", "")
	t.Setenv("XDG_CONFIG_HOME", base)
	t.Setenv("HOME", base)

	path, err := configPath()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	if err := os.WriteFile(path, []byte(`{"keyring_backend":"file"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	info, err := ResolveKeyringBackendInfo()
	if err != nil {
		t.Fatalf("resolve keyring backend: %v", err)
	}

	if info.Value != "file" {
		t.Fatalf("expected backend file, got %q", info.Value)
	}

	if info.Source != "config" {
		t.Fatalf("expected source config, got %q", info.Source)
	}
}

func TestResolveKeyringBackendInfoDefault(t *testing.T) {
	base := t.TempDir()
	t.Setenv("POSTER_KEYRING_BACKEND", "")
	t.Setenv("XDG_CONFIG_HOME", base)
	t.Setenv("HOME", base)

	info, err := ResolveKeyringBackendInfo()
	if err != nil {
		t.Fatalf("resolve keyring backend: %v", err)
	}

	if info.Value != "auto" {
		t.Fatalf("expected backend auto, got %q", info.Value)
	}

	if info.Source != "default" {
		t.Fatalf("expected source default, got %q", info.Source)
	}
}
