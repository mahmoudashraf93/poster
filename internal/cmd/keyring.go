package cmd

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/mahmoudashraf93/poster/internal/config"
	"github.com/mahmoudashraf93/poster/internal/secrets"
)

type KeyringCmd struct {
	Backend  string `arg:"" optional:"" name:"backend" help:"Keyring backend: auto|keychain|file"`
	Backend2 string `arg:"" optional:"" name:"backend2" help:"(compat) Use: poster keyring set <backend>"`
}

func (c *KeyringCmd) Run() error {
	backend := strings.ToLower(strings.TrimSpace(c.Backend))
	backend2 := strings.ToLower(strings.TrimSpace(c.Backend2))

	// Backwards compat for earlier suggestion: `poster keyring set <backend>`.
	if backend == "set" {
		backend = backend2
		backend2 = ""
	}

	// No args: show current config.
	if backend == "" {
		return printKeyringInfo()
	}

	if backend2 != "" {
		return usage(fmt.Sprintf("too many args: %q %q", c.Backend, c.Backend2))
	}

	if backend == "default" {
		backend = "auto"
	}

	if _, ok := allowedKeyringBackends()[backend]; !ok {
		return usage(fmt.Sprintf("invalid backend: %q (expected auto, keychain, or file)", c.Backend))
	}

	cfg, err := config.ReadProfiles()
	if err != nil {
		return err
	}
	cfg.KeyringBackend = backend
	if err := config.WriteProfiles(cfg); err != nil {
		return err
	}

	path, _ := config.ProfilesPath()

	// Env var wins; warn so it doesn't look "broken".
	if v := strings.TrimSpace(os.Getenv("POSTER_KEYRING_BACKEND")); v != "" {
		_, _ = fmt.Fprintf(os.Stderr, "NOTE: POSTER_KEYRING_BACKEND=%s overrides config.json\n", v)
	}

	if backend == "file" {
		if v := strings.TrimSpace(os.Getenv("POSTER_KEYRING_PASSWORD")); v != "" {
			_, _ = fmt.Fprintln(os.Stderr, "POSTER_KEYRING_PASSWORD found in environment.")
		} else if !term.IsTerminal(int(os.Stdin.Fd())) {
			_, _ = fmt.Fprintln(os.Stderr, "NOTE: file keyring backend in non-interactive context requires POSTER_KEYRING_PASSWORD")
		} else {
			_, _ = fmt.Fprintln(os.Stderr, "Hint: set POSTER_KEYRING_PASSWORD for non-interactive use (CI/ssh)")
		}
	}

	_, _ = fmt.Fprintln(os.Stdout, "written\ttrue")
	_, _ = fmt.Fprintf(os.Stdout, "path\t%s\n", path)
	_, _ = fmt.Fprintf(os.Stdout, "keyring_backend\t%s\n", backend)
	return nil
}

func printKeyringInfo() error {
	info, err := secrets.ResolveKeyringBackendInfo()
	if err != nil {
		return err
	}

	path, _ := config.ProfilesPath()
	_, _ = fmt.Fprintf(os.Stdout, "path\t%s\n", path)
	_, _ = fmt.Fprintf(os.Stdout, "keyring_backend\t%s\n", info.Value)
	_, _ = fmt.Fprintf(os.Stdout, "source\t%s\n", info.Source)
	_, _ = fmt.Fprintln(os.Stderr, "Hint: poster keyring <auto|keychain|file>")
	return nil
}

func allowedKeyringBackends() map[string]struct{} {
	return map[string]struct{}{
		"auto":     {},
		"keychain": {},
		"file":     {},
	}
}
