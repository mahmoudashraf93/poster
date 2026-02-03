package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mahmoud/igpostercli/internal/config"
	"github.com/mahmoud/igpostercli/internal/graph"
)

type AccountCmd struct{}

func (c *AccountCmd) Run(root *RootFlags) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if root != nil && root.PageID != "" {
		cfg.PageID = root.PageID
	}

	err = cfg.ValidateForAccountLookup()
	if err != nil {
		return err
	}

	ctx := context.Background()
	igUserID, err := graph.FetchIGUserID(ctx, cfg.PageID, cfg.AccessToken)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "IG_USER_ID=%s\n", igUserID)
	return nil
}
