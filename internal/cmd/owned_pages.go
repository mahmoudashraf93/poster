package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mahmoudashraf93/poster/internal/config"
	"github.com/mahmoudashraf93/poster/internal/graph"
)

type OwnedPagesCmd struct{}

func (c *OwnedPagesCmd) Run(root *RootFlags) error {
	cfg, err := config.LoadWithProfile(root.Profile)
	if err != nil {
		return err
	}

	if root != nil && root.BusinessID != "" {
		cfg.BusinessID = root.BusinessID
	}

	err = cfg.ValidateForBusinessLookup()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pages, err := graph.FetchOwnedPages(ctx, cfg.BusinessID, cfg.AccessToken)
	if err != nil {
		return err
	}

	for _, page := range pages {
		_, _ = fmt.Fprintf(os.Stdout, "PAGE_ID=%s\n", page.ID)
		_, _ = fmt.Fprintf(os.Stdout, "PAGE_NAME=%s\n", page.Name)
		if page.IGUserID != "" {
			_, _ = fmt.Fprintf(os.Stdout, "IG_USER_ID=%s\n", page.IGUserID)
		} else {
			_, _ = fmt.Fprintln(os.Stdout, "IG_USER_ID=")
		}
		_, _ = fmt.Fprintln(os.Stdout, "---")
	}

	if len(pages) == 0 {
		_, _ = fmt.Fprintln(os.Stdout, "NO_PAGES_FOUND")
	}

	return nil
}
