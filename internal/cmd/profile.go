package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/mahmoud/igpostercli/internal/config"
	"github.com/mahmoud/igpostercli/internal/secrets"
)

type ProfileCmd struct {
	Set    ProfileSetCmd    `cmd:"" help:"Create or update a profile"`
	Show   ProfileShowCmd   `cmd:"" help:"Show profile details"`
	List   ProfileListCmd   `cmd:"" help:"List profiles"`
	Delete ProfileDeleteCmd `cmd:"" help:"Delete a profile"`
}

type ProfileSetCmd struct {
	Name        string  `arg:"" optional:"" help:"Profile name (defaults to current)"`
	AccessToken *string `help:"Access token (stored in keychain)"`
	UserID      *string `help:"Instagram user ID"`
	PageID      *string `help:"Facebook Page ID"`
	BusinessID  *string `help:"Business ID"`
}

func (c *ProfileSetCmd) Run(root *RootFlags) error {
	name := c.Name
	if name == "" && root != nil {
		name = root.Profile
	}

	name, err := config.NormalizeProfileNameOrDefault(name)
	if err != nil {
		return err
	}

	cfg, err := config.ReadProfiles()
	if err != nil {
		return err
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]config.Profile)
	}

	profile := cfg.Profiles[name]

	if c.UserID != nil {
		if *c.UserID == "" {
			return usage("--user-id cannot be empty")
		}
		profile.IGUserID = *c.UserID
	}

	if c.PageID != nil {
		if *c.PageID == "" {
			return usage("--page-id cannot be empty")
		}
		profile.PageID = *c.PageID
	}

	if c.BusinessID != nil {
		if *c.BusinessID == "" {
			return usage("--business-id cannot be empty")
		}
		profile.BusinessID = *c.BusinessID
	}

	cfg.Profiles[name] = profile

	if err := config.WriteProfiles(cfg); err != nil {
		return err
	}

	if c.AccessToken != nil {
		if *c.AccessToken == "" {
			return usage("--access-token cannot be empty")
		}
		if err := secrets.SetAccessToken(name, *c.AccessToken); err != nil {
			return err
		}
	}

	return printProfile(name, profile)
}

type ProfileShowCmd struct {
	Name string `arg:"" optional:"" help:"Profile name (defaults to current)"`
}

func (c *ProfileShowCmd) Run(root *RootFlags) error {
	name := c.Name
	if name == "" && root != nil {
		name = root.Profile
	}

	resolved, err := config.NormalizeProfileNameOrDefault(name)
	if err != nil {
		return err
	}

	cfg, err := config.ReadProfiles()
	if err != nil {
		return err
	}

	profile, ok := cfg.Profiles[resolved]
	if !ok {
		return fmt.Errorf("profile not found: %s", resolved)
	}

	return printProfile(resolved, profile)
}

type ProfileListCmd struct{}

func (c *ProfileListCmd) Run() error {
	cfg, err := config.ReadProfiles()
	if err != nil {
		return err
	}

	profiles := config.ListProfiles(cfg)
	if len(profiles) == 0 {
		_, _ = fmt.Fprintln(os.Stdout, "NO_PROFILES_FOUND")
		return nil
	}

	for _, name := range profiles {
		profile := cfg.Profiles[name]
		if err := printProfile(name, profile); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, "---")
	}

	return nil
}

type ProfileDeleteCmd struct {
	Name string `arg:"" help:"Profile name"`
}

func (c *ProfileDeleteCmd) Run() error {
	name, err := config.NormalizeProfileNameOrDefault(c.Name)
	if err != nil {
		return err
	}

	cfg, err := config.ReadProfiles()
	if err != nil {
		return err
	}

	_, ok := cfg.Profiles[name]
	if !ok {
		return fmt.Errorf("profile not found: %s", name)
	}

	delete(cfg.Profiles, name)

	err = config.WriteProfiles(cfg)
	if err != nil {
		return err
	}

	_, err = secrets.DeleteAccessToken(name)
	if err != nil && !errors.Is(err, secrets.ErrTokenNotFound()) {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "DELETED_PROFILE=%s\n", name)
	return nil
}

func printProfile(name string, profile config.Profile) error {
	_, _ = fmt.Fprintf(os.Stdout, "PROFILE=%s\n", name)
	_, _ = fmt.Fprintf(os.Stdout, "IG_USER_ID=%s\n", profile.IGUserID)
	_, _ = fmt.Fprintf(os.Stdout, "PAGE_ID=%s\n", profile.PageID)
	_, _ = fmt.Fprintf(os.Stdout, "BUSINESS_ID=%s\n", profile.BusinessID)

	_, ok, err := secrets.GetAccessToken(name)
	if err != nil {
		return err
	}

	if ok {
		_, _ = fmt.Fprintln(os.Stdout, "ACCESS_TOKEN=SET")
	} else {
		_, _ = fmt.Fprintln(os.Stdout, "ACCESS_TOKEN=EMPTY")
	}

	return nil
}
